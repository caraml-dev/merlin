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
# pipenv run pytest -n=3 -W=ignore --cov=merlin test/integration_test.py
pipenv run pytest test/integration_test.py::test_trasformer_pytorch_logger
kubectl get pod -l gojek.com/app=transformer-logger -n ${E2E_PROJECT_NAME}
echo "---------------Get Predictor Initializer -------------------------"
kubectl logs -l  gojek.com/app=transformer-logger,component=predictor -c storage-initializer -n ${E2E_PROJECT_NAME}
echo "---------------Get Inferenceservice -------------------------"
kubectl describe inferenceservice -l gojek.com/app=transformer-logger -n ${E2E_PROJECT_NAME}
echo "---------------Get Revision -------------------------"
kubectl get revision -l gojek.com/app=transformer-logger -n ${E2E_PROJECT_NAME} -o yaml
echo "---------------Get Kfserving controller log -------------------------"
kubectl logs kfserving-controller-manager-0 manager -n kfserving-system
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
