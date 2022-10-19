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


TIMEOUT=120s

install_mlp() {
  echo "::group::MLP Deployment"
  helm upgrade --install --debug mlp mlp/charts/mlp --namespace mlp --create-namespace -f mlp/charts/mlp/values-e2e.yaml \
    --set mlp.image.tag=v1.7.2 \
    --set mlp.apiHost=http://mlp.mlp.${INGRESS_HOST}/v1 \
    --set mlp.mlflowTrackingUrl=http://mlflow.mlp.${INGRESS_HOST} \
    --set mlp.ingress.enabled=true \
    --set mlp.ingress.class=istio \
    --set mlp.ingress.host=mlp.mlp.${INGRESS_HOST} \
    --set mlp.ingress.path="/*" \
    --wait --timeout=${TIMEOUT}

   kubectl apply -f config/mock/message-dumper.yaml

   kubectl rollout status deployment/mlp -n mlp -w --timeout=${TIMEOUT}
}

install_merlin() {
    echo "::group::Merlin Deployment"
  # Merlin uses vault-secret to connect to vault
  kubectl create secret generic vault-secret --namespace=mlp --from-literal=address=http://vault.vault.svc.cluster.local --from-literal=token=root --dry-run=client -o yaml | kubectl apply -f -

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
