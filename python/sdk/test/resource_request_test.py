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

import pytest

from merlin.resource_request import ResourceRequest


@pytest.mark.unit
def test_resource_request():
    # Valid resource_request
    ResourceRequest(
        min_replica=1, max_replica=2, cpu_request="100m", memory_request="128Mi"
    )

    # Valid resource_request even though min_replica and max_replica are None
    ResourceRequest(cpu_request="100m", memory_request="128Mi")

    with pytest.raises(
        TypeError,
        match="'>' not supported between instances of 'int' and 'NoneType'",
    ):
        ResourceRequest(min_replica=1, cpu_request="100m", memory_request="128Mi")

    with pytest.raises(
        TypeError,
        match="'>' not supported between instances of 'NoneType' and 'int'",
    ):
        ResourceRequest(max_replica=1, cpu_request="100m", memory_request="128Mi")

    with pytest.raises(
        Exception, match="Min replica must be less or equal to max replica"
    ):
        ResourceRequest(
            min_replica=100, max_replica=2, cpu_request="100m", memory_request="128Mi"
        )

    with pytest.raises(Exception, match="Max replica must be greater than 0"):
        ResourceRequest(
            min_replica=0, max_replica=0, cpu_request="100m", memory_request="128Mi"
        )
