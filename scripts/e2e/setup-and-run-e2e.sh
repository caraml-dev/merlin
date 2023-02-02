export K3D_CLUSTER=merlin-cluster
export INGRESS_HOST=127.0.0.1.nip.io
export LOCAL_REGISTRY_PORT=12345
export LOCAL_REGISTRY=dev.localhost
export DOCKER_REGISTRY=${LOCAL_REGISTRY}:${LOCAL_REGISTRY_PORT}
export VERSION=test-local


# Create k3d cluster and managed registry
k3d registry create $LOCAL_REGISTRY --port $LOCAL_REGISTRY_PORT
k3d cluster create $K3D_CLUSTER --image rancher/k3s:v1.22.15-k3s1 --k3s-arg '--no-deploy=traefik,metrics-server@server:*' --port 80:80@loadbalancer
# Install all dependencies
./setup-cluster.sh $K3D_CLUSTER $INGRESS_HOST

# Build all necessary docker image and import to k3d managed registry
make docker-build
make k3d-import

# Deploy MLP and Merlin
./deploy-merlin.sh $INGRESS_HOST $DOCKER_REGISTRY ../../charts/merlin $VERSION /ref/rework_ci

# Execute End to end Test
./run-e2e.sh $INGRESS_HOST
