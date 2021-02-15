#!/bin/bash

set -ex

CHART_PATH="$1"

kubectl create secret generic vault-secret --namespace=mlp --from-literal=address=http://vault.vault.svc.cluster.local --from-literal=token=root

helm install merlin ${CHART_PATH} --namespace=mlp --values=${CHART_PATH}/values-e2e.yaml \
  --set merlin.image.tag=${MERLIN_VERSION} \
  --set merlin.oauthClientID=${OAUTH_CLIENT_ID} \
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
