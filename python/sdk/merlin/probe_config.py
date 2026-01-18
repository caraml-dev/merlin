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
from __future__ import annotations

from typing import Optional

import client


class ProbeConfig:
    """
    Configuration for Kubernetes liveness/readiness probes.
    """

    def __init__(
        self,
        path: Optional[str] = None,
        port: Optional[int] = None,
        scheme: Optional[str] = None,
        initial_delay_seconds: Optional[int] = None,
        timeout_seconds: Optional[int] = None,
        period_seconds: Optional[int] = None,
        success_threshold: Optional[int] = None,
        failure_threshold: Optional[int] = None,
    ):
        self._path = path
        self._port = port
        self._scheme = scheme
        self._initial_delay_seconds = initial_delay_seconds
        self._timeout_seconds = timeout_seconds
        self._period_seconds = period_seconds
        self._success_threshold = success_threshold
        self._failure_threshold = failure_threshold

    @classmethod
    def from_response(cls, response) -> Optional["ProbeConfig"]:
        """Create a ProbeConfig from a client.ProbeConfig response object."""
        if response is None:
            return None
        return ProbeConfig(
            path=response.path,
            port=response.port,
            scheme=response.scheme,
            initial_delay_seconds=response.initial_delay_seconds,
            timeout_seconds=response.timeout_seconds,
            period_seconds=response.period_seconds,
            success_threshold=response.success_threshold,
            failure_threshold=response.failure_threshold,
        )

    def to_client_probe_config(self):
        """Convert to a client.ProbeConfig object for API calls."""
        return client.ProbeConfig(
            path=self._path,
            port=self._port,
            scheme=self._scheme,
            initial_delay_seconds=self._initial_delay_seconds,
            timeout_seconds=self._timeout_seconds,
            period_seconds=self._period_seconds,
            success_threshold=self._success_threshold,
            failure_threshold=self._failure_threshold,
        )

    @property
    def path(self) -> Optional[str]:
        return self._path

    @path.setter
    def path(self, path):
        self._path = path

    @property
    def port(self) -> Optional[int]:
        return self._port

    @port.setter
    def port(self, port):
        self._port = port

    @property
    def scheme(self) -> Optional[str]:
        return self._scheme

    @scheme.setter
    def scheme(self, scheme):
        self._scheme = scheme

    @property
    def initial_delay_seconds(self) -> Optional[int]:
        return self._initial_delay_seconds

    @initial_delay_seconds.setter
    def initial_delay_seconds(self, initial_delay_seconds):
        self._initial_delay_seconds = initial_delay_seconds

    @property
    def timeout_seconds(self) -> Optional[int]:
        return self._timeout_seconds

    @timeout_seconds.setter
    def timeout_seconds(self, timeout_seconds):
        self._timeout_seconds = timeout_seconds

    @property
    def period_seconds(self) -> Optional[int]:
        return self._period_seconds

    @period_seconds.setter
    def period_seconds(self, period_seconds):
        self._period_seconds = period_seconds

    @property
    def success_threshold(self) -> Optional[int]:
        return self._success_threshold

    @success_threshold.setter
    def success_threshold(self, success_threshold):
        self._success_threshold = success_threshold

    @property
    def failure_threshold(self) -> Optional[int]:
        return self._failure_threshold

    @failure_threshold.setter
    def failure_threshold(self, failure_threshold):
        self._failure_threshold = failure_threshold

