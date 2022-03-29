#!/usr/bin/env bash
# Bash3 Boilerplate. Copyright (c) 2014, kvz.io

set -o errexit
set -o pipefail
set -o nounset

# Software requirements:
# - yq 4.24.2 
# - helm 3
# - k3d
# - kubectl

# Prerequisites:
# - cluster have been created using k3d

export VAULT_VERSION=0.7.0
export ISTIO_VERSION=1.13.2
export KNATIVE_VERSION=1.3.0
export CERT_MANAGER_VERSION=1.1.0
export MINIO_VERSION=8.0.10 
export KSERVE_VERSION=0.8.0

export CLUSTER_NAME=merlin-cluster
export TIMEOUT=120s


install_vault() {
    helm repo add hashicorp https://helm.releases.hashicorp.com
    helm repo update
    helm upgrade --install vault hashicorp/vault --version=${VAULT_VERSION} \
        -f config/vault/values.yaml \
        --namespace=vault --create-namespace \
        --wait --timeout=${TIMEOUT}

    kubectl wait pod/vault-0 --namespace=vault --for=condition=ready --timeout=${TIMEOUT}

    # Downgrade to Vault KV secrets engine version 1
    kubectl exec vault-0 --namespace=vault -- vault secrets disable secret
    kubectl exec vault-0 --namespace=vault -- vault secrets enable -version=1 -path=secret kv

    k3d kubeconfig get merlin-e2e > kubeconfig.yaml

    # Write cluster credential to be saved in Vault
    cat <<EOF > cluster-credential.json
{
"name": "$(yq '.clusters[0].name' kubeconfig.yaml)",
"master_ip": "$(yq '.clusters[0].cluster.server' kubeconfig.yaml)",
"certs": "$(yq '.clusters[0].cluster."certificate-authority-data"' kubeconfig.yaml | base64 --decode | awk '{printf "%s\\n", $0}')",
"client_certificate": "$(yq '.users[0].user."client-certificate-data"' kubeconfig.yaml | base64 --decode | awk '{printf "%s\\n", $0}')",
"client_key": "$(yq '.users[0].user."client-key-data"' kubeconfig.yaml | base64 --decode | awk '{printf "%s\\n", $0}')"
}
EOF

    # Save KinD cluster credential to Vault
    kubectl cp cluster-credential.json vault/vault-0:/tmp/cluster-credential.json
    kubectl exec vault-0 --namespace=vault -- vault kv put secret/${CLUSTER_NAME} @/tmp/cluster-credential.json

    # Clean created credential files
    rm kubeconfig.yaml
    rm cluster-credential.json
}

install_istio() {
    helm repo add istio https://istio-release.storage.googleapis.com/charts
    helm repo update
    helm upgrade --install istio-base istio/base --version=${ISTIO_VERSION} -n istio-system --create-namespace
    helm upgrade --install istiod istio/istiod --version=${ISTIO_VERSION} -n istio-system --create-namespace \
        -f config/istio/istiod.yaml \
        --wait --timeout=${TIMEOUT}
    kubectl rollout status deployment/istiod -w -n istio-system --timeout=${TIMEOUT}

    helm upgrade --install istio-ingress istio/gateway -n istio-system --create-namespace \
        -f config/istio/ingressgateway.yaml \
        --wait --timeout=${TIMEOUT}
    kubectl rollout status deployment/istio-ingress -n istio-system -w --timeout=${TIMEOUT}
}

install_knative() {
    # Install CRD
    kubectl apply -f https://github.com/knative/serving/releases/download/knative-v${KNATIVE_VERSION}/serving-crds.yaml

    # Install knative serving
    wget https://github.com/knative/serving/releases/download/knative-v${KNATIVE_VERSION}/serving-core.yaml -O config/knative/serving-core.yaml
    kubectl apply -k config/knative

    # Install knative-istio
    kubectl apply -f https://github.com/knative/net-istio/releases/download/knative-v${KNATIVE_VERSION}/net-istio.yaml

    kubectl rollout status deployment/autoscaler -n knative-serving -w --timeout=${TIMEOUT}
    kubectl rollout status deployment/controller -n knative-serving -w --timeout=${TIMEOUT}
    kubectl rollout status deployment/activator -n knative-serving -w --timeout=${TIMEOUT}
    kubectl rollout status deployment/domain-mapping -n knative-serving -w --timeout=${TIMEOUT}
    kubectl rollout status deployment/domainmapping-webhook -n knative-serving -w --timeout=${TIMEOUT}
    kubectl rollout status deployment/webhook -n knative-serving -w --timeout=${TIMEOUT}
    kubectl rollout status deployment/net-istio-controller -n knative-serving -w --timeout=${TIMEOUT}
    kubectl rollout status deployment/net-istio-webhook -n knative-serving -w --timeout=${TIMEOUT}
}

install_cert_manager() {
    kubectl apply --filename=https://github.com/jetstack/cert-manager/releases/download/v${CERT_MANAGER_VERSION}/cert-manager.yaml
    
    kubectl rollout status deployment/cert-manager-webhook -n cert-manager -w --timeout=${TIMEOUT}
    kubectl rollout status deployment/cert-manager-cainjector -n cert-manager -w --timeout=${TIMEOUT}
    kubectl rollout status deployment/cert-manager -n cert-manager -w --timeout=${TIMEOUT}
}

install_minio() {
    helm repo add minio https://helm.min.io/
    helm upgrade --install minio minio/minio --version=${MINIO_VERSION} -f config/minio/values.yaml \
        --namespace=minio --create-namespace \
        --set accessKey=YOURACCESSKEY --set secretKey=YOURSECRETKEY

    kubectl rollout status deployment/minio -n minio -w --timeout=${TIMEOUT}
}

install_kserve() {
    wget https://raw.githubusercontent.com/kserve/kserve/master/install/v${KSERVE_VERSION}/kserve.yaml -O config/kserve/kserve.yaml

    kubectl apply -k config/kserve
    kubectl rollout status statefulset/kserve-controller-manager -w --timeout=${TIMEOUT}

    kubectl apply -f https://raw.githubusercontent.com/kserve/kserve/master/install/v${KSERVE_VERSION}/kserve-runtimes.yaml
}

install_istio
install_knative
install_vault
install_cert_manager
install_minio
install_kserve
