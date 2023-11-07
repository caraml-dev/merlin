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

ARG MODEL_URL
ARG GOOGLE_APPLICATION_CREDENTIALS

WORKDIR /pyfunc-server

# Run docker build using the credentials if provided
RUN if [ ! -z "$GOOGLE_APPLICATION_CREDENTIALS" ]; then gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS}; fi
RUN gsutil cp -r ${MODEL_URL} .
RUN /bin/bash -c ". activate merlin-model && \
    sed -i 's/mlflow\(\s*==\s*[^ ]*\)\{0,1\}/mlflow==1.26.1/g' model/conda.yaml && \
    conda env update --name merlin-model --file model/conda.yaml && \
    python -m pyfuncserver --model_dir model --dry_run"

CMD ["/bin/bash", "./run.sh"]
