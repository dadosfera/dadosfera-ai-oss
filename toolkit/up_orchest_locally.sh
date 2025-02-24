minikube start --cpus 2 --memory 4800 --addons ingress metrics-server --mount-string="$(pwd):/orchest-dev-repo" --mount

sleep 30

eval $(minikube -p minikube docker-env)      

echo "$MINIKUBE_ACTIVE_DOCKERD"  

export TAG="v2023.01.0" 
# export TAG="$(orchest version --latest)" 

echo "TAG"
echo $TAG
echo "TAG"

scripts/build_container.sh -M -t $TAG -o $TAG

sleep 30

orchest install --dev  