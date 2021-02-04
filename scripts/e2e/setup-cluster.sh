#!/bin/bash

set -ex

export CLUSTER_NAME=dev
export KIND_NODE_VERSION=v1.15.12
export VAULT_VERSION=0.7.0

########################################
# Install tools
#
if ! command -v jq &> /dev/null
then
  sudo apt-get update && sudo apt-get install jq
fi
if ! command -v yq &> /dev/null
then
  pip3 install yq
fi

########################################
# Provision KinD cluster
#
kind create cluster --name=${CLUSTER_NAME} --image=kindest/node:${KIND_NODE_VERSION}
kind get kubeconfig --name ${CLUSTER_NAME} --internal > kubeconfig.yaml

########################################
# Install Vault
#
kubectl create namespace vault

# Helm 3 already installed in GitHub Actions
helm repo add hashicorp https://helm.releases.hashicorp.com
helm install vault hashicorp/vault --version=${VAULT_VERSION} --namespace=vault \
  --set injector.enabled=false \
  --set server.dev.enabled=true \
  --set server.dataStorage.enabled=false \
  --set server.resources.requests.cpu=25m \
  --set server.resources.requests.memory=64Mi \
  --set server.affinity=null \
  --set server.tolerations=null \
  --wait --timeout=600s
sleep 15
kubectl wait pod/vault-0 --namespace=vault --for=condition=ready --timeout=600s

# Downgrade to Vault KV secrets engine version 1
kubectl exec vault-0 --namespace=vault -- vault secrets disable secret
kubectl exec vault-0 --namespace=vault -- vault secrets enable -version=1 -path=secret kv

# Write cluster credential to be saved in Vault
cat <<EOF > cluster-credential.json
{
  "name": "$(yq -r '.clusters[0].name' kubeconfig.yaml)",
  "master_ip": "$(yq -r '.clusters[0].cluster.server' kubeconfig.yaml)",
  "certs": "$(yq -r '.clusters[0].cluster."certificate-authority-data"' kubeconfig.yaml | base64 --decode | awk '{printf "%s\\n", $0}')",
  "client_certificate": "$(yq -r '.users[0].user."client-certificate-data"' kubeconfig.yaml | base64 --decode | awk '{printf "%s\\n", $0}')",
  "client_key": "$(yq -r '.users[0].user."client-key-data"' kubeconfig.yaml | base64 --decode | awk '{printf "%s\\n", $0}')"
}
EOF

# Save KinD cluster credential to Vault
kubectl cp cluster-credential.json vault/vault-0:/tmp/cluster-credential.json
kubectl exec vault-0 --namespace=vault -- vault kv put secret/${CLUSTER_NAME} @/tmp/cluster-credential.json

# Clean created credential files
rm kubeconfig.yaml
rm cluster-credential.json

set +ex
