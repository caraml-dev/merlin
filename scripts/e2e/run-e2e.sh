#!/bin/bash

export API_PATH="merlin/api"

export INGRESS_HOST=127.0.0.1
export MLP_API_BASEPATH="http://mlp.mlp.${INGRESS_HOST}.nip.io/v1"
export MERLIN_API_BASEPATH="http://merlin.mlp.${INGRESS_HOST}.nip.io"

export E2E_MLP_URL="http://mlp.mlp.${INGRESS_HOST}.nip.io"
export E2E_MERLIN_URL="http://merlin.mlp.${INGRESS_HOST}.nip.io"
export E2E_PROJECT_NAME="merlin-e2e-${GITHUB_SHA::8}"

export AWS_ACCESS_KEY_ID=YOURACCESSKEY
export AWS_SECRET_ACCESS_KEY=YOURSECRETKEY
export MLFLOW_S3_ENDPOINT_URL=http://minio.minio.${INGRESS_HOST}.nip.io

curl "${E2E_MERLIN_URL}/v1/projects"
curl -X POST "${E2E_MLP_URL}/v1/projects" -d "{\"name\": \"${E2E_PROJECT_NAME}\", \"team\": \"gojek\", \"stream\": \"gojek\", \"mlflow_tracking_url\": \"${E2E_MLFLOW_URL}\"}"
curl "${E2E_MERLIN_URL}/v1/projects"

cd ./merlin/python/sdk
pip install pipenv
pipenv install --dev --skip-lock
pipenv run pytest -n=1 -W=ignore --cov=merlin test/integration_test.py -k 'not test_standard_transformer_feast_pyfunc'
