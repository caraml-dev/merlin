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
COPY model /model
RUN /bin/bash -c ". activate merlin-model && conda env update --name merlin-model --file /model/conda.yaml && python ${HOME}/merlin-spark-app/main.py --dry-run-model /model"