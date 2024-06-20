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


class ResourceRequest:
    """
    The resource requirement and replicas requests for Merlin components (e.g. version endpoint, batch prediction job, image builder).
    """

    def __init__(
        self,
        min_replica: Optional[int] = None,
        max_replica: Optional[int] = None,
        cpu_request: Optional[str] = None,
        cpu_limit: Optional[str] = None,
        memory_request: Optional[str] = None,
        gpu_request: Optional[str] = None,
        gpu_name: Optional[str] = None,
    ):
        self._min_replica = min_replica
        self._max_replica = max_replica
        self._cpu_request = cpu_request
        self._cpu_limit = cpu_limit
        self._memory_request = memory_request
        self._gpu_request = gpu_request
        self._gpu_name = gpu_name
        self.validate()

    @classmethod
    def from_response(cls, response: client.ResourceRequest) -> ResourceRequest:
        return ResourceRequest(
            min_replica=response.min_replica,
            max_replica=response.max_replica,
            cpu_request=response.cpu_request,
            cpu_limit=response.cpu_limit,
            memory_request=response.memory_request,
            gpu_request=response.gpu_request,
            gpu_name=response.gpu_name,
        )

    @property
    def min_replica(self) -> Optional[int]:
        return self._min_replica

    @min_replica.setter
    def min_replica(self, min_replica):
        self._min_replica = min_replica

    @property
    def max_replica(self) -> Optional[int]:
        return self._max_replica

    @max_replica.setter
    def max_replica(self, max_replica):
        self._max_replica = max_replica

    @property
    def cpu_request(self) -> Optional[str]:
        return self._cpu_request

    @cpu_request.setter
    def cpu_request(self, cpu_request):
        self._cpu_request = cpu_request

    @property
    def cpu_limit(self) -> Optional[str]:
        return self._cpu_limit

    @cpu_limit.setter
    def cpu_limit(self, cpu_limit):
        self._cpu_limit = cpu_limit

    @property
    def memory_request(self) -> Optional[str]:
        return self._memory_request

    @memory_request.setter
    def memory_request(self, memory_request):
        self._memory_request = memory_request

    @property
    def gpu_request(self) -> Optional[str]:
        return self._gpu_request

    @gpu_request.setter
    def gpu_request(self, gpu_request):
        self._gpu_request = gpu_request

    @property
    def gpu_name(self) -> Optional[str]:
        return self._gpu_name

    @gpu_name.setter
    def gpu_name(self, gpu_name):
        self._gpu_name = gpu_name

    def validate(self):
        if self._min_replica is None and self._max_replica is None:
            return

        if self._min_replica > self._max_replica:
            raise Exception("Min replica must be less or equal to max replica")

        if self._max_replica < 1:
            raise Exception("Max replica must be greater than 0")
