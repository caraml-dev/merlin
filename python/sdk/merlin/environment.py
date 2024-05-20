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

import client
from typing import Optional

from merlin.resource_request import ResourceRequest

class Environment:
    def __init__(self, env: client.Environment):
        self._name = env.name
        self._cluster = env.cluster
        self._is_default = env.is_default if env.is_default is not None else False
        self._default_resource_request = None
        if env.default_resource_request is not None:
            self._default_resource_request = ResourceRequest(env.default_resource_request.min_replica,
                                                            env.default_resource_request.max_replica,
                                                            env.default_resource_request.cpu_request,
                                                            env.default_resource_request.cpu_limit,
                                                            env.default_resource_request.memory_request)

    @property
    def name(self) -> str:
        return self._name

    @property
    def cluster(self) -> str:
        return self._cluster

    @property
    def is_default(self) -> bool:
        return self._is_default

    @property
    def default_resource_request(self) -> Optional[ResourceRequest]:
        return self._default_resource_request
