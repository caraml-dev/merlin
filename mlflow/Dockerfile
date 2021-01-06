FROM python:3.7.2-slim

RUN pip install google-cloud-storage
RUN pip install psycopg2-binary

ARG BOTO3_VERSION=1.7.12
ARG MLFLOW_VERSION=1.3.0

RUN pip install boto3==${BOTO3_VERSION}
RUN pip install mlflow==${MLFLOW_VERSION}

ENV BACKEND_STORE_URI="/data/mlruns"
ENV ARTIFACT_ROOT="/data/artifacts"
ENV HOST="0.0.0.0"
ENV PORT="5000"

CMD "mlflow" "server" "--backend-store-uri=${BACKEND_STORE_URI}" "--default-artifact-root=${ARTIFACT_ROOT}" "--host=${HOST}" "--port=${PORT}"
