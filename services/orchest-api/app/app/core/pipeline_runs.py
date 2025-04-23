"""Module about pipeline runs.

Essentially, it covers:
- transforming a Pipeline instance to a valid k8s workflow
    definition, where the pipeline is run as an argo workflow.
- the required function to actually perform a pipeline run.

As a client of this module you are most likely interested in how to
perform a pipeline run given a Pipeline instance and some configuration
parameters.

"""
import json
import os
import time
from typing import Any, Dict, List, Set
from kubernetes import client
from celery.contrib.abortable import AbortableAsyncResult

import app.utils as utils
from _orchest.internals import config as _config
from _orchest.internals import utils as _utils
from _orchest.internals.two_phase_executor import TwoPhaseExecutor
from _orchest.internals.utils import get_step_and_kernel_volumes_and_volume_mounts
from app import models
from app import utils
from app.apis.namespace_jobs import UpdateJobPipelineRun
from app.apis.namespace_runs import UpdateInteractivePipelineRun
from app.connections import db, k8s_core_api, k8s_custom_obj_api
from app.core import pod_scheduling
from app.core.pipelines import Pipeline, PipelineStep
from app.types import RunConfig
from config import CONFIG_CLASS

logger = utils.get_logger()


def _step_to_workflow_manifest_task(
    pipeline: Pipeline, step: PipelineStep, run_config: RunConfig
) -> dict:
    # The working directory is the location of the file being executed.
    project_relative_file_path = os.path.join(
        os.path.split(run_config["pipeline_path"])[0], step.properties["file_path"]
    )
    working_dir = os.path.split(project_relative_file_path)[0]

    user_env_variables = [
        {"name": key, "value": str(value)}
        for key, value in run_config["user_env_variables"].items()
    ]
    orchest_env_variables = [
        {"name": "ORCHEST_STEP_UUID", "value": step.properties["uuid"]},
        {"name": "ORCHEST_SESSION_UUID", "value": run_config["session_uuid"]},
        {"name": "ORCHEST_SESSION_TYPE", "value": run_config["session_type"]},
        {"name": "ORCHEST_PIPELINE_UUID", "value": run_config["pipeline_uuid"]},
        {"name": "ORCHEST_PIPELINE_PATH", "value": _config.PIPELINE_FILE},
        {"name": "ORCHEST_PROJECT_UUID", "value": run_config["project_uuid"]},
        {"name": "ORCHEST_NAMESPACE", "value": _config.ORCHEST_NAMESPACE},
        {"name": "ORCHEST_CLUSTER", "value": _config.ORCHEST_CLUSTER},
    ]
    # Note that the order of concatenation matters, so that there is no
    # risk that the user overwrites internal variables accidentally.
    env_variables = (
        user_env_variables + orchest_env_variables + _utils.get_aws_env_vars()
    )

    # Need to reference the ip because the local docker engine will run
    # the container, and if the image is missing it will prompt a pull
    # which will fail because the FQDN can't be resolved by the local
    # engine on the node. K8S_TODO: fix this.
    registry_ip = k8s_core_api.read_namespaced_service(
        _config.REGISTRY, _config.ORCHEST_NAMESPACE
    ).spec.cluster_ip
    # The image of the step is the registry address plus the image name.
    image = (
        registry_ip
        + "/"
        + run_config["env_uuid_to_image"][step.properties["environment"]]
    )

    if _run_as_container_set(pipeline):
        # NOTE: In this case we don't run an initContainer to pre-pull
        # images.
        task = {
            # "Name cannot begin with a digit when using either
            # 'depends' or 'dependencies'".
            "name": f'step-{step.properties["uuid"]}',
            "dependencies": [
                f'step-{pstep.properties["uuid"]}' for pstep in step.parents
            ],
            "restartPolicy": "Never",
            # NOTE: Should never need to pull given that the pipeline
            # run is scheduled on the only node in the cluster (which
            # thus has the image).
            "imagePullPolicy": "IfNotPresent",
            "env": env_variables,
            "image": image,
            "command": [
                "/orchest/bootscript.sh",
                "runnable",
                working_dir,
                project_relative_file_path,
            ],
            "resources": {
                "requests": {
                    "cpu": _config.USER_CONTAINERS_CPU_SHARES,
                    "memory": _config.USER_CONTAINERS_MEMORY_SHARES,
                }
            },
        }
    else:
        # This allows us to edit the container that argo runs for us.
        pod_spec_dict = {
            "terminationGracePeriodSeconds": 1,
            "containers": [
                {
                    "name": "main",
                    "env": env_variables,
                    "restartPolicy": "Never",
                }
            ],
        }
        if utils.should_use_priority_class():
            pod_spec_dict["priorityClassName"] = CONFIG_CLASS.PRIORITY_CLASS_NAME
        pod_spec_patch = json.dumps(pod_spec_dict)

        task = {
            # "Name cannot begin with a digit when using either
            # 'depends' or 'dependencies'".
            "name": f'step-{step.properties["uuid"]}',
            "dependencies": [
                f'step-{pstep.properties["uuid"]}' for pstep in step.parents
            ],
            "template": "step",
            "arguments": {
                "parameters": [
                    {
                        # Used to keep track of the step when getting
                        # workflow status, since the name we have set is
                        # not reliable, argo will change it.
                        "name": "step_uuid",
                        "value": step.properties["uuid"],
                    },
                    {
                        "name": "image",
                        "value": image,
                    },
                    {"name": "working_dir", "value": working_dir},
                    {
                        "name": "project_relative_file_path",
                        "value": project_relative_file_path,
                    },
                    {"name": "pod_spec_patch", "value": pod_spec_patch},
                    {
                        # NOTE: only used by tests.
                        "name": "tests_uuid",
                        "value": step.properties["uuid"],
                    },
                ]
            },
        }

    return task


def _run_as_container_set(pipeline: Pipeline) -> bool:
    # About the parallelism check, see:
    # https://linear.app/orchest/issue/ORC-1121/allow-limiting-step-parallelism-at-the-pipeline-level
    return (
        CONFIG_CLASS.SINGLE_NODE
        and pipeline.properties["settings"]["max_steps_parallelism"] <= 0
    )

def _get_pipeline_argo_templates(
    entrypoint_name: str,
    volume_mounts: List[dict],
    pipeline: Pipeline,
    run_config: RunConfig,
) -> List[Dict[str, Any]]:
    # !Note that modify_pipeline_scheduling_behaviour relies on the
    # structure and the content of the manifest to inject changes,
    # verify that you aren't breaking anything when changing thins here.
    # TODO: add tests for this.
    # https://argoproj.github.io/argo-workflows/fields/#template
    if _run_as_container_set(pipeline):

        pod_spec_patch = {
            "terminationGracePeriodSeconds": 1,
        }
        if utils.should_use_priority_class():
            pod_spec_patch["priorityClassName"] = CONFIG_CLASS.PRIORITY_CLASS_NAME

        templates = [
            {
                "name": entrypoint_name,
                "failFast": True,
                "retryStrategy": {"limit": "0", "backoff": {"maxDuration": "0s"}},
                "podSpecPatch": json.dumps(pod_spec_patch),
                # Security context for the Pod in which the containers
                # of the containerSet run.
                "securityContext": {
                    "runAsUser": 0,
                    "runAsGroup": int(os.environ.get("ORCHEST_HOST_GID", "1")),
                    "fsGroup": int(os.environ.get("ORCHEST_HOST_GID", "1")),
                },
                # NOTE: Argo only allows a "dag" or "steps" template to
                # reference another template, thus we just create
                # containerSpecs here.
                "containerSet": {
                    "retryStrategy": {"retries": 0},
                    "volumeMounts": volume_mounts,
                    "containers": [
                        _step_to_workflow_manifest_task(pipeline, step, run_config)
                        for step in pipeline.steps
                    ],
                },
            },
        ]

    else:
        # The first entry of this list is the definition of the DAG,
        # while the second entry is the step definition.
        templates = [
            {
                "name": entrypoint_name,
                "retryStrategy": {"limit": "0", "backoff": {"maxDuration": "0s"}},
                "dag": {
                    "failFast": True,
                    "tasks": [
                        _step_to_workflow_manifest_task(pipeline, step, run_config)
                        for step in pipeline.steps
                    ],
                },
            },
            {
                # ! The name is important for the logic that alters the
                # scheduling behaviour.
                "name": "step",
                "securityContext": {
                    "runAsUser": 0,
                    "runAsGroup": int(os.environ.get("ORCHEST_HOST_GID", "1")),
                    "fsGroup": int(os.environ.get("ORCHEST_HOST_GID", "1")),
                },
                "inputs": {
                    "parameters": [
                        {"name": param}
                        for param in [
                            "step_uuid",
                            "image",
                            "working_dir",
                            "project_relative_file_path",
                            "pod_spec_patch",
                            "tests_uuid",
                        ]
                    ]
                },
                "retryStrategy": {"limit": "0", "backoff": {"maxDuration": "0s"}},
                "container": {
                    "image": "{{inputs.parameters.image}}",
                    "command": [
                        "/orchest/bootscript.sh",
                        "runnable",
                        "{{inputs.parameters.working_dir}}",
                        "{{inputs.parameters.project_relative_file_path}}",
                    ],
                    "volumeMounts": volume_mounts,
                    "resources": {
                        "requests": {
                            "cpu": _config.USER_CONTAINERS_CPU_SHARES,
                            "memory": _config.USER_CONTAINERS_MEMORY_SHARES,
                        }
                    },
                },
                "podSpecPatch": "{{inputs.parameters.pod_spec_patch}}",
            },
        ]

    return templates


def _pipeline_to_workflow_manifest(
    session_uuid: str,
    workflow_name: str,
    pipeline: Pipeline,
    run_config: RunConfig,
) -> dict:
    volumes, volume_mounts = get_step_and_kernel_volumes_and_volume_mounts(
        userdir_pvc=run_config["userdir_pvc"],
        project_dir=run_config["project_dir"],
        pipeline_file=run_config["pipeline_path"],
        container_project_dir=_config.PROJECT_DIR,
        container_pipeline_file=_config.PIPELINE_FILE,
        container_runtime_socket=_config.CONTAINER_RUNTIME_SOCKET,
    )

    # these parameters will be fed by _step_to_workflow_manifest_task
    entrypoint_name = "pipeline"
    manifest = {
        "apiVersion": "argoproj.io/v1alpha1",
        "kind": "Workflow",
        "metadata": {
            "name": workflow_name,
            "labels": {
                "project_uuid": run_config["project_uuid"],
                "session_uuid": session_uuid,
            },
        },
        "spec": {
            "entrypoint": entrypoint_name,
            "volumes": volumes,
            # The celery task actually takes care of deleting the
            # workflow, this is just a failsafe.
            "ttlStrategy": {
                "secondsAfterCompletion": 1000,
                "secondsAfterSuccess": 1000,
                "secondsAfterFailure": 1000,
            },
            # NOTE: It would be "better" to set `dnsPolicy` to `None`
            # and specify the `search` directive in the `dnsConfig`
            # as to only search within the namespace that Orchest is
            # installed (and then on the internet). This reduced the
            # number of DNS queries, since we know that these are the
            # only two valid options when doing pipeline runs.
            "dnsPolicy": "ClusterFirst",
            "dnsConfig": {
                "options": [
                    {"name": "timeout", "value": "10"},  # 30 is max
                    {"name": "attempts", "value": "5"},  # 5 is max
                ],
            },
            "restartPolicy": "Never",
            "templates": _get_pipeline_argo_templates(
                entrypoint_name=entrypoint_name,
                volume_mounts=volume_mounts,
                pipeline=pipeline,
                run_config=run_config,
            ),
        },
    }

    if pipeline.properties["settings"]["max_steps_parallelism"] > 0:
        manifest["spec"]["parallelism"] = pipeline.properties["settings"][
            "max_steps_parallelism"
        ]
    pod_scheduling.modify_pipeline_scheduling_behaviour(
        run_config["session_type"], manifest
    )
    return manifest


def _is_step_allowed_to_run(
    step: PipelineStep,
    steps_to_finish: Set[PipelineStep],
) -> bool:
    """Returns whether the given step is allowed to run.

    Args:
        step: The step we are checking the run requirements for.
        steps_to_finish: The steps, part of a pipeline run, that have
            not yet finished (or started) executing.

    Returns:
        Have all incoming steps completed?

    """
    return all(s.properties["uuid"] not in steps_to_finish for s in step.parents)


def _update_pipeline_run_status(
    run_config: Dict[str, Any], run_uuid: str, status: str
) -> None:
    with TwoPhaseExecutor(db.session) as tpe:
        if run_config["session_type"] == "interactive":
            UpdateInteractivePipelineRun(tpe).transaction(run_uuid, status)
        else:
            run = (
                db.session.query(models.NonInteractivePipelineRun.job_uuid)
                .filter_by(uuid=run_uuid)
                .one()
            )
            UpdateJobPipelineRun(tpe).transaction(run.job_uuid, run_uuid, status)


def _pipeline_has_reached_end_state(run_config: Dict[str, Any], run_uuid: str) -> bool:
    if AbortableAsyncResult(run_uuid).is_aborted():
        return True

    if run_config["session_type"] == "interactive":
        model = models.InteractivePipelineRun
    else:
        model = models.NonInteractivePipelineRun

    run = db.session.query(model.status).filter_by(uuid=run_uuid).one_or_none()
    return run is None or run.status in ["SUCCESS", "FAILURE", "ABORTED"]


def run_pipeline_workflow(
    session_uuid: str, task_id: str, pipeline: Pipeline, *, run_config: RunConfig
):
    _update_pipeline_run_status(run_config, task_id, "STARTED")

    namespace = _config.ORCHEST_NAMESPACE

    steps_to_start = {step.properties["uuid"] for step in pipeline.steps}
    steps_to_finish = set(steps_to_start)
    had_failed_steps = False
    run_as_container_set = _run_as_container_set(pipeline)
    try:
        manifest = _pipeline_to_workflow_manifest(
            session_uuid, f"pipeline-run-task-{task_id}", pipeline, run_config
        )
        try:
            k8s_custom_obj_api.create_namespaced_custom_object(
                "argoproj.io", "v1alpha1", namespace, "workflows", body=manifest
            )
        # It's difficult to reproduce but it looks like that, on some
        # cases during a restart, rabbitmq has given the task to the
        # worker again, likely due to the worker losing connection (?).
        # This makes it so that the workflow is not cancelled and failed
        # unnecessarily.
        except client.rest.ApiException as api_exception:
            if not (
                api_exception.status == 409 and "AlreadyExists" in api_exception.body
            ):
                raise api_exception

        while steps_to_finish:
            resp = k8s_custom_obj_api.get_namespaced_custom_object(
                "argoproj.io",
                "v1alpha1",
                namespace,
                "workflows",
                f"pipeline-run-task-{task_id}",
            )
            workflow_nodes: dict = resp.get("status", {}).get("nodes", {})
            for argo_node in workflow_nodes.values():
                if run_as_container_set:
                    if argo_node.get("type", "") != "Container":
                        continue

                    # Argo doesn't allow to work with templates for a
                    # containerSet. Thus we fall back to the name we
                    # gave to steps, which includes its uuid.
                    step_uuid = argo_node.get("displayName")
                    if step_uuid is None:
                        # Should never happen.
                        raise Exception(
                            "Did not find `displayName` in Argo workflow node:"
                            f" {argo_node}."
                        )
                    else:
                        step_uuid = step_uuid.replace("step-", "")

                else:
                    # The nodes includes the entire "pipeline" node.
                    if argo_node["templateName"] != "step":
                        continue
                    # The step was not run because the workflow failed.
                    if "inputs" not in argo_node:
                        continue
                    if argo_node.get("type", "") != "Pod":
                        continue

                    for param in argo_node["inputs"]["parameters"]:
                        if param["name"] == "step_uuid":
                            step_uuid = param["value"]
                            break
                    else:
                        # Should never happen.
                        raise Exception(
                            "Did not find `step_uuid` in parameters of Argo node:"
                            f" {argo_node}."
                        )

                pipeline_step = pipeline.get_step(step_uuid)
                argo_node_status = argo_node["phase"]
                argo_node_message = argo_node.get("message", "")
                step_status_update = None

                # Argo does not fail a step if the container is stuck in
                # a waiting state. Doesn't look like the pull backoff
                # behavior can be tuned.
                if argo_node_status in ["Pending", "Running"] and (
                    "ImagePullBackOff" in argo_node_message
                    or "ErrImagePull" in argo_node_message
                ):
                    step_status_update = "FAILURE"

                elif (
                    argo_node_status == "Running"
                    and step_uuid in steps_to_start
                    # Strictly speaking only needed in single-node
                    # context as otherwise Argo takes care of correctly
                    # putting a Step in "Running".
                    and _is_step_allowed_to_run(pipeline_step, steps_to_finish)
                ):
                    step_status_update = "STARTED"
                    steps_to_start.remove(step_uuid)

                elif (
                    argo_node_status in ["Succeeded", "Failed", "Error"]
                    and step_uuid in steps_to_finish
                ):
                    step_status_update = {
                        "Succeeded": "SUCCESS",
                        "Failed": "FAILURE",
                        "Error": "FAILURE",
                    }[argo_node_status]

                if step_status_update is not None:
                    if step_status_update == "FAILURE":
                        had_failed_steps = True

                    if step_status_update in ["FAILURE", "ABORTED", "SUCCESS"]:
                        steps_to_finish.remove(step_uuid)
                        if step_uuid in steps_to_start:
                            steps_to_start.remove(step_uuid)

                    utils.update_steps_status(task_id, [step_uuid], step_status_update)
                    db.session.commit()

            if not steps_to_finish or had_failed_steps:
                break

            if _pipeline_has_reached_end_state(run_config, task_id):
                logger.info(f"Run {task_id} was aborted or deleted, exiting task.")
                break

            time.sleep(0.25)

        if steps_to_finish:
            utils.update_steps_status(task_id, steps_to_finish, "ABORTED")
            db.session.commit()

        pipeline_status = "SUCCESS" if not had_failed_steps else "FAILURE"
        _update_pipeline_run_status(run_config, task_id, pipeline_status)

    except Exception as e:
        logger.error(e)
        utils.update_steps_status(task_id, steps_to_finish, "ABORTED")
        db.session.commit()
        _update_pipeline_run_status(run_config, task_id, "FAILURE")
