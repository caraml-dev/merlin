#!/usr/bin/env bash
# Bash3 Boilerplate. Copyright (c) 2014, kvz.io

set -o errexit
set -o pipefail
set -o nounset

CHART_PATH="$1"
VERSION="$2"
INGRESS_HOST=127.0.0.1
TIMEOUT=120s

install_mlp() {
  helm upgrade --install mlp mlp/charts/mlp --namespace mlp --create-namespace -f mlp/charts/mlp/values-e2e.yaml \
    --set mlp.image.tag=main \
    --set mlp.apiHost=http://mlp.mlp.${INGRESS_HOST}.nip.io/v1 \
    --set mlp.mlflowTrackingUrl=http://mlflow.mlp.${INGRESS_HOST}.nip.io \
    --set mlp.ingress.enabled=true \
    --set mlp.ingress.class=istio \
    --set mlp.ingress.host=mlp.mlp.${INGRESS_HOST}.nip.io \
    --set mlp.ingress.path="/*" \
    --wait --timeout=${TIMEOUT}

   kubectl rollout status deployment/mlp -n mlp -w --timeout=${TIMEOUT}
}

install_merlin() {
  # Merlin uses vault-secret to connect to vault
  kubectl create secret generic vault-secret --namespace=mlp --from-literal=address=http://vault.vault.svc.cluster.local --from-literal=token=root

  helm upgrade --install --debug merlin ${CHART_PATH} --namespace=mlp --create-namespace -f ${CHART_PATH}/values-e2e.yaml \
    --set merlin.image.tag=${VERSION} \
    --set merlin.transformer.image=dev.local/merlin-transformer:${VERSION} \
    --set merlin.apiHost=http://merlin.mlp.${INGRESS_HOST}.nip.io/v1 \
    --set merlin.ingress.host=merlin.mlp.${INGRESS_HOST}.nip.io \
    --set mlflow.ingress.host=merlin-mlflow.mlp.${INGRESS_HOST}.nip.io \
    --wait --timeout=${TIMEOUT} \

  kubectl rollout status deployment/merlin -n mlp -w --timeout=${TIMEOUT}
}

install_mlp
install_merlin
