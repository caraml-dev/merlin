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

FROM condaforge/miniforge3:4.14.0-0

ARG PYTHON_VERSION

ENV GCLOUD_VERSION=405.0.1
RUN wget -qO- https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${GCLOUD_VERSION}-linux-x86_64.tar.gz | tar xzf -
ENV PATH=$PATH:/google-cloud-sdk/bin

COPY pyfunc-server /pyfunc-server
COPY sdk /sdk
ENV SDK_PATH=/sdk
RUN conda env create -f /pyfunc-server/docker/env${PYTHON_VERSION}.yaml && \
    rm -rf /root/.cache

RUN mkdir /prom_dir
ENV PROMETHEUS_MULTIPROC_DIR=/prom_dir
# For backward compatibility
ENV prometheus_multiproc_dir=/prom_dir
