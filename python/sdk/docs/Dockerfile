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

FROM condaforge/miniforge3:23.3.1-1

RUN apt-get -qq update \
    && apt-get -y install vim make wget curl \
    && conda install pandoc=1.19.2

COPY . /tmp/merlin-sdk
WORKDIR /tmp/merlin-sdk/docs

RUN pip install -r requirements_docs.txt \
&& make html


FROM node:13-stretch
RUN npm install -g serve
COPY --from=0 /tmp/merlin-sdk/docs/_build /opt/merlin-docs
COPY --from=0 /tmp/merlin-sdk/docs/serve.json /opt/merlin-docs/html/serve.json

WORKDIR  /opt/merlin-docs

CMD ["serve", "html"]