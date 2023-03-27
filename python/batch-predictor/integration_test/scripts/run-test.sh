#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SERVICE_ACCOUNT_PATH="$1"
CI_COMMIT_SHORT_SHA="$2"

RELEASE_NAME=test-spark-e2e-${CI_COMMIT_SHORT_SHA}
NAMESPACE=integration-test

wait_spark_done(){
    TIMEOUT=600
    app_state=$(kubectl get sparkapplication -l release="${RELEASE_NAME}" -n $NAMESPACE -o json | jq ".items[0].status.applicationState.state" | tr -d '"')
    until [ "$app_state" = "COMPLETED" ] || [ "$app_state" = "FAILED" ] ; do
        sleep 10
        TIMEOUT=$(( TIMEOUT - 10 ))
        app_state=$(kubectl get sparkapplication -l release="${RELEASE_NAME}" -n $NAMESPACE -o json | jq ".items[0].status.applicationState.state" | tr -d '"')
        echo "Spark app state: $app_state"
        if [[ $TIMEOUT -eq 0 ]];then
            echo "Timeout waiting for spark application to complete"
            kubectl get sparkapplication -l release="${RELEASE_NAME}" -n $NAMESPACE -o yaml
            exit 1
        fi
    done
    
    if [ "$app_state" = "FAILED" ]; then
        echo "Spark application failed"
        kubectl get sparkapplication -l release=${RELEASE_NAME} -n $NAMESPACE -o yaml
        exit 1
    fi
}

BASE_IMAGE=caraml-dev/base-merlin-pyspark:${CI_COMMIT_SHORT_SHA}
echo "Building base spark application docker image: ${BASE_IMAGE}"
docker build -t "${BASE_IMAGE}" -f docker/base.Dockerfile .
echo "Pushing base spark application docker image"
docker push "${BASE_IMAGE}"

USER_IMAGE=caraml-dev/app-merlin-pyspark:${CI_COMMIT_SHORT_SHA}
echo "Building user spark application docker image: ${USER_IMAGE}"
docker build -t "${USER_IMAGE}" -f docker/app.Dockerfile --build-arg BASE_IMAGE=${BASE_IMAGE} \
--build-arg MODEL_URL=gs://bucket-name/e2e/artifacts/model .
echo "Pushing user spark application docker image"
docker push "${USER_IMAGE}"

SERVICE_ACCOUNT_JSON_BASE64="$(cat "$SERVICE_ACCOUNT_PATH" | jq -c "." |  base64 -w 0)"
helm upgrade --install ${RELEASE_NAME} \
integration_test/merlin -f integration_test/merlin/values.yaml \
--namespace $NAMESPACE \
--set image.repository="caraml-dev/app-merlin-pyspark" \
--set-string image.tag="${CI_COMMIT_SHORT_SHA}" \
--set-string serviceAccount="$SERVICE_ACCOUNT_JSON_BASE64" \
--set jobSpec.bigquerySink.table="project.dataset.table_iris_result_${CI_COMMIT_SHORT_SHA}"

wait_spark_done


bq query --project_id project \
--use_legacy_sql=false \
"SELECT sepal_length, sepal_width, petal_length, petal_width, prediction FROM project.dataset.table_iris_result_${CI_COMMIT_SHORT_SHA} LIMIT 10"
