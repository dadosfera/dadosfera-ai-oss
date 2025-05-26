import datetime
import json
import os

from _orchest.internals import config as _config

assert (
    _config.CONTAINER_RUNTIME is not None
), "CONTAINER_RUNTIME should be set for the orchest-api and celery-worker."


def _get_worker_plane_selector():
    selector_string = os.environ.get("WORKER_PLANE_SELECTOR")
    if selector_string is None:
        return

    try:
        selector_dict = json.loads(selector_string)
        if not isinstance(selector_dict, dict):
            return
        return selector_dict
    except json.decoder.JSONDecodeError:
        return


class Config:
    # TODO: Should we read these from ENV variables instead?
    DEBUG = False
    TESTING = False
    DEV_MODE = os.environ.get("FLASK_ENV") == "development"

    # Whether Orchest is running on single-node. This in turn determines
    # how pipeline runs are executed; in 1 pod or 1 pod per step.
    SINGLE_NODE = os.environ.get("SINGLE_NODE") == "TRUE"
    ORCHEST_VERSION = os.environ["ORCHEST_VERSION"]
    # must be uppercase
    SQLALCHEMY_DATABASE_URI = "postgresql://postgres@orchest-database/orchest_api"

    WORKER_PLANE_SELECTOR = _get_worker_plane_selector()

    SQLALCHEMY_TRACK_MODIFICATIONS = False

    # TODO: for now this is put here.
    ORCHEST_API_ADDRESS = "http://orchest-api:80/api"
    ORCHEST_WEBSERVER_ADDRESS = "http://orchest-webserver:80"
    REGISTRY_ADDRESS = f"https://{_config.REGISTRY_FQDN}"
    # This is mounted to both the celery worker and orchest-api.
    REGISTRY_TLS_CERT_BUNDLE = "/usr/lib/ssl/certs/additional-ca-cert-bundle.crt"
    KNOWN_HOSTS_CONFIGMAP = "orchest-known-hosts"

    # How often different internal scheduler tasks run.
    CLEANUP_OLD_SCHEDULER_JOB_RECORDS_INTERVAL = 5 * 60
    IMAGES_DELETION_INTERVAL = 2 * 60
    NOTIFICATIONS_DELIVERIES_INTERVAL = 1
    SCHEDULER_INTERVAL = 10

    GPU_ENABLED_INSTANCE = _config.GPU_ENABLED_INSTANCE

    # Used to decide when client heartbeats are too old to represent
    # activity.
    CLIENT_HEARTBEATS_IDLENESS_THRESHOLD = datetime.timedelta(minutes=30)

    # Image building.
    _RUNTIME_TO_IMAGE_BUILDER = {
        "docker": os.environ.get(
            "DOCKER_IMAGE_BUILDER_IMAGE",
            f"docker.io/dadosfera/image-builder-buildx:{ORCHEST_VERSION}",
        ),
        "containerd": os.environ.get(
            "CONTAINERD_IMAGE_BUILDER_IMAGE",
            f"docker.io/dadosfera/image-builder-buildkit:{ORCHEST_VERSION}",
        ),
        "cri-o": os.environ.get(
            "CRIO_IMAGE_BUILDER_IMAGE",
            f"docker.io/dadosfera/image-builder-buildkit:{ORCHEST_VERSION}",
        ),
    }
    IMAGE_BUILDER_IMAGE = _RUNTIME_TO_IMAGE_BUILDER[_config.CONTAINER_RUNTIME]

    BUILD_IMAGE_LOG_FLAG = "_ORCHEST_RESERVED_LOG_FLAG_"
    BUILD_IMAGE_ERROR_FLAG = "_ORCHEST_RESERVED_ERROR_FLAG_"
    JUPYTER_SERVER_IMAGE = os.environ.get(
        "JUPYTER_SERVER_IMAGE", "docker.io/dadosfera/jupyter-server"
    )
    DOCKERHUB_SECRET_NAME = os.environ.get("DOCKERHUB_SECRET_NAME")
    # ---- Celery configurations ----
    # NOTE: the configurations have to be lowercase.
    # NOTE: Flask will not configure lowercase variables. Therefore the
    # config class will be loaded directly by the Celery instance.
    broker_url = "amqp://guest:guest@rabbitmq-server:5672//"

    # NOTE: the database might require trimming from time to time, to
    # enable having the db trimmed automatically use:
    # https://docs.celeryproject.org/en/master/userguide/configuration.html#std:setting-result_expires
    # note that "When using the database backend, celery beat must be
    # running for the results to be expired." The db reaching a large
    # size is probably an unreasonable edge case, given that currently
    # our celery tasks do not produce output.  For this reason no
    # solution is provided for this, taking into account that we will
    # eventually implement cronjobs and such trimming might be an
    # internal cronjob, or automatically managed by celery if we end
    # using "celery beat".
    _result_backend_server = "postgres@orchest-database/celery_result_backend"

    # used to create the db if it does not exist, the function needs
    # this exact url format
    result_backend_sqlalchemy_uri = f"postgresql://{_result_backend_server}"
    # this format is used by celery
    result_backend = f"db+postgresql+psycopg2://{_result_backend_server}"
    worker_prefetch_multiplier = 1

    imports = ("app.core.tasks",)
    task_create_missing_queues = True
    task_default_queue = "celery"
    task_routes = {
        "app.core.tasks.start_non_interactive_pipeline_run": {"queue": "jobs"},
        "app.core.tasks.delete_job_pipeline_run_directories": {"queue": "jobs"},
        "app.core.tasks.run_pipeline": {"queue": "celery"},
        "app.core.tasks.build_environment_image": {"queue": "builds"},
        "app.core.tasks.build_jupyter_image": {"queue": "builds"},
        "app.core.tasks.registry_garbage_collection": {"queue": "builds"},
        "app.core.tasks.process_notifications_deliveries": {"queue": "deliveries"},
        "app.core.tasks.git_import": {"queue": "other-tasks"},
    }

    # Priority class name for k8s scheduling, None means don't use priority class
    # This is set by the controller and especially useful for clusters that have
    # many pods and run on single node. This is useful to ensure that the critical
    # pipelines and the core of the orchest can be scheduled properly.
    PRIORITY_CLASS_NAME = os.environ.get("PRIORITY_CLASS_NAME", None)


# ---- CONFIGURATIONS ----
# Production
CONFIG_CLASS = Config
