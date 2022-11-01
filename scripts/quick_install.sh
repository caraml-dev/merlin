#!/bin/sh

export CLUSTER_NAME=dev

export ISTIO_VERSION=1.5.4
export KNATIVE_VERSION=v0.14.3
export KNATIVE_NET_ISTIO_VERSION=v0.15.0
export CERT_MANAGER_VERSION=v1.1.0
export KFSERVING_VERSION=v0.4.0

export VAULT_VERSION=0.7.0
export MINIO_VERSION=7.0.2

export OAUTH_CLIENT_ID="<put your oauth client id here>"
export MERLIN_VERSION=v0.9.0

# Install Istio
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

# Install Knative
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

# Install Knative Net Istio
kubectl apply --filename=https://github.com/knative/net-istio/releases/download/${KNATIVE_NET_ISTIO_VERSION}/release.yaml

# Install Cert Manager
kubectl apply --filename=https://github.com/jetstack/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.yaml
kubectl wait deployment.apps/cert-manager-webhook --namespace=cert-manager --for=condition=available --timeout=600s
sleep 15

# Install KFServing
kubectl apply --filename=https://raw.githubusercontent.com/kubeflow/kfserving/master/install/${KFSERVING_VERSION}/kfserving.yaml

cat <<EOF > ./patch-config-inferenceservice.json
{
  "data": {
    "storageInitializer": "{\n    \"image\" : \"ghcr.io/ariefrahmansyah/kfserving-storage-init:latest\",\n    \"memoryRequest\": \"100Mi\",\n    \"memoryLimit\": \"1Gi\",\n    \"cpuRequest\": \"25m\",\n    \"cpuLimit\": \"1\"\n}",
    "logger": "{\n    \"image\" : \"gcr.io\/kfserving\/logger:v0.4.0\",\n    \"memoryRequest\": \"100Mi\",\n    \"memoryLimit\": \"1Gi\",\n    \"cpuRequest\": \"25m\",\n    \"cpuLimit\": \"1\",\n    \"defaultUrl\": \"http:\/\/default-broker\"\n}"
  }
}
EOF
kubectl patch configmap/inferenceservice-config --namespace=kfserving-system --type=merge --patch="$(cat patch-config-inferenceservice.json)"

# Install Vault
cat <<EOF > ./vault-values.yaml
injector:
  enabled: false
server:
  dev:
    enabled: true
  dataStorage:
    enabled: false
  resources:
    requests:
      cpu: 25m
      memory: 64Mi
    limits:
      memory: 128Mi
  affinity: null
  tolerations: null
EOF

kubectl create namespace vault
helm repo add hashicorp https://helm.releases.hashicorp.com
helm install vault hashicorp/vault --namespace=vault --version=${VAULT_VERSION} --values=vault-values.yaml --wait --timeout=600s
sleep 5
kubectl wait pod/vault-0 --namespace=vault --for=condition=ready --timeout=600s

# Downgrade to Vault KV secrets engine version 1
kubectl exec vault-0 --namespace=vault -- vault secrets disable secret
kubectl exec vault-0 --namespace=vault -- vault secrets enable -version=1 -path=secret kv

# Write cluster credential to be saved in Vault
cat <<EOF > cluster-credential.json
{
  "name": "dev",
  "master_ip": "https://kubernetes.default.svc:443",
  "certs": "$(cat ~/.minikube/ca.crt | awk '{printf "%s\\n", $0}')",
  "client_certificate": "$(cat ~/.minikube/profiles/minikube/client.crt | awk '{printf "%s\\n", $0}'))",
  "client_key": "$(cat ~/.minikube/profiles/minikube/client.key | awk '{printf "%s\\n", $0}'))"
}
EOF

kubectl cp cluster-credential.json vault/vault-0:/tmp/cluster-credential.json
kubectl exec vault-0 --namespace=vault -- vault kv put secret/${CLUSTER_NAME} @/tmp/cluster-credential.json

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

# Install MLP
kubectl create namespace mlp

git clone git@github.com:gojek/mlp.git

helm install mlp ./mlp/chart --namespace=mlp --values=./mlp/chart/values-e2e.yaml \
  --set mlp.image.tag=main \
  --set mlp.apiHost=http://mlp.mlp.${INGRESS_HOST}.nip.io/v1 \
  --set mlp.oauthClientID=${OAUTH_CLIENT_ID} \
  --set mlp.mlflowTrackingUrl=http://mlflow.mlp.${INGRESS_HOST}.nip.io \
  --set mlp.ingress.enabled=true \
  --set mlp.ingress.class=istio \
  --set mlp.ingress.host=mlp.mlp.${INGRESS_HOST}.nip.io \
  --wait --timeout=5m

cat <<EOF > mlp-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: mlp
  namespace: mlp
  annotations:
    kubernetes.io/ingress.class: istio
  labels:
    app: mlp
spec:
  rules:
    - host: 'mlp.mlp.${INGRESS_HOST}.nip.io'
      http:
        paths:
          - path: /*
            pathType: Prefix
            backend:
              service:
                name: mlp
                port:
                  number: 8080
EOF

# Install Merlin
kubectl create secret generic vault-secret --namespace=mlp --from-literal=address=http://vault.vault.svc.cluster.local --from-literal=token=root

helm install merlin ../charts/merlin --namespace=mlp --values=../charts/merlin/values-e2e.yaml \
  --set merlin.image.tag=${MERLIN_VERSION} \
  --set merlin.oauthClientID=${OAUTH_CLIENT_ID} \
  --set merlin.apiHost=http://merlin.mlp.${INGRESS_HOST}.nip.io/v1 \
  --set merlin.mlpApi.apiHost=http://mlp.mlp.${INGRESS_HOST}.nip.io/v1 \
  --set merlin.ingress.enabled=true \
  --set merlin.ingress.class=istio \
  --set merlin.ingress.host=merlin.mlp.${INGRESS_HOST}.nip.io \
  --set merlin.ingress.path="/*" \
  --set mlflow.ingress.enabled=true \
  --set mlflow.ingress.class=istio \
  --set mlflow.ingress.host=mlflow.mlp.${INGRESS_HOST}.nip.io \
  --set mlflow.ingress.path="/*" \
  --timeout=5m \
  --wait
