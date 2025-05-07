orchest uninstall
eval $(minikube -p minikube docker-env)
echo "$MINIKUBE_ACTIVE_DOCKERD"
export TAG="v2023.01.0"
scripts/build_container.sh -M -t $TAG -o $TAG
sleep 30
orchest install --dev
