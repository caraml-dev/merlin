#!/bin/bash

set -ex

CHART_PATH="$1"
export INGRESS_HOST=$(kubectl get po -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].status.hostIP}')
export MERLIN_VERSION=v0.10.0


helm install --debug merlin ${CHART_PATH} --namespace=mlp --values=${CHART_PATH}/values-e2e.yaml \
  --set merlin.image.tag=${MERLIN_VERSION} \
  --set merlin.apiHost=http://merlin.mlp.${INGRESS_HOST}.nip.io/v1 \
  --set merlin.mlpApi.apiHost=http://mlp.mlp.svc.cluster.local:8080/v1 \
  --set merlin.ingress.enabled=true \
  --set merlin.ingress.class=istio \
  --set merlin.ingress.host=merlin.mlp.${INGRESS_HOST}.nip.io \
  --set merlin.ingress.path="/*" \
  --set mlflow.ingress.enabled=true \
  --set mlflow.ingress.class=istio \
  --set mlflow.ingress.host=merlin-mlflow.mlp.${INGRESS_HOST}.nip.io \
  --set mlflow.extraEnvs.MLFLOW_S3_ENDPOINT_URL=minio.minio.${INGRESS_HOST}.nip.io \
  --set mlflow.ingress.path="/*" \
  --set mlflow.postgresql.requests.cpu="25m" \
  --set mlflow.postgresql.requests.memory="256Mi" \
  --timeout=5m \
  --dry-run

helm install --debug merlin ${CHART_PATH} --namespace=mlp --values=${CHART_PATH}/values-e2e.yaml \
  --set merlin.image.tag=${MERLIN_VERSION} \
  --set merlin.apiHost=http://merlin.mlp.${INGRESS_HOST}.nip.io/v1 \
  --set merlin.mlpApi.apiHost=http://mlp.mlp.svc.cluster.local:8080/v1 \
  --set merlin.ingress.enabled=true \
  --set merlin.ingress.class=istio \
  --set merlin.ingress.host=merlin.mlp.${INGRESS_HOST}.nip.io \
  --set merlin.ingress.path="/*" \
  --set mlflow.ingress.enabled=true \
  --set mlflow.ingress.class=istio \
  --set mlflow.ingress.host=merlin-mlflow.mlp.${INGRESS_HOST}.nip.io \
  --set mlflow.extraEnvs.MLFLOW_S3_ENDPOINT_URL=minio.minio.${INGRESS_HOST}.nip.io \
  --set mlflow.ingress.path="/*" \
  --set mlflow.resources.requests.cpu="25m" \
  --set mlflow.resources.requests.memory="256Mi" \
  --timeout=5m \
  --wait

kubectl get service istio-ingressgateway --namespace=istio-system -o yaml

set +ex
