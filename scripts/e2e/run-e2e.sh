#!/bin/bash

export API_PATH="merlin/api"

export INGRESS_HOST=127.0.0.1
export MLP_API_BASEPATH="http://mlp.mlp.${INGRESS_HOST}.nip.io/v1"
export MERLIN_API_BASEPATH="http://merlin.mlp.${INGRESS_HOST}.nip.io"

export E2E_MERLIN_URL="http://merlin.mlp.${INGRESS_HOST}.nip.io"
export E2E_PROJECT_NAME="merlin-e2e-$(git rev-parse --short "$GITHUB_SHA")"


sleep 15

echo "Creating merlin project: e2e-test"
curl "${MLP_API_BASEPATH}/projects" -d '{"name": "e2e-test", "team": "gojek", "stream": "gojek"}'
sleep 5

cd ${API_PATH}
for example in client/examples/*; do
    [[ -e $example ]] || continue
    echo $example
    go run $example/main.go
done

# TODO: Run python/sdk e2e test
cd ../python/sdk
export PATH=$PATH:/root/.pyenv/shims/
make setup

export VENV_HOME_DIR=$(pipenv --venv)
source $VENV_HOME_DIR/bin/activate
make test

kill ${MLP_SVC_PID}
kill ${MERLIN_SVC_PID}
