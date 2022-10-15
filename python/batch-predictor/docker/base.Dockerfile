# Copyright 2020 The Merlin Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ARG SPARK_VERSION=v3.0.0
FROM gcr.io/spark-operator/spark-py:${SPARK_VERSION}

# Switch to user root so we can add additional jars and configuration files.
USER root

ENV LANG=C.UTF-8 LC_ALL=C.UTF-8
ENV PATH /opt/conda/bin:$PATH
ARG PYTHON_VERSION
ARG SPARK_OPERATOR_VERSION=v1beta2-1.2.2-3.0.0
ARG SPARK_BQ_CONNECTOR_VERSION=0.19.1

# Setup dependencies for Google Cloud Storage access.
RUN rm $SPARK_HOME/jars/guava-14.0.1.jar
ADD https://repo1.maven.org/maven2/com/google/guava/guava/23.0/guava-23.0.jar \
    $SPARK_HOME/jars

# Add the connector jar needed to access Google Cloud Storage using the Hadoop FileSystem API.
ADD https://storage.googleapis.com/hadoop-lib/gcs/gcs-connector-hadoop2-2.0.1.jar \
    $SPARK_HOME/jars
ADD https://repo1.maven.org/maven2/com/google/cloud/spark/spark-bigquery-with-dependencies_2.12/${SPARK_BQ_CONNECTOR_VERSION}/spark-bigquery-with-dependencies_2.12-${SPARK_BQ_CONNECTOR_VERSION}.jar \
    $SPARK_HOME/jars
RUN chmod 644 -R $SPARK_HOME/jars/*

# Setup for the Prometheus JMX exporter.
RUN mkdir -p /etc/metrics/conf
# Add the Prometheus JMX exporter Java agent jar for exposing metrics sent to the JmxSink to Prometheus.
ADD https://repo1.maven.org/maven2/io/prometheus/jmx/jmx_prometheus_javaagent/0.11.0/jmx_prometheus_javaagent-0.11.0.jar /prometheus/
RUN chmod 644 /prometheus/jmx_prometheus_javaagent-0.11.0.jar

ADD https://raw.githubusercontent.com/GoogleCloudPlatform/spark-on-k8s-operator/$SPARK_OPERATOR_VERSION/spark-docker/conf/metrics.properties /etc/metrics/conf
ADD https://raw.githubusercontent.com/GoogleCloudPlatform/spark-on-k8s-operator/$SPARK_OPERATOR_VERSION/spark-docker/conf/prometheus.yaml /etc/metrics/conf
RUN chmod 644 -R /etc/metrics/conf/*

# Configure non-root user
ARG username=spark
ARG spark_uid=185
ARG spark_gid=100
ENV USER $username
ENV UID $spark_uid
ENV GID $spark_gid
ENV HOME /home/$USER
RUN adduser --disabled-password --uid $UID --gid $GID --home $HOME $USER

# Switch to Spark user
USER ${USER}
WORKDIR $HOME

RUN apt-get update --fix-missing --allow-releaseinfo-change && apt-get install -y wget bzip2 ca-certificates \
    libglib2.0-0 libxext6 libsm6 libxrender1 \
    git mercurial subversion

# Install miniconda
RUN wget --quiet https://github.com/conda-forge/miniforge/releases/download/4.14.0-0/Miniforge3-4.14.0-0-Linux-x86_64.sh -O ~/miniconda.sh && \
    /bin/bash ~/miniconda.sh -b -p /opt/conda && \
    rm ~/miniconda.sh && \
    /opt/conda/bin/conda clean -all -y && \
    ln -s /opt/conda/etc/profile.d/conda.sh /etc/profile.d/conda.sh && \
    echo ". /opt/conda/etc/profile.d/conda.sh" >> ~/.bashrc && \
    echo "conda activate base" >> ~/.bashrc

# Install gcloud SDK
WORKDIR /
ARG GCLOUD_VERSION=332.0.0
RUN wget -qO- https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${GCLOUD_VERSION}-linux-x86_64.tar.gz | tar xzf -
ENV PATH=$PATH:/google-cloud-sdk/bin

# Copy batch predictor application
COPY batch-predictor /merlin-spark-app
COPY sdk /sdk
COPY batch-predictor/merlin-entrypoint.sh /opt/merlin-entrypoint.sh

# Setup base conda environment
RUN conda env create -f /merlin-spark-app/env${PYTHON_VERSION}.yaml && \
    rm -rf /root/.cache

ENTRYPOINT [ "/opt/merlin-entrypoint.sh" ]
