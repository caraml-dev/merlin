#!/usr/bin/env bash
# Bash3 Boilerplate. Copyright (c) 2014, kvz.io

set -o errexit
set -o pipefail
set -o nounset

INGRESS_HOST=$1
PYTHON_VERSION=$2

API_PATH="merlin/api"
MLP_API_BASEPATH="http://mlp.mlp.${INGRESS_HOST}/v1"
MERLIN_API_BASEPATH="http://merlin.mlp.${INGRESS_HOST}"

# These configuration must be exported since it's being read by conftest.py
export E2E_MLP_URL="http://mlp.mlp.${INGRESS_HOST}"
export E2E_MERLIN_URL="http://merlin.mlp.${INGRESS_HOST}"
export E2E_MLFLOW_URL="http://merlin-mlflow.mlp.${INGRESS_HOST}"
export E2E_PROJECT_NAME="merlin-e2e"
export E2E_MERLIN_ENVIRONMENT="dev"
export AWS_ACCESS_KEY_ID=YOURACCESSKEY
export AWS_SECRET_ACCESS_KEY=YOURSECRETKEY
export MLFLOW_S3_ENDPOINT_URL=http://minio.minio.${INGRESS_HOST}

curl "${E2E_MERLIN_URL}/v1/projects"
curl -X POST "${E2E_MLP_URL}/v1/projects" -d "{\"name\": \"${E2E_PROJECT_NAME}\", \"team\": \"gojek\", \"stream\": \"gojek\", \"mlflow_tracking_url\": \"${E2E_MLFLOW_URL}\"}"
curl "${E2E_MERLIN_URL}/v1/projects"

kubectl create namespace ${E2E_PROJECT_NAME} --dry-run=client -o yaml | kubectl apply -f -

cd ../../python/sdk
pip install pipenv==2022.8.19
pipenv install --dev --skip-lock --python ${PYTHON_VERSION}
pipenv run pytest -n=4 -W=ignore --cov=merlin -m "not (feast or local_server_test or cli or customtransformer)"
