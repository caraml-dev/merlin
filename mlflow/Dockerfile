FROM python:3.7.2-slim

RUN apt-get update && \
    apt-get install -y \
        build-essential \
        make \
        gcc \
        locales \
        libgdal20 libgdal-dev && \
    apt-get clean

RUN pip install google-cloud-storage
RUN pip install psycopg2-binary==2.8.6

ARG BOTO3_VERSION=1.7.12
ARG MLFLOW_VERSION=1.3.0

RUN pip install boto3==${BOTO3_VERSION}
RUN pip install mlflow==${MLFLOW_VERSION}

ENV BACKEND_STORE_URI="/data/mlruns"
ENV ARTIFACT_ROOT="/data/artifacts"
ENV HOST="0.0.0.0"
ENV PORT="5000"

CMD "mlflow" "server" "--backend-store-uri=${BACKEND_STORE_URI}" "--default-artifact-root=${ARTIFACT_ROOT}" "--host=${HOST}" "--port=${PORT}"
