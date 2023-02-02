#!/usr/bin/env bash
# Bash3 Boilerplate. Copyright (c) 2014, kvz.io

set -o errexit
set -o pipefail
set -o nounset

INGRESS_HOST="$1"
DOCKER_REGISTRY="$2"
CHART_PATH="$3"
VERSION="$4"
GIT_REF="$5"
MLP_CHART_VERSION="$6"

TIMEOUT=300s

install_mlp() {
  echo "::group::MLP Deployment"

  helm repo add caraml https://caraml-dev.github.io/helm-charts

  helm upgrade --install --debug mlp caraml/mlp --namespace mlp --create-namespace \
    --version ${MLP_CHART_VERSION} \
    --set fullnameOverride=mlp \
    --set deployment.apiHost=http://mlp.mlp.${INGRESS_HOST}/v1 \
    --set deployment.mlflowTrackingUrl=http://mlflow.mlp.${INGRESS_HOST} \
    --set ingress.enabled=true \
    --set ingress.class=istio \
    --set ingress.host=mlp.mlp.${INGRESS_HOST} \
    --wait --timeout=${TIMEOUT}

   kubectl apply -f config/mock/message-dumper.yaml

   kubectl rollout status deployment/mlp -n mlp -w --timeout=${TIMEOUT}
}

install_merlin() {
    echo "::group::Merlin Deployment"
# build in json first, then convert to yaml
  output=$(yq e -o json '.k8s_config' /tmp/temp_k8sconfig.yaml | jq -r -M -c .)
  output="$output" yq '.merlin.environmentConfigs[0] *= load("/tmp/temp_k8sconfig.yaml") | .merlin.imageBuilder.k8sConfig |= strenv(output)' -i "${CHART_PATH}/values-e2e.yaml"

  helm upgrade --install --debug merlin ${CHART_PATH} --namespace=mlp --create-namespace -f ${CHART_PATH}/values-e2e.yaml \
    --set merlin.image.registry=${DOCKER_REGISTRY} \
    --set merlin.image.tag=${VERSION} \
    --set merlin.transformer.image=${DOCKER_REGISTRY}/merlin-transformer:${VERSION} \
    --set merlin.imageBuilder.dockerRegistry=${DOCKER_REGISTRY} \
    --set merlin.imageBuilder.predictionJobBaseImages."3\.7\.*".imageName=${DOCKER_REGISTRY}/merlin/merlin-pyspark-base-py37:${VERSION} \
    --set merlin.imageBuilder.predictionJobBaseImages."3\.7\.*".buildContextURI=git://github.com/gojek/merlin.git#${GIT_REF} \
    --set merlin.imageBuilder.predictionJobBaseImages."3\.7\.*".dockerfilePath=docker/app.Dockerfile \
    --set merlin.imageBuilder.predictionJobBaseImages."3\.7\.*".mainAppPath=/home/spark/merlin-spark-app/main.py \
    --set merlin.imageBuilder.predictionJobBaseImages."3\.8\.*".imageName=${DOCKER_REGISTRY}/merlin/merlin-pyspark-base-py38:${VERSION} \
    --set merlin.imageBuilder.predictionJobBaseImages."3\.8\.*".buildContextURI=git://github.com/gojek/merlin.git#${GIT_REF} \
    --set merlin.imageBuilder.predictionJobBaseImages."3\.8\.*".dockerfilePath=docker/app.Dockerfile \
    --set merlin.imageBuilder.predictionJobBaseImages."3\.8\.*".mainAppPath=/home/spark/merlin-spark-app/main.py \
    --set merlin.imageBuilder.predictionJobBaseImages."3\.9\.*".imageName=${DOCKER_REGISTRY}/merlin/merlin-pyspark-base-py39:${VERSION} \
    --set merlin.imageBuilder.predictionJobBaseImages."3\.9\.*".buildContextURI=git://github.com/gojek/merlin.git#${GIT_REF} \
    --set merlin.imageBuilder.predictionJobBaseImages."3\.9\.*".dockerfilePath=docker/app.Dockerfile \
    --set merlin.imageBuilder.predictionJobBaseImages."3\.9\.*".mainAppPath=/home/spark/merlin-spark-app/main.py \
    --set merlin.imageBuilder.predictionJobBaseImages."3\.10\.*".imageName=${DOCKER_REGISTRY}/merlin/merlin-pyspark-base-py310:${VERSION} \
    --set merlin.imageBuilder.predictionJobBaseImages."3\.10\.*".buildContextURI=git://github.com/gojek/merlin.git#${GIT_REF} \
    --set merlin.imageBuilder.predictionJobBaseImages."3\.10\.*".dockerfilePath=docker/app.Dockerfile \
    --set merlin.imageBuilder.predictionJobBaseImages."3\.10\.*".mainAppPath=/home/spark/merlin-spark-app/main.py \
    --set merlin.apiHost=http://merlin.mlp.${INGRESS_HOST}/v1 \
    --set merlin.ingress.host=merlin.mlp.${INGRESS_HOST} \
    --set mlflow.ingress.host=merlin-mlflow.mlp.${INGRESS_HOST} \
    --wait --timeout=${TIMEOUT}

  kubectl rollout status deployment/merlin -n mlp -w --timeout=${TIMEOUT}
}

install_mlp
install_merlin
