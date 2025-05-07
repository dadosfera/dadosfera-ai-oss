We have an software that has the following components:
orchest-controller (responsible for creating components in K8s, it creates the core components like orchest-api, orchest-webserver, rabbitmq, postgresql, celery-worker, buildkit-daemon, node-agent) -> 

orchest-api is the main API -> use celery-worker for background tasks

In orchest, we have an feature called Environments where the use can place custom code and generate an docker image. This images are generated in a pod with the buildkitd and buildctl.

The issue is that this worked with containerd and docker, but not with cri-o. Which is currently used by oracle.

We have faced some issues:
1. orchest-controller does not validate orchest clusters that uses cri-o. So this has to be changed. so we have to make this possible.
2. For using cri-o, we would also need orchest-controller to deploy buildkit-daemon service for cri-o, not just for containerd as it was today.
3. Had to update the manifesto for buildkitdaemon to have a specialization for cri-o.
4. Had to update orchest-api and orchest-worker to create the manifests of the image-build-task correctly for cri-o.
5. when using containerd runtime, our builkit daemon use containerd workers, but they're not compatible with cri-o. So, we have to change to oci workers in this cases. However, ocis worker does not store images, they just ignore the store=true parameter. In this way, we had two options, to save the image as tarball and load it to the node using some other tool or use push=true and push it directly to the registry. While the second one sounds more intuitive, this would involve architectural changes in the node-agent because currently, the node-agent checks if an image is in the registry or not, and if not, it push to the registry. IT also checks if an image is not required anymore and delete from the registry. That's why we went with the first approach. 
6. Using the first approach had its issues as well:
  a. First, we would need to generate a tarball. This was possible using --output type=docker,name=...,dest=/tmp/output.tar. However, while trying to load it to the registry using podman, we found the following issue: https://github.com/containers/podman/issues/12560. So, decided to identify a version where this bug was corrected.
  b. The podman in the current alpine version is too low, so had to update the alpine pkg to newer versions. 
  c. However, this doesn't solve one problem that is: the image built by buildctl should be available for the node-agent. Solved this by mounting the folder /var/lib/containers/storage inside the image-build-task pod.
  d. But, this generates a conflict with podman newer version > 4.9.0. Had to downgrade podman to version 4.5.1.
  e. After downgrading podman, I was able to build the image, store it on /var/lib/containers/storage and access the new built image from node-agent.
7. Some adaptations in the image-build-task to use crictl for image pull when teh CONTAINER_RUNTIME_SOCKET is cri-o.
8. We have to give the possibility for the user to customize the image that will be used to build images for all runtimes (containerd, docker and crio). There will be a default image (pointing to docker.io/orchest). But, the users will be able to use the envvars DOCKER_IMAGE_BUILDER_IMAGE, CONTAINERD_IMAGE_BUILDER_IMAGE and CRIO_IMAGE_BUILDER_IMAGE. This approach was taken because propagating the env variable from the OrchestCluster -> OrchestComponent -> Orchest API deployment would involve significant changes because we would have to change the OrchestComponent definition. Today is possible to customized images for the core components, but not for components that are created by the API and Celery workers.
9. Made some adaptations on the node-agent to be compatible with crio by using crictl when the ENV Variable CONTAINER_RUNTIME_SOCKET=cri-o
