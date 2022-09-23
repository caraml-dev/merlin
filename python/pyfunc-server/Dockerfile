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

WORKDIR /pyfunc-server
RUN gsutil cp -r ${MODEL_URL} .
RUN /bin/bash -c ". activate merlin-model && \
    sed -i 's/pip$/pip=20.2.4/' model/conda.yaml && \
    conda env update --name merlin-model --file model/conda.yaml && \
    python -m pyfuncserver --model_dir model --dry_run"

CMD ["/bin/bash", "./run.sh"]
