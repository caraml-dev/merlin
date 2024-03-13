#!/bin/bash

CONDA_ENV_PATH="$1"

echo "Processing conda environment file: ${CONDA_ENV_PATH}"
echo "Current conda environment file content:"
cat "${CONDA_ENV_PATH}"

# Remove `merlin-sdk` from conda's pip dependencies
yq --inplace 'del(.dependencies[].pip[] | select(. == "*merlin-sdk*"))' "${CONDA_ENV_PATH}"

# Add `merlin-pyfunc-server` to conda's pip dependencies
yq --inplace 'with(.dependencies[].pip; select(all_c(. != "*merlin-pyfunc-server*")) | . += ["merlin-pyfunc-server==0.40.2.dev20"] )' "${CONDA_ENV_PATH}"

echo "Processed conda environment file content:"
cat "${CONDA_ENV_PATH}"
