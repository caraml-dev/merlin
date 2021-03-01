#!/bin/bash

set -ex

CHART_PATH="$1"

helm install merlin ${CHART_PATH} --namespace=mlp \
  --values=${CHART_PATH}/values-e2e.yaml \
  --set merlin.image.tag=${GITHUB_HEAD_REF:-${GITHUB_REF#refs/*/}} \
  --dry-run

helm install merlin ${CHART_PATH} --namespace=mlp \
  --values=${CHART_PATH}/values-e2e.yaml \
  --set merlin.image.tag=${GITHUB_HEAD_REF:-${GITHUB_REF#refs/*/}} \
  --wait --timeout=5m

kubectl get all --namespace=mlp

set +ex
