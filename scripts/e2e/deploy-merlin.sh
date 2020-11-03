#!/bin/bash

set -ex

helm install merlin ${MERLIN_CHART_PATH} --namespace=mlp \
  --values=${MERLIN_CHART_PATH}/values-e2e.yaml \
  --set merlin.image.tag=${MERLIN_IMAGE_TAG} \
  --dry-run

helm install merlin ${MERLIN_CHART_PATH} --namespace=mlp \
  --values=${MERLIN_CHART_PATH}/values-e2e.yaml \
  --set merlin.image.tag=${MERLIN_IMAGE_TAG} \
  --wait --timeout=5m

kubectl get all --namespace=mlp

set +ex
