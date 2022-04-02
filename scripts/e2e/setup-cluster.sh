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

CLUSTER_NAME=$1
INGRESS_HOST=$2

ISTIO_VERSION=1.13.2
KNATIVE_VERSION=1.3.0
CERT_MANAGER_VERSION=1.7.2
MINIO_VERSION=3.6.3 
KSERVE_VERSION=0.8.0
VAULT_VERSION=0.19.0
TIMEOUT=180s


add_helm_repo() {
    helm repo add hashicorp https://helm.releases.hashicorp.com
    helm repo add minio https://charts.min.io/
    helm repo add istio https://istio-release.storage.googleapis.com/charts
    helm repo update
}

install_vault() {
    echo "::group::Vault Deployment"
    helm upgrade --install vault hashicorp/vault --version=${VAULT_VERSION} \
        -f config/vault/values.yaml \
        --namespace=vault --create-namespace \
        --wait --timeout=${TIMEOUT}

    kubectl wait pod/vault-0 --namespace=vault --for=condition=ready --timeout=${TIMEOUT}
}

store_cluster_secret() {
    echo "::group::Setting up Cluster Secret"
    # Downgrade to Vault KV secrets engine version 1
    kubectl exec vault-0 --namespace=vault -- vault secrets disable secret
    kubectl exec vault-0 --namespace=vault -- vault secrets enable -version=1 -path=secret kv

    k3d kubeconfig get $CLUSTER_NAME > kubeconfig.yaml

    # Write cluster credential to be saved in Vault
    cat <<EOF > cluster-credential.json
{
"name": "$(yq '.clusters[0].name' kubeconfig.yaml)",
"master_ip": "kubernetes.default:443",
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
    echo "::group::Istio Deployment"
    helm upgrade --install istio-base istio/base --version=${ISTIO_VERSION} -n istio-system --create-namespace
    helm upgrade --install istiod istio/istiod --version=${ISTIO_VERSION} -n istio-system --create-namespace \
        -f config/istio/istiod.yaml --timeout=${TIMEOUT}
    
    helm upgrade --install istio-ingressgateway istio/gateway -n istio-system --create-namespace \
        -f config/istio/ingress-gateway.yaml --timeout=${TIMEOUT}
    
    helm upgrade --install cluster-local-gateway istio/gateway -n istio-system --create-namespace \
        -f config/istio/clusterlocal-gateway.yaml --timeout=${TIMEOUT}

    kubectl rollout status deployment/istio-ingressgateway -n istio-system -w --timeout=${TIMEOUT}
    kubectl rollout status deployment/istiod -w -n istio-system --timeout=${TIMEOUT}
    kubectl rollout status deployment/cluster-local-gateway -n istio-system -w --timeout=${TIMEOUT}
}

install_knative() {
    echo "::group::Knative Deployment"
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
    echo "::group::Cert Manager Deployment"
    kubectl apply --filename=https://github.com/jetstack/cert-manager/releases/download/v${CERT_MANAGER_VERSION}/cert-manager.yaml

    kubectl rollout status deployment/cert-manager-webhook -n cert-manager -w --timeout=${TIMEOUT}
    kubectl rollout status deployment/cert-manager-cainjector -n cert-manager -w --timeout=${TIMEOUT}
    kubectl rollout status deployment/cert-manager -n cert-manager -w --timeout=${TIMEOUT}
}

install_minio() {
    echo "::group::Minio Deployment"
    helm upgrade --install minio minio/minio --version=${MINIO_VERSION} -f config/minio/values.yaml \
        --namespace=minio --create-namespace --timeout=${TIMEOUT} \
        --set accessKey=YOURACCESSKEY --set secretKey=YOURSECRETKEY \
        --set "ingress.hosts[0]=minio.minio.${INGRESS_HOST}" \
        --set "consoleIngress.hosts[0]=console.minio.${INGRESS_HOST}"

    kubectl rollout status statefulset/minio -n minio -w --timeout=${TIMEOUT}
}

install_kserve() {
    echo "::group::Kserve Deployment"
    wget https://raw.githubusercontent.com/kserve/kserve/master/install/v${KSERVE_VERSION}/kserve.yaml -O config/kserve/kserve.yaml
    kubectl apply -k config/kserve
    kubectl rollout status statefulset/kserve-controller-manager -n kserve -w --timeout=${TIMEOUT}
    kubectl apply -f https://raw.githubusercontent.com/kserve/kserve/master/install/v${KSERVE_VERSION}/kserve-runtimes.yaml
}

add_helm_repo
install_istio
install_minio
install_knative
install_vault
install_cert_manager
install_kserve
store_cluster_secret
