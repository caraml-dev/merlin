#!/bin/bash

CONDA_ENV_PATH="$1"

echo "Processing conda environment file: ${CONDA_ENV_PATH}"
echo "Current conda environment file content:"
cat "${CONDA_ENV_PATH}"

# Remove `merlin-sdk` from conda's pip dependencies
yq --inplace 'del(.dependencies[].pip[] | select(. == "*merlin-sdk*"))' "${CONDA_ENV_PATH}"

# Add `merlin-batch-predictor` with pinned version to conda's pip dependencies, if not exist
yq --inplace 'with(.dependencies[].pip; select(all_c(. != "*merlin-batch-predictor*")) | . += ["merlin-batch-predictor"] )' "${CONDA_ENV_PATH}"

echo "Processed conda environment file content:"
cat "${CONDA_ENV_PATH}"
