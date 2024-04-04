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

FROM apache/spark-py:v3.1.3

# Switch to user root so we can add additional jars and configuration files.
USER root

ENV LANG=C.UTF-8 LC_ALL=C.UTF-8
ENV PATH /opt/conda/bin:$PATH

ENV SPARK_OPERATOR_VERSION=v1beta2-1.3.7-3.1.1
ENV SPARK_BQ_CONNECTOR_VERSION=0.27.0

# Setup dependencies for Google Cloud Storage access.
RUN rm $SPARK_HOME/jars/guava-14.0.1.jar
ADD https://repo1.maven.org/maven2/com/google/guava/guava/23.0/guava-23.0.jar \
    $SPARK_HOME/jars

# Add the connector jar needed to access Google Cloud Storage using the Hadoop FileSystem API.
ADD https://storage.googleapis.com/hadoop-lib/gcs/gcs-connector-hadoop2-2.2.8.jar \
    $SPARK_HOME/jars
ADD https://repo1.maven.org/maven2/com/google/cloud/spark/spark-bigquery-with-dependencies_2.12/${SPARK_BQ_CONNECTOR_VERSION}/spark-bigquery-with-dependencies_2.12-${SPARK_BQ_CONNECTOR_VERSION}.jar \
    $SPARK_HOME/jars
RUN chmod 644 -R $SPARK_HOME/jars/*

# Setup for the Prometheus JMX exporter.
RUN mkdir -p /etc/metrics/conf
# Add the Prometheus JMX exporter Java agent jar for exposing metrics sent to the JmxSink to Prometheus.
ADD https://repo1.maven.org/maven2/io/prometheus/jmx/jmx_prometheus_javaagent/0.11.0/jmx_prometheus_javaagent-0.11.0.jar /prometheus/
RUN chmod 644 /prometheus/jmx_prometheus_javaagent-0.11.0.jar

ADD https://raw.githubusercontent.com/GoogleCloudPlatform/spark-on-k8s-operator/${SPARK_OPERATOR_VERSION}/spark-docker/conf/metrics.properties /etc/metrics/conf
ADD https://raw.githubusercontent.com/GoogleCloudPlatform/spark-on-k8s-operator/${SPARK_OPERATOR_VERSION}/spark-docker/conf/prometheus.yaml /etc/metrics/conf
RUN chmod 644 -R /etc/metrics/conf/*

RUN apt-get update --fix-missing --allow-releaseinfo-change && apt-get install -y wget bzip2 ca-certificates \
    libglib2.0-0 libxext6 libsm6 libxrender1 \
    git mercurial subversion

# Install yq
ENV YQ_VERSION=v4.42.1
RUN wget https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_amd64 -O /usr/bin/yq && \
    chmod +x /usr/bin/yq

# Install gcloud SDK
ENV PATH=$PATH:/google-cloud-sdk/bin
ENV GCLOUD_VERSION=405.0.1
RUN wget -qO- https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${GCLOUD_VERSION}-linux-x86_64.tar.gz | tar xzf - -C /

# Configure non-root user
ENV USER spark
ENV UID 185
ENV GID 100
ENV HOME /home/$USER
RUN adduser --disabled-password --uid $UID --gid $GID --home $HOME $USER

# Switch to Spark user
USER ${USER}
WORKDIR ${HOME}

# Install miniconda
ENV CONDA_DIR ${HOME}/miniconda3
ENV PATH ${CONDA_DIR}/bin:$PATH
ENV MINIFORGE_VERSION=23.11.0-0

RUN wget --quiet https://github.com/conda-forge/miniforge/releases/download/${MINIFORGE_VERSION}/Miniforge3-${MINIFORGE_VERSION}-Linux-x86_64.sh -O miniconda.sh && \
    /bin/bash miniconda.sh -b -p ${CONDA_DIR} && \
    rm ~/miniconda.sh && \
    $CONDA_DIR/bin/conda clean -afy && \
    echo "source $CONDA_DIR/etc/profile.d/conda.sh" >> $HOME/.bashrc

COPY batch-predictor/docker/main.py ${HOME}/merlin-spark-app/main.py
COPY batch-predictor/docker/merlin_entrypoint.sh /usr/bin/merlin_entrypoint.sh
COPY batch-predictor/docker/process_conda_env.sh /usr/bin/process_conda_env.sh
