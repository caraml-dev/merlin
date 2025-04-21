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

import json

import mlflow
import urllib3
import pytest

import client as cl
import merlin
from merlin.model import ModelVersion

from unittest import mock
from unittest.mock import patch, MagicMock

default_resource_request = cl.ResourceRequest(min_replica=1, max_replica=1, cpu_request="100m", memory_request="128Mi")
env_1 = cl.Environment(
    id=1,
    name="dev",
    cluster="cluster-1",
    is_default=True,
    default_resource_request=default_resource_request,
)
env_2 = cl.Environment(
    id=2,
    name="dev-2",
    cluster="cluster-2",
    is_default=False,
    default_resource_request=default_resource_request,
)


def test_set_url(url, use_google_oauth):
    merlin.set_url(url, use_google_oauth=use_google_oauth)
    assert url == merlin.get_url()


def test_set_project(url, project, use_google_oauth):
    # expect exception when setting project but client is not set
    with pytest.raises(Exception):
        merlin.set_project(project.name)

    with patch("urllib3.PoolManager.request") as mock_request:
        mock_response = _mock_get_project_call(project)
        mock_request.return_value = mock_response

        merlin.set_url(url, use_google_oauth=use_google_oauth)
        merlin.set_project(project.name)

        assert merlin.active_project().name == project.name
        assert merlin.active_project().id == project.id
        assert merlin.active_project().mlflow_tracking_url == project.mlflow_tracking_url

def _mock_get_project_call(project) -> MagicMock:
    mock_response = MagicMock()
    mock_response.method = "GET"
    mock_response.status = 200
    mock_response.path = "/v1/projects"
    mock_response.data = bytes(f"""[{{
                        "id": {project.id},
                        "name": "{project.name}",
                        "mlflow_tracking_url": "{project.mlflow_tracking_url}",
                        "created_at": "{project.created_at}",
                        "updated_at": "{project.updated_at}"
                      }}]""", 'utf-8')
    mock_response.headers = {
        'content-type': 'application/json',
        'charset': 'utf-8'
    }
    
    return mock_response