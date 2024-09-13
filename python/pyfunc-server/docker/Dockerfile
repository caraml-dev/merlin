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

ARG MLFLOW_ARTIFACT_STORAGE_TYPE

ARG GOOGLE_APPLICATION_CREDENTIALS

ARG AWS_ACCESS_KEY_ID
ARG AWS_SECRET_ACCESS_KEY
ARG AWS_DEFAULT_REGION
ARG AWS_ENDPOINT_URL

RUN if [ "${MLFLOW_ARTIFACT_STORAGE_TYPE}" = "gcs" ]; then  \
        if [ ! -z "${GOOGLE_APPLICATION_CREDENTIALS}" ]; then \
            gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS} > gcloud-auth.txt; \
        else echo "GOOGLE_APPLICATION_CREDENTIALS is not set. Skipping gcloud auth command." > gcloud-auth.txt; \
        fi \
    elif [ "${MLFLOW_ARTIFACT_STORAGE_TYPE}" = "s3" ]; then \
       echo "S3 credentials used"; \
    else \
       echo "No credentials are used"; \
    fi

# Download and install user model dependencies
ARG MODEL_DEPENDENCIES_URL
RUN if [ "${MLFLOW_ARTIFACT_STORAGE_TYPE}" = "gcs" ]; then  \
        gsutil cp ${MODEL_DEPENDENCIES_URL} conda.yaml; \
    elif [ "${MLFLOW_ARTIFACT_STORAGE_TYPE}" = "s3" ]; then \
        S3_KEY=${MODEL_DEPENDENCIES_URL##*s3://}; \
        aws s3api get-object --bucket ${S3_KEY%%/*} --key ${S3_KEY#*/} conda.yaml; \
    else \
        echo "No credentials are used"; \
    fi

ARG MERLIN_DEP_CONSTRAINT
RUN process_conda_env.sh conda.yaml "merlin-pyfunc-server" "${MERLIN_DEP_CONSTRAINT}"
RUN conda env create --name merlin-model --file conda.yaml

# Download and dry-run user model artifacts and code
ARG MODEL_ARTIFACTS_URL
RUN if [ "${MLFLOW_ARTIFACT_STORAGE_TYPE}" = "gcs" ]; then  \
        gsutil -m cp -r ${MODEL_ARTIFACTS_URL} .; \
    elif [ "${MLFLOW_ARTIFACT_STORAGE_TYPE}" = "s3" ]; then \
        aws s3 cp ${MODEL_ARTIFACTS_URL} model --recursive; \
    else \
        echo "No credentials are used"; \
    fi
RUN /bin/bash -c ". activate merlin-model && merlin-pyfunc-server --model_dir model --dry_run"

CMD ["run.sh"]
