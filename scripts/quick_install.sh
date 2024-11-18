#!/bin/sh

# Prerequisites:
# - cluster have been created using k3d
# - cluster have enabled load balancer

export CLUSTER_NAME=merlin-cluster

cd e2e; ./setup-cluster.sh $CLUSTER_NAME

export INGRESS_HOST=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
export DOCKER_REGISTRY=ghcr.io/caraml-dev
export VERSION=0.33.0
export GIT_REF=v0.33.0
export MERLIN_CHART_VERSION=0.13.18
export OAUTH_CLIENT_ID=""

cd e2e; ./deploy-merlin.sh $INGRESS_HOST $DOCKER_REGISTRY $VERSION $GIT_REF $MERLIN_CHART_VERSION $OAUTH_CLIENT_ID
