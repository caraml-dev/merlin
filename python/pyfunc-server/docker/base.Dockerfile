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

FROM condaforge/miniforge3:22.11.1-4

ARG PYTHON_VERSION

ENV GCLOUD_VERSION=405.0.1
RUN wget -qO- https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${GCLOUD_VERSION}-linux-x86_64.tar.gz | tar xzf -
ENV PATH=$PATH:/google-cloud-sdk/bin

COPY pyfunc-server /pyfunc-server
COPY sdk /sdk
ENV SDK_PATH=/sdk

# Install conda addons for faster search
RUN conda install -n base conda-libmamba-solver && \
    conda config --set solver libmamba

RUN conda env create -f /pyfunc-server/docker/env${PYTHON_VERSION}.yaml && \
    rm -rf /root/.cache

ENV GRPC_HEALTH_PROBE_VERSION=v0.4.4
RUN wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

RUN mkdir /prom_dir
ENV PROMETHEUS_MULTIPROC_DIR=/prom_dir
# For backward compatibility
ENV prometheus_multiproc_dir=/prom_dir
