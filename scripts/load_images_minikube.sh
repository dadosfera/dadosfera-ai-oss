#!/bin/bash

# Define the image list
IMAGES=(
  "orchest/buildkit-daemon:v2023.04.2"
  "orchest/image-builder-buildx:v2023.04.2"
  "orchest/image-puller:v2023.04.2"
  "orchest/auth-server:v2023.04.2"
  "orchest/orchest-webserver:v2023.04.2"
  "orchest/orchest-api:v2023.04.2"
  "orchest/node-agent:v2023.04.2"
  "orchest/orchest-controller:v2023.04.2"
  "orchest/celery-worker:v2023.04.2"
  "orchest/image-builder-buildkit:v2023.04.2"
)

# Loop through and load each image into Minikube
for image in "${IMAGES[@]}"; do
  echo "Loading $image into Minikube..."
  minikube image load "$image" --logtostderr --v=8
done

echo "All images loaded successfully."
