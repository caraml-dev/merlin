#!/usr/bin/env bash
# Bash3 Boilerplate. Copyright (c) 2014, kvz.io

set -o errexit
set -o pipefail
set -o nounset

INGRESS_HOST="$1"
DOCKER_REGISTRY="$2"
VERSION="$3"
GIT_REF="$4"
MERLIN_CHART_VERSION="$5"

TIMEOUT=300s

install_merlin() {
  echo "::group::Merlin Deployment"

  # parse cluster credentials to chart values file
  output=$(yq '.k8s_config' /tmp/temp_k8sconfig.yaml)
  output="$output" yq ".environmentConfigs[0] *= load(\"/tmp/temp_k8sconfig.yaml\") | .imageBuilder.k8sConfig |= env(output)" -i "./values-e2e.yaml"

  helm repo add caraml https://caraml-dev.github.io/helm-charts

  helm upgrade --install --debug merlin caraml/merlin --namespace=caraml --create-namespace \
    --version ${MERLIN_CHART_VERSION} \
    --values values-e2e.yaml \
    --set deployment.image.registry=${DOCKER_REGISTRY} \
    --set deployment.image.repository=merlin \
    --set deployment.image.tag=${VERSION} \
    --set transformer.image=${DOCKER_REGISTRY}/merlin-transformer:${VERSION} \
    --set imageBuilder.dockerRegistry=${DOCKER_REGISTRY} \
    --set imageBuilder.predictionJobBaseImages."3\.7\.*".imageName=${DOCKER_REGISTRY}/merlin/merlin-pyspark-base-py37:${VERSION} \
    --set imageBuilder.predictionJobBaseImages."3\.7\.*".buildContextURI=git://github.com/caraml-dev/merlin.git#${GIT_REF} \
    --set imageBuilder.predictionJobBaseImages."3\.7\.*".dockerfilePath=docker/app.Dockerfile \
    --set imageBuilder.predictionJobBaseImages."3\.7\.*".mainAppPath=/home/spark/merlin-spark-app/main.py \
    --set imageBuilder.predictionJobBaseImages."3\.8\.*".imageName=${DOCKER_REGISTRY}/merlin/merlin-pyspark-base-py38:${VERSION} \
    --set imageBuilder.predictionJobBaseImages."3\.8\.*".buildContextURI=git://github.com/caraml-dev/merlin.git#${GIT_REF} \
    --set imageBuilder.predictionJobBaseImages."3\.8\.*".dockerfilePath=docker/app.Dockerfile \
    --set imageBuilder.predictionJobBaseImages."3\.8\.*".mainAppPath=/home/spark/merlin-spark-app/main.py \
    --set imageBuilder.predictionJobBaseImages."3\.9\.*".imageName=${DOCKER_REGISTRY}/merlin/merlin-pyspark-base-py39:${VERSION} \
    --set imageBuilder.predictionJobBaseImages."3\.9\.*".buildContextURI=git://github.com/caraml-dev/merlin.git#${GIT_REF} \
    --set imageBuilder.predictionJobBaseImages."3\.9\.*".dockerfilePath=docker/app.Dockerfile \
    --set imageBuilder.predictionJobBaseImages."3\.9\.*".mainAppPath=/home/spark/merlin-spark-app/main.py \
    --set imageBuilder.predictionJobBaseImages."3\.10\.*".imageName=${DOCKER_REGISTRY}/merlin/merlin-pyspark-base-py310:${VERSION} \
    --set imageBuilder.predictionJobBaseImages."3\.10\.*".buildContextURI=git://github.com/caraml-dev/merlin.git#${GIT_REF} \
    --set imageBuilder.predictionJobBaseImages."3\.10\.*".dockerfilePath=docker/app.Dockerfile \
    --set imageBuilder.predictionJobBaseImages."3\.10\.*".mainAppPath=/home/spark/merlin-spark-app/main.py \
    --set ingress.host=merlin.caraml.${INGRESS_HOST} \
    --set mlflow.ingress.host=merlin-mlflow.caraml.${INGRESS_HOST} \
    --set mlflow.image.repository=caraml-dev/mlflow \
    --set mlflow.image.tag=1.26.1 \
    --set mlp.deployment.apiHost=http://mlp.caraml.${INGRESS_HOST}/v1 \
    --set mlp.deployment.mlflowTrackingUrl=http://merlin-mlflow.caraml.${INGRESS_HOST} \
    --set mlp.ingress.host=mlp.caraml.${INGRESS_HOST} \
    --wait --timeout=${TIMEOUT}

  kubectl rollout status deployment/mlp -n caraml -w --timeout=${TIMEOUT}
  kubectl rollout status deployment/merlin -n caraml -w --timeout=${TIMEOUT}
}

install_merlin
