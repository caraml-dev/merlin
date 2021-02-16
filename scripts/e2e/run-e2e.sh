#!/bin/bash

export API_PATH="merlin/api"

export INGRESS_HOST=$(kubectl get po -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].status.hostIP}')
export MLP_API_BASEPATH="http://mlp.mlp.${INGRESS_HOST}.nip.io/v1"
export MERLIN_API_BASEPATH="http://merlin.mlp.${INGRESS_HOST}.nip.io"

export E2E_MLP_URL="http://mlp.mlp.${INGRESS_HOST}.nip.io:8080"
export E2E_MERLIN_URL="http://merlin.mlp.${INGRESS_HOST}.nip.io"
export E2E_PROJECT_NAME="merlin-e2e-${GITHUB_SHA::8}"

curl "${E2E_MERLIN_URL}/v1/projects"
curl -X POST "${E2E_MLP_URL}/v1/projects" -d "{\"name\": \"${E2E_PROJECT_NAME}\", \"team\": \"gojek\", \"stream\": \"gojek\", \"mlflow_tracking_url\": \"${E2E_MLFLOW_URL}\"}"
curl "${E2E_MERLIN_URL}/v1/projects"

cd ./merlin/python/sdk
pip install pipenv
pipenv install --dev --skip-lock
pipenv run pytest -n=8 -W=ignore --cov=merlin test/integration_test.py

# echo "Creating merlin project: e2e-test"
# curl "${MLP_API_BASEPATH}/projects" -d '{"name": "e2e-test", "team": "gojek", "stream": "gojek"}'
# sleep 5

# cd ${API_PATH}
# for example in client/examples/*; do
#     [[ -e $example ]] || continue
#     echo $example
#     go run $example/main.go
# done

# # TODO: Run python/sdk e2e test
# cd ../python/sdk
# export PATH=$PATH:/root/.pyenv/shims/
# make setup
