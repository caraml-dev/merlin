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

import datetime
import json
import urllib3
import pytest

from sys import version_info
from unittest import mock
from unittest.mock import patch, MagicMock

import client as cl
from client import ApiClient, Configuration
from merlin.client import MerlinClient
from merlin.endpoint import Status
from merlin.model import Model, ModelType, Project
from merlin.util import guess_mlp_ui_url
from merlin.version import VERSION


@pytest.fixture
def mock_url():
    return "http://merlin.dev/api"


@pytest.fixture
def api_client(mock_url):
    config = Configuration()
    config.host = mock_url + "/v1"
    return ApiClient(config)


mlflow_tracking_url = "http://mlflow.api.merlin.dev"
created_at = "2019-08-29T08:13:12.377Z"
updated_at = "2019-08-29T08:13:12.377Z"

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
mdl_endpoint_1 = cl.ModelEndpoint(
    id=1,
    model_id=1,
    model=None,
    status="serving",
    url="localhost",
    rule=None,
    environment_name=env_1.name,
    environment=env_1,
    created_at=created_at,
    updated_at=updated_at,
)


def serialize_datetime(obj):
    if isinstance(obj, datetime.datetime):
        return obj.isoformat()
    raise TypeError("Type is not serializable")

def test_get_project(mock_url, use_google_oauth):
    with patch("urllib3.PoolManager.request") as mock_request:
        mock_response = MagicMock()
        mock_response.status = 200
        mock_response.data = bytes(f"""[{{
                            "id": 0,
                            "name": "my-project",
                            "mlflow_tracking_url": "http://mlflow.api.merlin.dev",
                            "created_at": "{created_at}",
                            "updated_at": "{updated_at}"
                          }}]""", 'utf-8')
        mock_response.headers = {
            'content-type': 'application/json',
            'charset': 'utf-8'
        }
        mock_request.return_value = mock_response
    
        m = MerlinClient(mock_url, use_google_oauth=use_google_oauth)
        p = m.get_project("my-project")
        
        args, _ = mock_request.call_args
        assert args[0] == "GET"
        assert args[1] == "http://merlin.dev/api/v1/projects?name=my-project"
    
        assert p.id == 0
        assert p.name == "my-project"
        assert p.mlflow_tracking_url == "http://mlflow.api.merlin.dev"
        assert p.url == mock_url
        assert isinstance(p.created_at, datetime.datetime)
        assert isinstance(p.updated_at, datetime.datetime)
        
def test_create_invalid_project_name(
    mock_url, use_google_oauth
):
    project_name = "invalidProjectName"

    client = MerlinClient(mock_url, use_google_oauth=use_google_oauth)

    # Try to create project with invalid name. It must be fail
    with pytest.raises(Exception):
        assert client.get_project(project_name)