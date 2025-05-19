FROM python:3.9-slim

RUN apt-get update && \
    apt-get install -y \
        build-essential \
        make \
        gcc \
        locales && \
    apt-get clean

RUN pip install google-cloud-storage
RUN pip install psycopg2-binary==2.9.10

ARG BOTO3_VERSION=1.35.39
ARG MLFLOW_VERSION=1.26.1
ARG SQLALCHEMY_VERSION=1.4.54

RUN pip install boto3==${BOTO3_VERSION}
RUN pip install SQLAlchemy==${SQLALCHEMY_VERSION}
RUN pip install mlflow==${MLFLOW_VERSION}

ENV BACKEND_STORE_URI="/data/mlruns"
ENV ARTIFACT_ROOT="/data/artifacts"
ENV HOST="0.0.0.0"
ENV PORT="5000"

CMD "mlflow" "server" "--backend-store-uri=${BACKEND_STORE_URI}" "--default-artifact-root=${ARTIFACT_ROOT}" "--host=${HOST}" "--port=${PORT}"
