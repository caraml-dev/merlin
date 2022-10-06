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

# base dockerfile using python 3.7
FROM continuumio/miniconda3

RUN wget -qO- https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-367.0.0-linux-x86_64.tar.gz | tar xzf -
ENV PATH=$PATH:/google-cloud-sdk/bin

COPY pyfunc-server /pyfunc-server
COPY sdk /sdk
ENV SDK_PATH=$PATH:/sdk
RUN conda env create -f /pyfunc-server/docker/env37.yaml && \
    rm -rf /root/.cache

RUN mkdir /prom_dir
ENV PROMETHEUS_MULTIPROC_DIR=/prom_dir
# For backward compatibility
ENV prometheus_multiproc_dir=/prom_dir
