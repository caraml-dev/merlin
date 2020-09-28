#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

CI_COMMIT_SHORT_SHA="$1"

helm delete --purge test-spark-e2e-${CI_COMMIT_SHORT_SHA}
bq rm -f --project_id project dsp.merlin_iris_result_${CI_COMMIT_SHORT_SHA}