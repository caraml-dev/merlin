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

ARG BASE_IMAGE

FROM ${BASE_IMAGE}
# Get the model

ARG MODEL_URL
ARG GOOGLE_APPLICATION_CREDENTIALS

# Run docker build using the provided credential
RUN if [[-z "$GOOGLE_APPLICATION_CREDENTIALS"]]; then gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS}; fi
RUN gsutil -m cp -r ${MODEL_URL} .
# pip 20.2.4 to allow dependency conflicts
RUN /bin/bash -c ". activate ${CONDA_ENVIRONMENT} && \
    sed -i 's/mlflow$/mlflow==1.23.0/' ${HOME}/model/conda.yaml && \
    conda env update --name ${CONDA_ENVIRONMENT} --file ${HOME}/model/conda.yaml && \
    python ${HOME}/merlin-spark-app/main.py --dry-run-model ${HOME}/model"
