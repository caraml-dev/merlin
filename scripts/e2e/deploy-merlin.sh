#!/bin/bash

set -ex

CHART_PATH="$1"
export INGRESS_HOST=127.0.0.1
export MERLIN_VERSION=v0.10

curl http://merlin.mlp.${INGRESS_HOST}.nip.io/v1/projects
helm install --debug merlin ${CHART_PATH} --namespace=mlp --values=${CHART_PATH}/values-e2e.yaml \
  --set merlin.image.tag=${MERLIN_VERSION} \
  --set merlin.apiHost=http://merlin.mlp.${INGRESS_HOST}.nip.io/v1 \
  --set merlin.mlpApi.apiHost=http://mlp.mlp.${INGRESS_HOST}.nip.io/v1 \
  --set merlin.ingress.enabled=true \
  --set merlin.ingress.class=istio \
  --set merlin.ingress.host=merlin.mlp.${INGRESS_HOST}.nip.io \
  --set merlin.ingress.path="/*" \
  --set mlflow.ingress.enabled=true \
  --set mlflow.ingress.class=istio \
  --set mlflow.ingress.host=mlflow.mlp.${INGRESS_HOST}.nip.io \
  --set mlflow.ingress.path="/*" \
  --timeout=5m \
  --wait

kubectl get all --namespace=mlp

set +ex
