#!/bin/sh
set -x

export CLUSTER_NAME=dev

export ISTIO_VERSION=1.18.2
export KNATIVE_VERSION=v1.10.2
export KNATIVE_NET_ISTIO_VERSION=v1.10.1
export CERT_MANAGER_VERSION=v1.12.2
export MINIO_VERSION=3.6.3
export KSERVE_VERSION=v0.11.0

export OAUTH_CLIENT_ID=""
export MLP_CHART_VERSION=0.6.1
export MERLIN_CHART_VERSION=0.11.7
export MERLIN_VERSION=0.31.1

# Install Istio
curl --location https://istio.io/downloadIstio | sh -

cat << EOF > ./istio-config.yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
  profile: default
  meshConfig:
    accessLogFile: /dev/stdout
    enableTracing: true
  components:
    egressGateways:
    - name: istio-egressgateway
      enabled: true
  values:
    global:
      proxy:
        autoInject: disabled
    gateways:
      istio-ingressgateway:
        runAsRoot: true
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

## Install Knative
kubectl apply --filename=https://github.com/knative/serving/releases/download/knative-${KNATIVE_VERSION}/serving-crds.yaml
kubectl apply --filename=https://github.com/knative/serving/releases/download/knative-${KNATIVE_VERSION}/serving-core.yaml

export INGRESS_HOST=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
cat <<EOF > ./patch-config-domain.json
{
  "data": {
    "${INGRESS_HOST}.nip.io": ""
  }
}
EOF
kubectl patch configmap/config-domain --namespace=knative-serving --type=merge --patch="$(cat patch-config-domain.json)"

# Install Knative Net Istio
kubectl apply --filename=https://github.com/knative-sandbox/net-istio/releases/download/knative-${KNATIVE_NET_ISTIO_VERSION}/release.yaml

# Install Cert Manager
kubectl apply --filename=https://github.com/jetstack/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.yaml
kubectl wait deployment.apps/cert-manager-webhook --namespace=cert-manager --for=condition=available --timeout=600s
sleep 15

# Install Minio
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

# Install KFServing
kubectl apply --filename=https://github.com/kserve/kserve/releases/download/${KSERVE_VERSION}/kserve.yaml
kubectl apply --filename=https://github.com/kserve/kserve/releases/download/${KSERVE_VERSION}/kserve-runtimes.yaml

cat <<EOF > ./patch-config-inferenceservice.json
{
  "data": {
    "storageInitializer": "{\n    \"image\" : \"ghcr.io/ariefrahmansyah/kfserving-storage-init:latest\",\n    \"memoryRequest\": \"100Mi\",\n    \"memoryLimit\": \"1Gi\",\n    \"cpuRequest\": \"25m\",\n    \"cpuLimit\": \"1\"\n}",
    "logger": "{\n    \"image\" : \"kserve/agent:v0.11.0\",\n    \"memoryRequest\": \"100Mi\",\n    \"memoryLimit\": \"1Gi\",\n    \"cpuRequest\": \"25m\",\n    \"cpuLimit\": \"1\",\n    \"defaultUrl\": \"http:\/\/default-broker\"\n}"
  }
}
EOF
kubectl patch configmap/inferenceservice-config --namespace=kserve --type=merge --patch="$(cat patch-config-inferenceservice.json)"

# Install MLP and Merlin

# create new helm value file containing cluster credentials
cat <<EOF | yq e -P - > k8s_config.yaml
{
  "k8s_config": {
    "name": "dev",
    "cluster": {
      "server": "https://kubernetes.default.svc.cluster.local:443",
      "certificate-authority-data": "$(awk '{printf "%s\n", $0}' ~/.minikube/ca.crt | base64)"
    },
    "user": {
      "client-certificate-data": "$(awk '{printf "%s\n", $0}' ~/.minikube/profiles/minikube/client.crt | base64)",
      "client-key-data": "$(awk '{printf "%s\n", $0}' ~/.minikube/profiles/minikube/client.key | base64)"
    }
  }
}
EOF

output=$(yq '.k8s_config' k8s_config.yaml)
output="$output" yq ".environmentConfigs[0] *= load(\"k8s_config.yaml\") | .imageBuilder.k8sConfig |= env(output)" -i "values-local.yaml"

helm repo add caraml https://caraml-dev.github.io/helm-charts

helm upgrade --install --create-namespace merlin caraml/merlin --namespace=caraml --values=./values-local.yaml \
  --version ${MERLIN_CHART_VERSION} \
  --set deployment.image.tag=${MERLIN_VERSION} \
  --set ui.oauthClientID=${OAUTH_CLIENT_ID} \
  --set config.MlpAPIConfig.APIHost=http://mlp.mlp.${INGRESS_HOST}.nip.io \
  --set ingress.enabled=true \
  --set ingress.class=istio \
  --set ingress.host=merlin.mlp.${INGRESS_HOST}.nip.io \
  --set ingress.path="/*" \
  --set mlflow.ingress.enabled=true \
  --set mlflow.ingress.class=istio \
  --set mlflow.ingress.host=merlin-mlflow.mlp.${INGRESS_HOST}.nip.io \
  --set mlflow.ingress.path="/*" \
  --set mlp.deployment.apiHost=http://mlp.mlp.${INGRESS_HOST}.nip.io/v1 \
  --set mlp.deployment.mlflowTrackingUrl=http://merlin-mlflow.mlp.${INGRESS_HOST}.nip.io \
  --set mlp.ingress.host=mlp.mlp.${INGRESS_HOST}.nip.io \
  --set mlp.ingress.path="/*" \
  --set minio.enabled=false \
  --set kserve.enabled=false \
  --timeout=5m \
  --wait
