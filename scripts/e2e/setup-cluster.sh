#!/usr/bin/env bash
# Bash3 Boilerplate. Copyright (c) 2014, kvz.io

set -o errexit
set -o pipefail
set -o nounset

# Software requirements:
# - yq
# - helm
# - k3d
# - kubectl

# Prerequisites:
# - cluster have been created using k3d

CLUSTER_NAME=$1
INGRESS_HOST=$2

ISTIO_VERSION=1.18.2
# queue proxy version in config/knative/overlay.yaml should also be modified to upgrade knative
KNATIVE_VERSION=1.10.2
KNATIVE_NET_ISTIO_VERSION=1.10.1
CERT_MANAGER_VERSION=1.12.2
MINIO_VERSION=5.0.7
KSERVE_VERSION=0.9.0
TIMEOUT=180s


add_helm_repo() {
    helm repo add istio https://istio-release.storage.googleapis.com/charts
    helm repo add minio https://operator.min.io/
    helm repo update
}

store_cluster_secret() {
  echo "::group::Storing Cluster Secret"

  # create and store K8sConfig in tmp dir
  cat <<EOF |  yq -P - > /tmp/temp_k8sconfig.yaml
{
    "k8s_config": {
        "name": $(k3d kubeconfig get "$CLUSTER_NAME" | yq '.clusters[0].name' -o json -),
        "cluster": $(k3d kubeconfig get "$CLUSTER_NAME" | yq '.clusters[0].cluster | .server = "https://kubernetes.default.svc.cluster.local:443"' -o json -),
        "user": $(k3d kubeconfig get "$CLUSTER_NAME" | yq '.users[0].user' -o json - )
    }
}
EOF
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

    kubectl apply -f config/istio/ingress-class.yaml
}

install_knative() {
    echo "::group::Knative Deployment"
    # Install CRD
    kubectl apply -f https://github.com/knative/serving/releases/download/knative-v${KNATIVE_VERSION}/serving-crds.yaml

    # Install knative serving
    wget https://github.com/knative/serving/releases/download/knative-v${KNATIVE_VERSION}/serving-core.yaml -O config/knative/serving-core.yaml
    kubectl apply -k config/knative

    kubectl rollout status deployment/autoscaler -n knative-serving -w --timeout=${TIMEOUT}
    kubectl rollout status deployment/controller -n knative-serving -w --timeout=${TIMEOUT}
    kubectl rollout status deployment/activator -n knative-serving -w --timeout=${TIMEOUT}
    kubectl rollout status deployment/domain-mapping -n knative-serving -w --timeout=${TIMEOUT}
    kubectl rollout status deployment/domainmapping-webhook -n knative-serving -w --timeout=${TIMEOUT}
    kubectl rollout status deployment/webhook -n knative-serving -w --timeout=${TIMEOUT}

    # Install knative-istio
    kubectl apply -f https://github.com/knative-sandbox/net-istio/releases/download/knative-v${KNATIVE_NET_ISTIO_VERSION}/net-istio.yaml

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
    helm install --namespace minio-operator --create-namespace minio-operator minio/operator --values=config/minio/minio-operator-values.yaml --wait --timeout=600s --version ${MINIO_VERSION}
    helm install --namespace minio-tenant --create-namespace minio-tenant minio/tenant --wait --timeout=600s --version ${MINIO_VERSION} \
        --values=config/minio/minio-tenant-values.yaml \
        --set "ingress.api.host=minio.minio.${INGRESS_HOST}" \
        --set "ingress.console.host=console.minio.minio.${INGRESS_HOST}"
}

install_kserve() {
    echo "::group::KServe Deployment"
    wget https://raw.githubusercontent.com/kserve/kserve/master/install/v${KSERVE_VERSION}/kserve.yaml -O config/kserve/kserve.yaml
    kubectl apply -k config/kserve
    kubectl rollout status deployment/kserve-controller-manager -n kserve -w --timeout=${TIMEOUT}
    kubectl apply -f https://raw.githubusercontent.com/kserve/kserve/master/install/v${KSERVE_VERSION}/kserve-runtimes.yaml
}

patch_coredns() {
    echo "::group::Patching CoreDNS"
    kubectl patch cm coredns -n kube-system --patch-file config/coredns/patch.yaml
    kubectl get cm coredns -n kube-system -o yaml
}

add_helm_repo
install_istio
install_knative
install_minio
install_cert_manager
install_kserve
store_cluster_secret
patch_coredns
