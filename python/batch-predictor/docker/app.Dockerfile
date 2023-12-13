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

ARG GOOGLE_APPLICATION_CREDENTIALS
RUN if [ ! -z "${GOOGLE_APPLICATION_CREDENTIALS}" ]; \
    then gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS} > gcloud-auth.txt; \
    else echo "GOOGLE_APPLICATION_CREDENTIALS is not set. Skipping gcloud auth command." > gcloud-auth.txt; \
    fi

# Download and install user model dependencies
ARG MODEL_DEPENDENCIES_URL
RUN gsutil cp ${MODEL_DEPENDENCIES_URL} conda.yaml
RUN conda env create --name merlin-model --file conda.yaml

# Copy and install batch predictor dependencies
COPY --chown=${UID}:${GID} batch-predictor ${HOME}/merlin-spark-app
COPY --chown=${UID}:${GID} sdk ${HOME}/sdk
ENV SDK_PATH=${HOME}/sdk

RUN /bin/bash -c ". activate merlin-model && pip install -r ${HOME}/merlin-spark-app/requirements.txt"

# Download and dry-run user model artifacts and code
ARG MODEL_ARTIFACTS_URL
RUN gsutil -m cp -r ${MODEL_ARTIFACTS_URL} .
RUN pwd
RUN ls -lah
RUN ls -lah ${HOME}
RUN ls -lah ${HOME}/merlin-spark-app
RUN ls -lah ${HOME}/model
RUN /bin/bash -c ". activate merlin-model && python ${HOME}/merlin-spark-app/main.py --dry-run-model ${HOME}/model"

# Copy batch predictor application entrypoint to protected directory
COPY batch-predictor/merlin-entrypoint.sh /opt/merlin-entrypoint.sh

ENTRYPOINT [ "/opt/merlin-entrypoint.sh" ]
