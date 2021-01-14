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

from enum import Enum
from typing import Dict

import client
from merlin.environment import Environment
from merlin.logger import Logger, LoggerConfig, LoggerMode
from merlin.util import autostr


class Status(Enum):
    PENDING = 'pending'
    RUNNING = 'running'
    SERVING = 'serving'
    FAILED = 'failed'
    TERMINATED = 'terminated'


@autostr
class VersionEndpoint:
    def __init__(self, endpoint: client.VersionEndpoint, log_url: str = None):
        self._url = f"{endpoint.url}"
        if ":predict" not in endpoint.url:
            self._url = f"{endpoint.url}:predict"

        self._status = Status(endpoint.status)
        self._id = endpoint.id
        self._environment_name = endpoint.environment_name
        self._environment = Environment(endpoint.environment)
        self._env_vars = endpoint.env_vars
        self._logger = Logger.from_logger_response(endpoint.logger)
        if log_url is not None:
            self._log_url = log_url

    @property
    def url(self):
        return self._url

    @property
    def status(self) -> Status:
        return self._status

    @property
    def id(self) -> str:
        return self._id

    @property
    def environment_name(self) -> str:
        return self._environment_name

    @property
    def environment(self) -> Environment:
        return self._environment

    @property
    def env_vars(self) -> Dict[str, str]:
        env_vars = {}
        for ev in self._env_vars:
            env_vars[ev.name] = ev.value
        return env_vars

    @property
    def logger(self) -> Logger:
        return self._logger

    @property
    def log_url(self) -> str:
        return self._log_url

    def _repr_html_(self):
        return f"""<a href="{self._url}">{self._url}</a>"""


@autostr
class ModelEndpoint:
    def __init__(self, endpoint: client.ModelEndpoint):
        self._url = f"{endpoint.url}/v1/predict" \
            if endpoint.url.startswith("http://") \
            else f"http://{endpoint.url}/v1/predict"
        self._status = Status(endpoint.status)
        self._id = endpoint.id
        self._environment_name = endpoint.environment_name
        self._environment = Environment(endpoint.environment)

    @property
    def url(self):
        return self._url

    @property
    def status(self) -> Status:
        return self._status

    @property
    def id(self) -> str:
        return str(self._id)

    @property
    def environment_name(self) -> str:
        return self._environment_name

    @property
    def environment(self) -> Environment:
        return self._environment

    def _repr_html_(self):
        return f"""<a href="{self._url}">{self._url}</a>"""
