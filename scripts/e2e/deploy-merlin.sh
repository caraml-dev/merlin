#!/bin/bash

set -ex

CHART_PATH="$1"
export INGRESS_HOST=127.0.0.1
export MERLIN_VERSION=${GITHUB_HEAD_REF:-${GITHUB_REF#refs/*/}}


helm install --debug merlin ${CHART_PATH} --namespace=mlp --values=${CHART_PATH}/values-e2e.yaml \
  --set merlin.image.tag=${MERLIN_VERSION} \
  --set merlin.apiHost=http://merlin.mlp.${INGRESS_HOST}.nip.io/v1 \
  --set merlin.mlpApi.apiHost=http://mlp.mlp.svc.cluster.local:8080/v1 \
  --set merlin.ingress.enabled=true \
  --set merlin.ingress.class=istio \
  --set merlin.ingress.host=merlin.mlp.${INGRESS_HOST}.nip.io \
  --set merlin.ingress.path="/*" \
  --set mlflow.ingress.enabled=true \
  --set mlflow.ingress.class=istio \
  --set mlflow.ingress.host=merlin-mlflow.mlp.${INGRESS_HOST}.nip.io \
  --set mlflow.extraEnvs.MLFLOW_S3_ENDPOINT_URL=http://minio.minio.svc.cluster.local:9000 \
  --set mlflow.ingress.path="/*" \
  --set mlflow.postgresql.requests.cpu="25m" \
  --set mlflow.postgresql.requests.memory="256Mi" \
  --timeout=5m \
  --dry-run

helm install --debug merlin ${CHART_PATH} --namespace=mlp --values=${CHART_PATH}/values-e2e.yaml \
  --set merlin.image.tag=${MERLIN_VERSION} \
  --set merlin.apiHost=http://merlin.mlp.${INGRESS_HOST}.nip.io/v1 \
  --set merlin.mlpApi.apiHost=http://mlp.mlp.svc.cluster.local:8080/v1 \
  --set merlin.ingress.enabled=true \
  --set merlin.ingress.class=istio \
  --set merlin.ingress.host=merlin.mlp.${INGRESS_HOST}.nip.io \
  --set merlin.ingress.path="/*" \
  --set mlflow.ingress.enabled=true \
  --set mlflow.ingress.class=istio \
  --set mlflow.ingress.host=merlin-mlflow.mlp.${INGRESS_HOST}.nip.io \
  --set mlflow.extraEnvs.MLFLOW_S3_ENDPOINT_URL=http://minio.minio.svc.cluster.local:9000 \
  --set mlflow.ingress.path="/*" \
  --set mlflow.resources.requests.cpu="25m" \
  --set mlflow.resources.requests.memory="256Mi" \
  --timeout=5m \
  --wait

cat <<EOF > ./patch-merlin-mlflow-nodeport.yaml
spec:
  type: NodePort
  ports:
  - name: http2
    nodePort: 31100
    port: 80
    protocol: TCP
    targetPort: 5000
EOF
kubectl patch service/merlin-mlflow -n mlp --patch="$(cat patch-merlin-mlflow-nodeport.yaml)"
sleep 12

cat <<EOF > ./logger-sample.yaml
apiVersion: serving.knative.dev/v1alpha1
kind: Service
metadata:
  name: message-dumper
  namespace: mlp
spec:
  template:
    spec:
      containers:
      - image: gcr.io/knative-releases/knative.dev/eventing-contrib/cmd/event_display
EOF
kubectl apply -f logger-sample.yaml
sleep 12

set +ex
