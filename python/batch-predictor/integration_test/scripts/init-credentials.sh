#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

CLUSTER="$1"

echo "Initializing credentials"
kubectl config set-credentials "${CLUSTER}" --username="$(vault read -field=username secret/${CLUSTER})" --password="$(vault read -field=password secret/${CLUSTER})"
echo "$(vault read -field=certs secret/${CLUSTER})" > ./certs
kubectl config set-cluster "${CLUSTER}" --server="https://$(vault read -field=master_ip secret/${CLUSTER})" --certificate-authority=./certs
kubectl config set-context "${CLUSTER}" --cluster="${CLUSTER}" --user="${CLUSTER}"
kubectl config use-context "${CLUSTER}"
echo "Credential initialized for cluster: ${CLUSTER}"