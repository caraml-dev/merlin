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
# queue proxy version in config/knative/overlay.yaml should also be modified to upgrade knative
KNATIVE_VERSION=1.7.4
KNATIVE_NET_ISTIO_VERSION=1.7.1
CERT_MANAGER_VERSION=1.7.2
MINIO_VERSION=3.6.3
KSERVE_VERSION=0.8.0
# VAULT_VERSION=0.19.0
TIMEOUT=180s

add_helm_repo() {
	helm repo add hashicorp https://helm.releases.hashicorp.com
	helm repo add minio https://charts.min.io/
	helm repo add istio https://istio-release.storage.googleapis.com/charts
	helm repo update
}

store_cluster_secret() {
	echo "::group::Storing Cluster Secret"

	k3d kubeconfig get "$CLUSTER_NAME" >/tmp/temp_kubeconfig.yaml
	cat <<EOF >/tmp/temp_k8sconfig.json
{
    "k8s_config": {
        "name": $(yq .clusters[0].name -o json /tmp/temp_kubeconfig.yaml),
        "cluster": $(yq .clusters[0].cluster -o json /tmp/temp_kubeconfig.yaml),
        "user": $(yq .users[0].user -o json /tmp/temp_kubeconfig.yaml)
    }
}
EOF
	# get yaml file
	yq e -P /tmp/temp_k8sconfig.json -o yaml >/tmp/temp_k8sconfig.yaml

	# use yaml file
	# yq '.[0] *= load("/tmp/temp_k8sconfig.yaml")' rest_config.yaml

	# NOTE: Do no delete files created here as they will be used in the next step
	# TODO: Figure out a better way to pass cluster K8sConfig
	# rm /tmp/temp_kubeconfig.yaml
	# rm /tmp/temp_k8sconfig.yaml
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
    kubectl apply -f https://github.com/knative/net-istio/releases/download/knative-v${KNATIVE_NET_ISTIO_VERSION}/net-istio.yaml

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
	echo "::group::KServe Deployment"
	wget https://raw.githubusercontent.com/kserve/kserve/master/install/v${KSERVE_VERSION}/kserve.yaml -O config/kserve/kserve.yaml
	kubectl apply -k config/kserve
	kubectl rollout status statefulset/kserve-controller-manager -n kserve -w --timeout=${TIMEOUT}
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
install_vault
install_cert_manager
install_kserve
store_cluster_secret
patch_coredns
