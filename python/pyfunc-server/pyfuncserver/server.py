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
import logging

import prometheus_client
from merlin.protocol import Protocol
from prometheus_client import CollectorRegistry, multiprocess

from pyfuncserver.config import Config
from pyfuncserver.metrics.pusher import labels, start_metrics_pusher
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
            http_server = HTTPServer(model=model, config=self._config, metrics_registry=registry)
            http_server.start()
        elif self._config.protocol == Protocol.UPI_V1:
            # Due to https://github.com/knative/serving/issues/8471, we have to resort to pushing metrics to
            # prometheus push gateway.
            if (self._config.push_gateway.enabled):
                target_info = labels(self._config)
                start_metrics_pusher(self._config.push_gateway.url,
                                     registry,
                                     target_info,
                                     self._config.push_gateway.push_interval_sec)
            else:
                # start prometheus metrics server and listen at http port
                logging.info(f"starting metrics server at {self._config.http_port}")
                prometheus_client.start_http_server(self._config.http_port, registry=registry)

            # start grpc/upi server and listen at grpc port
            upi_server = UPIServer(model=model, config=self._config)
            upi_server.start()
        else:
            raise ValueError(f"unknown protocol {self._config.protocol}")
