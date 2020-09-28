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

from __future__ import absolute_import

# flake8: noqa

# import apis into api package
from client.api.alert_api import AlertApi
from client.api.endpoint_api import EndpointApi
from client.api.environment_api import EnvironmentApi
from client.api.log_api import LogApi
from client.api.model_endpoints_api import ModelEndpointsApi
from client.api.models_api import ModelsApi
from client.api.prediction_jobs_api import PredictionJobsApi
from client.api.project_api import ProjectApi
from client.api.secret_api import SecretApi
from client.api.version_api import VersionApi
