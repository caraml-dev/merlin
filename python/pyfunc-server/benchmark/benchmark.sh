#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

trap "kill 0" EXIT

mkdir -p prometheus_multiproc_dir
export PROMETHEUS_MULTIPROC_DIR=prometheus_multiproc_dir

python -m pyfuncserver --model_dir echo-model/model --workers 1 &

sleep 5

echo "=================================="
echo "============small payload========="
echo "=================================="
cat small.cfg | vegeta attack -rate 100 -duration=60s | vegeta report

echo "=================================="
echo "============medium payload========="
echo "=================================="
cat medium.cfg | vegeta attack -rate 100 -duration=60s | vegeta report

echo "=================================="
echo "============large payload========="
echo "=================================="
cat large.cfg | vegeta attack -rate 100 -duration=60s | vegeta report

echo "=================================="
echo "===========max throughput========="
echo "=================================="
cat large.cfg | vegeta attack -rate 0 -max-workers=8 -duration=60s | vegeta report