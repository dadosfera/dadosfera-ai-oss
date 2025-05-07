## Development Flow with CRI-O Runtime

This application has been adapted to run using the CRI-O container runtime. The development workflow has been updated accordingly.

### Setting Up Minikube with CRI-O

To start a Minikube cluster using the CRI-O runtime and Docker driver:

```bash
minikube start --driver=docker \
    --container-runtime=cri-o \
    --cpus=2 \
    --memory=8196 \
    --addons=ingress,metrics-server \
    --mount-string="$(pwd):/orchest-dev-repo" --mount
```

Confirm that CRI-O is being used by running the following command:

```bash
kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name}{{"\t"}}{.status.nodeInfo.containerRuntimeVersion}{"\n"}{end}'
```

The output should be something like:

```
minikube	cri-o://1.24.6
```

### Setting Minikube to Use Docker.io as Default Registry

1. Connect to Minikube:

   ```bash
   minikube ssh
   ```

2. Switch to the root user:

   ```bash
   sudo su
   ```

3. Set Docker.io as the default registry:

   ```bash
   echo 'unqualified-search-registries = ["docker.io"]' >> /etc/containers/registries.conf
   systemctl restart crio
   ```

4. Exit the Minikube shell.

### Building Images and Loading to Orchest

#### Load the Core Images

* Build the core images:

  ```bash
  ./scripts/build_container.sh -M -t v2023.04.2 -v
  ```

* Load the images:

  ```bash
  chmod +x scripts/load_images.sh
  ./scripts/load_images.sh
  ```

* Logs for successfully loaded images:

  ```
  I0507 17:08:53.321695   68552 cache_images.go:262] succeeded pushing to: minikube
  I0507 17:08:53.321699   68552 cache_images.go:263] failed pushing to:
  ```

* To check images on Minikube:

  ```bash
  minikube image ls
  ```

### Creating the Orchest Cluster

```bash
kubectl apply -f development/controller.yaml
kubectl apply -f development/cluster.yaml
```

Monitor the application startup by using the following command:
```
kubectl get pods -n orchest -w
```


#### Autoreload (Only for orchest-webserver, auth-server, and orchest-api)

```bash
pnpm i && pnpm run dev -i
```

#### Updating a Component Image

* Build the image:

  ```bash
  ./scripts/build_container.sh -i node-agent -t v2023.04.2 -v
  ```

* Delete the component (if running):

  ```bash
  kubectl delete orchestcomponent node-agent -n orchest
  ```

* Load the image:

  ```bash
  minikube image load orchest/node-agent:v2023.04.2 --logtostderr --v=8
  ```

* Reapply the configuration:

  ```bash
  kubectl apply -f development/cluster.yaml
  ```
