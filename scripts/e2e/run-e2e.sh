#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

export API_PATH="$1"

export MERLIN_API_BASEPATH="http://127.0.0.1:8080/v1"

kubectl port-forward --namespace=mlp svc/merlin 8080 &
sleep 5

curl "${MERLIN_API_BASEPATH}/projects" -d '{"name": "e2e-test", "team": "gojek", "stream": "gojek"}'
sleep 5

cd ${API_PATH}
for example in client/examples/*; do
    [[ -e $example ]] || continue
    echo $example
    go run $example/main.go
done
