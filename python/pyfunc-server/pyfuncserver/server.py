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

import tornado
from kfserving import KFServer
from kfserving.handlers.http import ExplainHandler
from kfserving.kfserver import HealthHandler, ListHandler, LivenessHandler

from pyfuncserver.protocol.rest.handler import CustomPredictHandler
from pyfuncserver.metrics.handler import MetricsHandler


class PyFuncServer(KFServer):
    def __init__(self, registry):
        super().__init__()
        self.registry = registry

    def create_application(self):
        return tornado.web.Application([
            # Server Liveness API returns 200 if server is alive.
            (r"/", LivenessHandler),
            (r"/v1/models",
             ListHandler, dict(models=self.registered_models)),
            # Model Health API returns 200 if model is ready to serve.
            (r"/v1/models/([a-zA-Z0-9_-]+)",
             HealthHandler, dict(models=self.registered_models)),
            (r"/v1/models/([a-zA-Z0-9_-]+):predict",
             CustomPredictHandler, dict(models=self.registered_models)),
            (r"/v1/models/([a-zA-Z0-9_-]+):explain",
             ExplainHandler, dict(models=self.registered_models)),
            (r"/metrics", MetricsHandler, dict(registry=self.registry))
        ])
