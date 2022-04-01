#!/usr/bin/env bash
# Bash3 Boilerplate. Copyright (c) 2014, kvz.io

set -o errexit
set -o pipefail
set -o nounset

export API_PATH="merlin/api"

export INGRESS_HOST=127.0.0.1
export MLP_API_BASEPATH="http://mlp.mlp.${INGRESS_HOST}.nip.io/v1"
export MERLIN_API_BASEPATH="http://merlin.mlp.${INGRESS_HOST}.nip.io"

export E2E_MLP_URL="http://mlp.mlp.${INGRESS_HOST}.nip.io"
export E2E_MERLIN_URL="http://merlin.mlp.${INGRESS_HOST}.nip.io"
export E2E_MLFLOW_URL="http://merlin-mlflow.mlp.${INGRESS_HOST}.nip.io"
export E2E_PROJECT_NAME="merlin-e2e"
export E2E_MERLIN_ENVIRONMENT="dev"

export AWS_ACCESS_KEY_ID=YOURACCESSKEY
export AWS_SECRET_ACCESS_KEY=YOURSECRETKEY
export MLFLOW_S3_ENDPOINT_URL=http://minio.minio.${INGRESS_HOST}.nip.io

curl "${E2E_MERLIN_URL}/v1/projects"
curl -X POST "${E2E_MLP_URL}/v1/projects" -d "{\"name\": \"${E2E_PROJECT_NAME}\", \"team\": \"gojek\", \"stream\": \"gojek\", \"mlflow_tracking_url\": \"${E2E_MLFLOW_URL}\"}"
curl "${E2E_MERLIN_URL}/v1/projects"

cd ../../python/sdk
pip install pipenv
pipenv install --dev --skip-lock
pipenv run pytest -W=ignore --cov=merlin -m "not (feast or batch or serving or pytorch or pyfunc or local_server_test)"
