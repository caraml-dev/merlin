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

ARG BASE_IMAGE=ghcr.io/caraml-dev/merlin/merlin-pyfunc-base:0.38.1
FROM ${BASE_IMAGE}

# Download and install user model dependencies
ARG MODEL_DEPENDENCIES_URL
COPY ${MODEL_DEPENDENCIES_URL} conda.yaml
RUN conda env create --name merlin-model --file conda.yaml

# Copy and install pyfunc-server and merlin-sdk dependencies
COPY merlin/python/pyfunc-server /pyfunc-server
COPY merlin/python/sdk /sdk
ENV SDK_PATH=/sdk

WORKDIR /pyfunc-server
RUN /bin/bash -c ". activate merlin-model && pip uninstall -y merlin-sdk && pip install -r /pyfunc-server/requirements.txt"

# Download and dry-run user model artifacts and code
ARG MODEL_ARTIFACTS_URL
COPY ${MODEL_ARTIFACTS_URL} model
RUN /bin/bash -c ". activate merlin-model && python -m pyfuncserver --model_dir model --dry_run"

CMD ["/bin/bash", "/pyfunc-server/run.sh"]
