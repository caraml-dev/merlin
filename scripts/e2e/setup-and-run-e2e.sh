#!/bin/sh

export K3D_CLUSTER=merlin-cluster
export K3S_VERSION=v1.26.7-k3s1
export INGRESS_HOST=127.0.0.1.nip.io
export LOCAL_REGISTRY_PORT=12345
export LOCAL_REGISTRY=dev.localhost
export DOCKER_REGISTRY=${LOCAL_REGISTRY}:${LOCAL_REGISTRY_PORT}
export VERSION=test-local
export MLP_CHART_VERSION=0.4.18
export MERLIN_CHART_VERSION=0.13.18


# Create k3d cluster and managed registry
k3d registry create $LOCAL_REGISTRY --port $LOCAL_REGISTRY_PORT
k3d cluster create $K3D_CLUSTER --image rancher/k3s:$K3S_VERSION --k3s-arg '--disable=traefik,metrics-server@server:*' --port 80:80@loadbalancer

# Install all dependencies
./setup-cluster.sh $K3D_CLUSTER $INGRESS_HOST

# Build all necessary docker image and import to k3d managed registry
make docker-build
make k3d-import

# Deploy MLP and Merlin
./deploy-merlin.sh $INGRESS_HOST $DOCKER_REGISTRY $VERSION /ref/main ${MLP_CHART_VERSION} ${MERLIN_CHART_VERSION}

# Execute End to end Test
./run-e2e.sh $INGRESS_HOST
