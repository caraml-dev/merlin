# Local Development

In this guide, we will deploy Merlin on a local Minikube cluster.

If you already have existing development cluster, you can run [`quick_install.sh`](../../../scripts/quick_install.sh) to install Merlin and it's components.

## Prerequesites

1. Kubernetes v1.16.15
2. Minikube v1.16.0 with LoadBalancer enabled
3. Istio v1.5.4
4. Knative v0.14.3
5. Cert Manager v1.1.0
6. KFServing v0.4.0
7. Vault v0.7.0, with secret engine v1
8. Minio v7.0.2

## Provision Minikube cluster

First, you need to have Minikube installed on your machine. To install it, please follow this [documentation](https://minikube.sigs.k8s.io/docs/start/). You also need to have a [driver](https://minikube.sigs.k8s.io/docs/drivers/) to run Minikube cluster. This guide uses VirtualBox driver.

Next, create a new Minikube cluster with Kubernetes v1.16.15:

```bash
export CLUSTER_NAME=dev
minikube start --cpus=4 --memory=8192 --kubernetes-version=v1.16.15 --driver=virtualbox
```

Lastly, we need to enable Minikube's LoadBalancer services by running `minikube tunnel` in another terminal.

## Install Istio

We recommend installing Istio without service mesh (sidecar injection disabled). We also need to enable Istio Kubernetes Ingress enabled so we can access Merlin API and UI.

```bash
export ISTIO_VERSION=1.5.4

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
```

## Install Knative

In this step, we install Knative Serving and configure it to use Istio as ingress controller.

```bash
export KNATIVE_VERSION=v0.14.3
export KNATIVE_NET_ISTIO_VERSION=v0.15.0

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
```

## Install Cert Manager

```bash
export CERT_MANAGER_VERSION=v1.1.0

kubectl apply --filename=https://github.com/jetstack/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.yaml
```

## Install KFServing

KFServing manages the deployment of Merlin models.

```bash
export KFSERVING_VERSION=v0.4.0

kubectl apply --filename=https://raw.githubusercontent.com/kubeflow/kfserving/master/install/${KFSERVING_VERSION}/kfserving.yaml

cat <<EOF > ./patch-config-inferenceservice.json
{
  "data": {
    "storageInitializer": "{\n    \"image\" : \"ghcr.io/ariefrahmansyah/kfserving-storage-init:latest\",\n    \"memoryRequest\": \"100Mi\",\n    \"memoryLimit\": \"1Gi\",\n    \"cpuRequest\": \"100m\",\n    \"cpuLimit\": \"1\"\n}",
  }
}
EOF
kubectl patch configmap/inferenceservice-config --namespace=kfserving-system --type=merge --patch="$(cat patch-config-inferenceservice.json)"
```

> Notes that we change KFServing's Storage Initializer image here so it can download the model artifacts from Minio.

## Install Vault

Vault is needed to store the model cluster credential where models will be deployed. For local development, we will use the same Minikube cluster as model cluster. In production, you may have multiple model clusters.

```bash
export VAULT_VERSION=0.7.0

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
```

## Install Minio

Minio is used by MLflow to store model artifacts.

```bash
export MINIO_VERSION=7.0.2

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
```

## Install MLP

MLP and Merlin use Google Sign-in to authenticate the user to access the API and UI. Please follow [this documentation](https://developers.google.com/identity/protocols/oauth2/javascript-implicit-flow) to create Google Authorization credential. You must specify Javascript origins and redirect URIs with both `http://mlp.mlp.${INGRESS_HOST}.nip.io` and `http://merlin.mlp.${INGRESS_HOST}.nip.io`. After you get the client ID, specify it into `OAUTH_CLIENT_ID`.

```bash
export OAUTH_CLIENT_ID="<put your oauth client id here>"

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
  --set mlp.ingress.path="/*" \
  --wait --timeout=5m

cat <<EOF > mlp-ingress.yaml
apiVersion: networking.k8s.io/v1beta1
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
            backend:
              serviceName: mlp
              servicePort: 8080
EOF
```

## Install Merlin

```bash
export MERLIN_VERSION=v0.9.0

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
```

### Check Merlin installation

```bash
kubectl get po -n mlp
NAME                             READY   STATUS    RESTARTS   AGE
merlin-64c9c75dfc-djs4t          1/1     Running   0          12m
merlin-mlflow-5c7dd6d9df-g2s6v   1/1     Running   0          12m
merlin-postgresql-0              1/1     Running   0          12m
mlp-6877d8567-msqg9              1/1     Running   0          15m
mlp-postgresql-0                 1/1     Running   0          15m
```

Once everything is Running, you can open Merlin in <http://merlin.mlp.${INGRESS_HOST}.nip.io/merlin>. From here, you can run Jupyter notebook examples by setting `merlin.set_url("merlin.mlp.${INGRESS_HOST}.nip.io")`.
