#!/bin/bash

# This script is designed to manipulate a Conda environment file to add Merlin dependencies.
#
# Usage:
# ./process_conda_env.sh CONDA_ENV_PATH MERLIN_DEP MERLIN_DEP_CONSTRAINT
#
# Input:
# CONDA_ENV_PATH: Path to the Conda environment YAML file.
# MERLIN_DEP: The dependency to be added or removed.
# MERLIN_DEP_CONSTRAINT: The constraint for the dependency, such as a version specification.


CONDA_ENV_PATH="$1"
MERLIN_DEP="$2"
MERLIN_DEP_CONSTRAINT="$3"

echo "Processing conda environment file: ${CONDA_ENV_PATH}"
echo "Current conda environment file content:"
cat "${CONDA_ENV_PATH}"

# Remove `mlflow` constraint
yq --inplace 'del(.dependencies[].pip[] | select(. == "mlflow*"))' "${CONDA_ENV_PATH}"
yq --inplace ".dependencies[].pip += [\"mlflow\"]" "${CONDA_ENV_PATH}"

# Remove `merlin-sdk` from conda's pip dependencies
yq --inplace 'del(.dependencies[].pip[] | select(. == "*merlin-sdk*"))' "${CONDA_ENV_PATH}"

# Add `${MERLIN_DEP}` with its constaint (`${MERLIN_DEP_CONSTRAINT}`) to conda's pip dependencies, if not exist
yq --inplace "with(.dependencies[].pip; select(all_c(. != \"*${MERLIN_DEP}*\")) | . += [\"${MERLIN_DEP}${MERLIN_DEP_CONSTRAINT}\"] )" "${CONDA_ENV_PATH}"

echo "Processed conda environment file content:"
cat "${CONDA_ENV_PATH}"
