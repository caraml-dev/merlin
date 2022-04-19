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
def test_resource_request_validate():
    resource_request = ResourceRequest(1, 2, "100m", "128Mi")
    resource_request.validate()

    resource_request.min_replica = 10
    with pytest.raises(Exception, match="Min replica must be less or equal to max replica"):
        resource_request.validate()
