# Minikube

## Installing Minikube

To install Minikube on Linux:

```bash
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube
```

## Single Node Deployment

Create a Minikube cluster with the required resources and add-ons:
```bash
minikube start \
  --cpus 2 \               
  --memory 8196 \
  --addons ingress metrics-server
```

This command starts a Minikube cluster with the Ingress Controller and Metrics Server enabled.
The Ingress Controller is required to run Dadosfera AI OSS.

## Deploying Dadosfera AI OSS Components
1. Install the CRDs and Orchest Controller
```bash
kubectl apply -f deployment/controller/
```

### Create an Cluster
```bash
kubectl apply -f deployment/cluster/
```

## Customizing Application Ingress

If you want to define a custom domain or hostname for your Dadosfera AI OSS deployment, follow these steps:

1. Uncomment and edit the spec.orchest.orchestHost field in your OrchestCluster definition:

For example:
```yaml
apiVersion: orchest.io/v1alpha1
kind: OrchestCluster
metadata:
  name: cluster-1
  namespace: orchest
  annotations:
    # Comment the annotation below if you want orchest-controller
    # to deploy an ingress-controller for your cluster.
    # An ingress controller is required for orchest to work properly.
    controller.orchest.io/deploy-ingress: "false"
spec:
  singleNode: true
  corePriorityClassName: high-priority
  orchest:
    orchestHost: app.example.com
```

2. If a custom host is set, the application will only be accessible via that hostname.
For local development, you can map the hostname to your Minikube IP by adding a line to your /etc/hosts file:
```bash
sudo echo "$(minikube ip) app.example.com" | sudo tee -a /etc/hosts
```