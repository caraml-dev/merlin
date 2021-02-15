#!/bin/bash

set -ex

export CLUSTER_NAME=dev
export KIND_NODE_VERSION=v1.16.15
export VAULT_VERSION=0.7.0
export ISTIO_VERSION=1.5.4
export KNATIVE_VERSION=v0.14.3
export KNATIVE_NET_ISTIO_VERSION=v0.15.0
export CERT_MANAGER_VERSION=v1.1.0
export KFSERVING_VERSION=v0.4.0

export VAULT_VERSION=0.7.0
export MINIO_VERSION=7.0.2

export OAUTH_CLIENT_ID="<put your oauth client id here>"
export MERLIN_VERSION=0.10

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
# Install Istio
#
curl --location https://git.io/getLatestIstio | sh -
cat << EOF > ./istio-config.yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
  values:
    global:
      proxy:
        autoInject: disabled
      useMCP: false
      jwtPolicy: first-party-jwt
      k8sIngress:
        enabled: true
  addonComponents:
    pilot:
      enabled: true
    prometheus:
      enabled: false
  components:
    ingressGateways:
      - name: istio-ingressgateway
        enabled: true
        k8s:
          resources:
            requests:
              cpu: 20m
              memory: 64Mi
            limits:
              memory: 128Mi
      - name: cluster-local-gateway
        enabled: true
        label:
          istio: cluster-local-gateway
          app: cluster-local-gateway
        k8s:
          resources:
            requests:
              cpu: 20m
              memory: 64Mi
            limits:
              memory: 128Mi
          service:
            type: ClusterIP
            ports:
              - port: 15020
                name: status-port
              - port: 80
                name: http2
              - port: 443
                name: https
EOF
istio-${ISTIO_VERSION}/bin/istioctl manifest apply -f istio-config.yaml

########################################
# Install Knative
#
kubectl apply --filename=https://github.com/knative/serving/releases/download/${KNATIVE_VERSION}/serving-crds.yaml
kubectl apply --filename=https://github.com/knative/serving/releases/download/${KNATIVE_VERSION}/serving-core.yaml

export INGRESS_HOST=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
cat <<EOF > ./patch-config-domain.json
{
  "data": {
    "${INGRESS_HOST}.nip.io": ""
  }
}
EOF
kubectl patch configmap/config-domain --namespace=knative-serving --type=merge --patch="$(cat patch-config-domain.json)"

########################################
# Install Knative Net Istio
#
kubectl apply --filename=https://github.com/knative/net-istio/releases/download/${KNATIVE_NET_ISTIO_VERSION}/release.yaml

########################################
# Install Cert Manager
#
kubectl apply --filename=https://github.com/jetstack/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.yaml
kubectl wait deployment.apps/cert-manager-webhook --namespace=cert-manager --for=condition=available --timeout=600s
sleep 15

########################################
# Install KFServing
#
kubectl apply --filename=https://raw.githubusercontent.com/kubeflow/kfserving/master/install/${KFSERVING_VERSION}/kfserving.yaml

cat <<EOF > ./patch-config-inferenceservice.json
{
  "data": {
    "storageInitializer": "{\n\"image\":\"ghcr.io/ariefrahmansyah/kfserving-storage-init:latest\",\n\"memoryRequest\":\"100Mi\",\n\"memoryLimit\":\"1Gi\",\n\"cpuRequest\":\"100m\",\n\"cpuLimit\":\"1\"\n}"
  }
}
EOF
kubectl patch configmap/inferenceservice-config --namespace=kfserving-system --type=merge --patch="$(cat patch-config-inferenceservice.json)"


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

kubectl get pods -A
kubectl top nodes
kubectl describe pod vault-0 --namespace=vault
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

########################################
# Install Minio
#
cat <<EOF > minio-values.yaml
replicas: 1
persistence:
  enabled: false
resources:
  requests:
    cpu: 25m
    memory: 64Mi
livenessProbe:
  initialDelaySeconds: 30
defaultBucket:
  enabled: true
  name: mlflow
ingress:
  enabled: true
  annotations:
    kubernetes.io/ingress.class: istio
  path: /*
  hosts:
    - 'minio.minio.${INGRESS_HOST}.nip.io'
EOF

kubectl create namespace minio
helm repo add minio https://helm.min.io/
helm install minio minio/minio --version=${MINIO_VERSION} --namespace=minio --values=minio-values.yaml --wait --timeout=600s


set +ex
