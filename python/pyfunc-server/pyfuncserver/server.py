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
from pyfuncserver.config import Config
from pyfuncserver.model.model import PyFuncModel
from pyfuncserver.protocol.rest.server import HTTPServer


class PyFuncServer:
    def __init__(self, config: Config):
        self.config = config

    def start(self, model: PyFuncModel):
        http_server = HTTPServer(port=self.config.http_port, workers=self.config.workers)
        http_server.start(model=model)
