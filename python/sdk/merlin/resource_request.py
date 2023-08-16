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

from typing import Optional, Dict

class ResourceRequest:
    """
    The resource requirement and replicas requests for model version endpoint.
    """

    def __init__(self, min_replica: int, max_replica: int, cpu_request: str, memory_request: str, gpu_request: str=None, gpu_resource_type: str=None, gpu_node_selector: Dict=None):
        self._min_replica = min_replica
        self._max_replica = max_replica
        self._cpu_request = cpu_request
        self._memory_request = memory_request
        self._gpu_request = gpu_request
        self._gpu_resource_type = gpu_resource_type
        self._gpu_node_selector = gpu_node_selector
        self.validate()

    @property
    def min_replica(self) -> int:
        return self._min_replica

    @min_replica.setter
    def min_replica(self, min_replica):
        self._min_replica = min_replica

    @property
    def max_replica(self) -> int:
        return self._max_replica

    @max_replica.setter
    def max_replica(self, max_replica):
        self._max_replica = max_replica

    @property
    def cpu_request(self) -> str:
        return self._cpu_request

    @cpu_request.setter
    def cpu_request(self, cpu_request):
        self._cpu_request = cpu_request

    @property
    def memory_request(self) -> str:
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
    def gpu_resource_type(self) -> Optional[str]:
        return self._gpu_resource_type
    
    @gpu_resource_type.setter
    def gpu_resource_type(self, gpu_resource_type):
        self._gpu_resource_type = gpu_resource_type

    @property
    def gpu_node_selector(self) -> Optional[Dict]:
        return self._gpu_node_selector

    @gpu_node_selector.setter
    def gpu_node_selector(self, gpu_node_selector):
        self._gpu_node_selector = gpu_node_selector


    def validate(self):
        if self._min_replica > self._max_replica:
            raise Exception("Min replica must be less or equal to max replica")
