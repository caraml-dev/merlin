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
import threading

import prometheus_client
from merlin.protocol import Protocol
from prometheus_client import CollectorRegistry, multiprocess

from pyfuncserver.config import Config
from pyfuncserver.model.model import PyFuncModel
from pyfuncserver.protocol.rest.server import HTTPServer
from pyfuncserver.protocol.upi.server import UPIServer


class PyFuncServer:
    def __init__(self, config: Config):
        self._config = config

    def start(self, model: PyFuncModel):
        # initialize prometheus
        # register to MultiProcessCollector as PyFuncServer will run in multiple process
        registry = CollectorRegistry()
        multiprocess.MultiProcessCollector(registry)

        # start only one server depending on the chosen protocol
        # the intent is to save memory consumption
        if self._config.protocol == Protocol.HTTP_JSON:
            http_server = HTTPServer(port=self._config.http_port, workers=self._config.workers, metrics_registry=registry)
            http_server.start(model=model)
        elif self._config.protocol == Protocol.UPI_V1:
            # start prometheus metrics server and listen at http port
            prometheus_client.start_http_server(self._config.http_port, registry=registry)

            # start grpc/upi server and listen at grpc port
            upi_server = UPIServer(model=model, config=self._config)
            upi_server.start()
        else:
            raise ValueError(f"unknown protocol {self._config.protocol}")