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
from sys import version_info
from unittest import mock

import pytest

import client as cl
from client import ApiClient, Configuration
from merlin.client import MerlinClient
from merlin.endpoint import Status
from merlin.model import Model, ModelType, Project
from merlin.util import guess_mlp_ui_url
from merlin.version import VERSION

# get global mock responses that configured in conftest
responses = pytest.responses


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

default_resource_request = cl.ResourceRequest(1, 1, "100m", "128Mi")
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


@responses.activate
def test_get_project(mock_url, mock_oauth, use_google_oauth):
    responses.add(
        "GET",
        "/api/v1/projects",
        body=f"""[{{
                        "id": 0,
                        "name": "my-project",
                        "mlflow_tracking_url": "http://mlflow.api.merlin.dev",
                        "created_at": "{created_at}",
                        "updated_at": "{updated_at}"
                      }}]""",
        status=200,
        content_type="application/json",
    )

    m = MerlinClient(mock_url, use_google_oauth=use_google_oauth)
    p = m.get_project("my-project")

    assert responses.calls[-1].request.method == "GET"
    assert responses.calls[-1].request.url == "/api/v1/projects?name=my-project"
    assert responses.calls[-1].request.host == "merlin.dev"

    assert p.id == 0
    assert p.name == "my-project"
    assert p.mlflow_tracking_url == "http://mlflow.api.merlin.dev"
    assert p.url == mock_url
    assert isinstance(p.created_at, datetime.datetime)
    assert isinstance(p.updated_at, datetime.datetime)


@responses.activate
def test_create_invalid_project_name(
    mock_url, api_client, mock_oauth, use_google_oauth
):
    project_name = "invalidProjectName"

    client = MerlinClient(mock_url, use_google_oauth=use_google_oauth)

    # Try to create project with invalid name. It must be fail
    with pytest.raises(Exception):
        assert client.get_project(project_name)


@responses.activate
def test_create_model(mock_url, api_client, mock_oauth, use_google_oauth):
    project_id = "1010"
    mlflow_experiment_id = 1
    model_name = "my-model"
    project_name = "my-project"
    model_type = ModelType.XGBOOST
    mlflow_url = "http://mlflow.api.merlin.dev"

    responses.add(
        "GET",
        f"/api/v1/projects/{project_id}/models",
        body="[]",
        status=200,
        content_type="application/json",
    )
    responses.add(
        "POST",
        f"/api/v1/projects/{project_id}/models",
        body=f"""{{
                        "id": 0,
                        "project_id": {project_id},
                        "mlflow_experiment_id": {mlflow_experiment_id},
                        "name": "{model_name}",
                        "type": "{model_type.value}",
                        "mlflow_url": "{mlflow_url}",
                        "endpoints": [],
                        "created_at": "{created_at}",
                        "updated_at": "{updated_at}"
                      }}""",
        status=200,
        content_type="application/json",
    )

    client = MerlinClient(mock_url, use_google_oauth=use_google_oauth)
    prj = cl.Project(
        project_id, project_name, mlflow_tracking_url, created_at, updated_at
    )
    project = Project(prj, mock_url, api_client)
    with mock.patch.object(client, "get_project", return_value=project):
        model = client.get_or_create_model(
            "my-model", project_name=project_name, model_type=model_type
        )

        assert json.loads(responses.calls[-1].request.body) == json.loads(
            f"""
        {{
            "name" : "{model_name}",
            "type" : "{model_type.value}"
        }}
        """
        )
        assert model.id == 0
        assert model.mlflow_experiment_id == mlflow_experiment_id
        assert model.name == model_name
        assert model.type == model_type
        assert model.mlflow_url == mlflow_tracking_url
        assert model.mlflow_experiment_id == mlflow_experiment_id
        assert isinstance(model.created_at, datetime.datetime)
        assert isinstance(model.updated_at, datetime.datetime)
        assert model.project == project
        assert (
            f"merlin-sdk/{VERSION}" in responses.calls[-1].request.headers["User-Agent"]
        )
        assert (
            f"python/{version_info.major}.{version_info.minor}.{version_info.micro}"
            in responses.calls[-1].request.headers["User-Agent"]
        )


@responses.activate
def test_create_invalid_model_name(mock_url, api_client, mock_oauth, use_google_oauth):
    model_name = "invalidModelName"
    project_name = "my-project"
    model_type = ModelType.XGBOOST

    client = MerlinClient(mock_url, use_google_oauth=use_google_oauth)

    # Try to create model with invalid name. It must be fail
    with pytest.raises(Exception):
        assert client.get_or_create_model(model_name, project_name, model_type)


@responses.activate
def test_get_model(mock_url, api_client, mock_oauth, use_google_oauth):
    project_id = "1010"
    mlflow_experiment_id = 1
    model_name = "my-model"
    project_name = "my-project"
    model_type = ModelType.XGBOOST
    mlflow_url = "http://mlflow.api.merlin.dev"

    responses.add(
        "GET",
        f"/api/v1/projects/{project_id}/models",
        body=f"""[{{
                        "id": 1,
                        "project_id": {project_id},
                        "mlflow_experiment_id": {mlflow_experiment_id},
                        "name": "{model_name}",
                        "type": "{model_type.value}",
                        "mlflow_url": "{mlflow_url}",
                        "endpoints": [],
                        "created_at": "{created_at}",
                        "updated_at": "{updated_at}"
                      }}]""",
        status=200,
        content_type="application/json",
    )

    responses.add(
        "GET",
        f"/api/v1/models/1/endpoints",
        body=json.dumps([mdl_endpoint_1.to_dict()]),
        status=200,
        content_type="application/json",
    )

    client = MerlinClient(mock_url, use_google_oauth=use_google_oauth)
    prj = cl.Project(
        project_id, project_name, mlflow_tracking_url, created_at, updated_at
    )
    project = Project(prj, mock_url, api_client)
    with mock.patch.object(client, "get_project", return_value=project):
        model = client.get_or_create_model(
            "my-model", project_name=project_name, model_type=model_type
        )
        assert model.id == 1
        assert model.name == model_name
        assert model.type == model_type
        assert model.mlflow_url == mlflow_tracking_url
        assert model.mlflow_experiment_id == mlflow_experiment_id
        assert isinstance(model.created_at, datetime.datetime)
        assert isinstance(model.updated_at, datetime.datetime)
        assert model.project == project

        default_model_endpoint = model.endpoint
        assert default_model_endpoint is not None
        assert default_model_endpoint.status == Status.SERVING
        assert default_model_endpoint.environment_name == env_1.name


@responses.activate
def test_new_model_version(mock_url, api_client, mock_oauth, use_google_oauth):
    project_id = 1
    model_id = 1
    version_id = 2
    model_name = "my-model"
    project_name = "my-project"
    mlflow_experiment_id = 1
    mlflow_run_id = "c5c3b6b220b34c7496de8c0400b7c793"
    model_type = ModelType.TENSORFLOW
    mlflow_url = "http://mlflow.api.merlin.dev"
    artifact_uri = "gs://zltest/model"
    created_at = "2019-09-04T03:09:13.842Z"
    updated_at = "2019-09-04T03:09:13.843Z"

    responses.add(
        "POST",
        f"/api/v1/models/{model_id}/versions",
        body=f"""{{
                        "id": {version_id},
                        "model_id": {model_id},
                        "mlflow_run_id": "{mlflow_run_id}",
                        "mlflow_url": "{mlflow_url}",
                        "artifact_uri": "{artifact_uri}",
                        "endpoints": [],
                        "mlflow_url": "{mlflow_url}",
                        "created_at": "{created_at}",
                        "updated_at": "{updated_at}"
                      }}""",
        status=200,
        content_type="application/json",
    )

    client = MerlinClient(mock_url, use_google_oauth=use_google_oauth)
    prj = cl.Project(
        project_id, project_name, mlflow_tracking_url, created_at, updated_at
    )
    project = Project(prj, mock_url, api_client)
    mdl = cl.Model(
        model_id,
        project_id,
        mlflow_experiment_id,
        model_name,
        model_type.value,
        mlflow_url,
        None,
        created_at,
        updated_at,
    )
    mdl = Model(mdl, project, api_client)
    with mock.patch.object(client, "get_model", return_value=mdl):
        mv = client.new_model_version(model_name, project_name)

        assert mv.id == version_id
        assert mv.mlflow_run_id == mlflow_run_id
        assert mv.mlflow_url == mlflow_url
        assert mv.properties is None
        assert isinstance(mv.created_at, datetime.datetime)
        assert isinstance(mv.updated_at, datetime.datetime)
        assert mv.model == mdl
        ui_url = guess_mlp_ui_url(mock_url)
        assert mv.url == f"{ui_url}/projects/1/models/{model_id}/versions"


@responses.activate
def test_list_environments(mock_url, api_client, mock_oauth, use_google_oauth):
    responses.add(
        "GET",
        "/api/v1/environments",
        body=json.dumps([env_1.to_dict(), env_2.to_dict()]),
        status=200,
        content_type="application/json",
    )
    client = MerlinClient(mock_url, use_google_oauth=use_google_oauth)
    envs = client.list_environment()

    assert len(envs) == 2
    assert envs[0].name == env_1.name
    assert envs[0].cluster == env_1.cluster
    assert envs[0].is_default == env_1.is_default
    assert envs[1].name == env_2.name
    assert envs[1].cluster == env_2.cluster
    assert envs[1].is_default == env_2.is_default


@responses.activate
def test_get_environment(mock_url, api_client, mock_oauth, use_google_oauth):
    responses.add(
        "GET",
        "/api/v1/environments",
        body=json.dumps([env_1.to_dict(), env_2.to_dict()]),
        status=200,
        content_type="application/json",
    )
    client = MerlinClient(mock_url, use_google_oauth=use_google_oauth)
    env = client.get_environment(env_1.name)

    assert env is not None
    assert env.name == env_1.name
    assert env.cluster == env_1.cluster
    assert env.is_default == env_1.is_default

    env = client.get_environment("undefined_env")
    assert env is None


@responses.activate
def test_get_default_environment(mock_url, api_client, mock_oauth, use_google_oauth):
    responses.add(
        "GET",
        "/api/v1/environments",
        body=json.dumps([env_1.to_dict(), env_2.to_dict()]),
        status=200,
        content_type="application/json",
    )
    client = MerlinClient(mock_url, use_google_oauth=use_google_oauth)
    env = client.get_default_environment()

    assert env.name == env_1.name
    assert env.cluster == env_1.cluster
    assert env.is_default == env_1.is_default

    responses.reset()

    responses.add(
        "GET",
        "/api/v1/environments",
        body=json.dumps([env_2.to_dict()]),
        status=200,
        content_type="application/json",
    )
    env = client.get_default_environment()

    assert env is None


@responses.activate
def test_get_default_environment(mock_url, api_client, mock_oauth, use_google_oauth):
    client = MerlinClient(mock_url, use_google_oauth=use_google_oauth)
    responses.add(
        "GET",
        "/api/v1/environments",
        body=json.dumps([env_2.to_dict()]),
        status=200,
        content_type="application/json",
    )
    env = client.get_default_environment()

    assert env is None
