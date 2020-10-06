#!/bin/bash

export API_PATH="$1"

export MLP_API_BASEPATH="http://127.0.0.1:8081/v1"
export MERLIN_API_BASEPATH="http://127.0.0.1:8080/v1"

kubectl port-forward --namespace=mlp svc/mlp 8081:8080 &
MLP_SVC_PID=$!
kubectl port-forward --namespace=mlp svc/merlin 8080 &
MERLIN_SVC_PID=$!

sleep 15

echo "Creating merlin project: e2e-test"
curl "${MLP_API_BASEPATH}/projects" -d '{"name": "e2e-test-2", "team": "gojek", "stream": "gojek"}'
sleep 5

cd ${API_PATH}
for example in client/examples/*; do
    [[ -e $example ]] || continue
    echo $example
    go run $example/main.go
done

# TODO: Run python/sdk e2e test

kill ${MLP_SVC_PID}
kill ${MERLIN_SVC_PID}
