#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

CHART_PATH="$1"

helm install merlin ${CHART_PATH} --namespace=mlp \
  --values=${CHART_PATH}/values-e2e.yaml \
  --set merlin.image.tag=${GITHUB_REF#refs/heads/} \
  --dry-run

helm install merlin ${CHART_PATH} --namespace=mlp \
  --values=${CHART_PATH}/values-e2e.yaml \
  --set merlin.image.tag=${GITHUB_REF#refs/heads/} \
  --wait --timeout=5m

kubectl get all --namespace=mlp
