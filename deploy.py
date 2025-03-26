import os
import subprocess
import yaml

def update_kubeconfig(cluster_name: str, aws_region: str) -> None:
    print(f"Updating kubectl to using EKS cluster {cluster_name}")
    subprocess.check_call(
        f"aws eks update-kubeconfig --region {aws_region} --name {cluster_name}",
        shell=True,
    )

def get_orchest_api_pod():
    _process = subprocess.Popen(
        [
            "kubectl",
            "get",
            "pods",
            "-n",
            "orchest",
            "-l",
            "controller.orchest.io/component=orchest-api",
            "--field-selector=status.phase=Running",
            "--no-headers",
            "--output=jsonpath={.items..metadata.name}",
        ],
        stdout=subprocess.PIPE,
    )

    pod_name = _process.communicate()
    if isinstance(pod_name, tuple):
        return pod_name[0].decode('utf-8')
    raise Exception("Unknown Error")

def get_configs(environment):
    with open(f'backups/{environment}.config.yaml') as f:
        configs = yaml.safe_load(f.read())
    return configs

def deploy_project_locally(project, pod_name):
    rm_cmd = f"minikube kubectl -- exec {pod_name} -it -n orchest -- rm -rf /userdir/projects/{project}"
    print(f'running command: {rm_cmd}')
    subprocess.check_call(rm_cmd, shell=True)

    cp_cmd = f'minikube kubectl -- cp backups/projects/{project} orchest/{pod_name}:/userdir/projects'
    print(f'running command: {cp_cmd}')
    subprocess.check_call(cp_cmd, shell=True)


def deploy_project(project, pod_name):
    cp_cmd = f'kubectl cp backups/projects/{project} orchest/{pod_name}:/userdir/projects'
    print(f'running command: {cp_cmd}')
    subprocess.check_call(cp_cmd, shell=True)

def deploy(configs):
    for config in configs:
        update_kubeconfig(config['cluster_name'], 'us-east-1')
        pod_name = get_orchest_api_pod()
        for project in config['projects']:
            print(f"Deploying project: {project}")
            deploy_project(project, pod_name)

def deploy_locally(configs):
    for config in configs:
        pod_name = get_orchest_api_pod()
        for project in config['projects']:
            print(f"Deploying project: {project}")
            deploy_project_locally(project, pod_name)



if __name__ == "__main__":
    environment = os.environ["ENVIRONMENT"]
    import os
    print(os.getcwd())
    if environment == "local":
        configs = get_configs(environment)
        deploy_locally(configs)
    else:
        configs = get_configs(environment)
        deploy(configs)
