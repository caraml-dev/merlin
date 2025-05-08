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
    with patch("urllib3.PoolManager.request") as mock_request:
        # expect exception when setting project but client is not set
        mock_request.return_value = _mock_get_project_call_empty_result()
        with pytest.raises(Exception):
            merlin.set_project(project.name)
            
        mock_request.return_value = _mock_get_project_call(project)

        merlin.set_url(url, use_google_oauth=use_google_oauth)
        merlin.set_project(project.name)

        assert merlin.active_project().name == project.name
        assert merlin.active_project().id == project.id
        assert merlin.active_project().mlflow_tracking_url == project.mlflow_tracking_url
        
def test_set_model(url, project, model, use_google_oauth):
    with patch("urllib3.PoolManager.request") as mock_request:
        # expect exception when setting model but client and project is not set
        mock_request.return_value = _mock_get_project_call_empty_result()
        with pytest.raises(Exception):
            merlin.set_model(model.name, model.type)

        merlin.set_url(url, use_google_oauth=use_google_oauth)

        with pytest.raises(Exception):
            merlin.set_model(model.name, model.type)
        
        mock_request.return_value = _mock_get_project_call(project)
        merlin.set_project(project.name)

        mock_request.side_effect = [_mock_get_project_call(project), _mock_get_model_call(project, model)]
        merlin.set_model(model.name, model.type)

        assert merlin.active_model().name == model.name
        assert merlin.active_model().type == model.type
        assert merlin.active_model().id == model.id
        assert merlin.active_model().mlflow_experiment_id == model.mlflow_experiment_id
        
def test_new_model_version(url, project, model, version, use_google_oauth):
    with patch("urllib3.PoolManager.request") as mock_request:
        # expect exception when creating new model  version but client and
        # project is not set
        mock_request.return_value = _mock_get_model_call_empty_result(project)
        with pytest.raises(Exception):
            with merlin.new_model_version() as v:
                print(v)

        merlin.set_url(url, use_google_oauth=use_google_oauth)

        with pytest.raises(Exception):
            with merlin.new_model_version() as v:
                print(v)

        mock_request.return_value = _mock_get_project_call(project)
        merlin.set_project(project.name)

        mock_request.return_value = _mock_get_model_call_empty_result(project)
        with pytest.raises(Exception):
            with merlin.new_model_version() as v:
                print(v)

        mock_request.side_effect = [_mock_get_project_call(project), _mock_get_model_call(project, model)]
        merlin.set_model(model.name, model.type)

        mock_request.side_effect = [_mock_get_project_call(project), _mock_get_model_call(project, model), _mock_new_model_version_call(model, version)]
        with merlin.new_model_version() as v:
            assert v is not None
            assert isinstance(v, ModelVersion)

            assert v.mlflow_run_id == version.mlflow_run_id
            
def test_new_model_version_with_labels(
    url, project, model, version, use_google_oauth
):
    with patch("urllib3.PoolManager.request") as mock_request:
        merlin.set_url(url, use_google_oauth=use_google_oauth)
        
        mock_request.return_value = _mock_get_project_call(project)
        merlin.set_project(project.name)
        
        mock_request.side_effect = [_mock_get_project_call(project), _mock_get_model_call(project, model)]
        merlin.set_model(model.name, model.type)

        # # Insert labels
        labels = {"model": "T-800", "software": "skynet"}
        mock_request.side_effect = [_mock_get_project_call(project), _mock_get_model_call(project, model), _mock_new_model_version_call(model, version, labels)]

        with merlin.new_model_version(labels=labels) as v:
            assert v is not None
            assert isinstance(v, ModelVersion)

            assert v.mlflow_run_id == version.mlflow_run_id
            for key, value in v.labels.items():
                assert labels[key] == value

def test_list_environment(url, use_google_oauth):
    with patch("urllib3.PoolManager.request") as mock_request:
        merlin.set_url(url, use_google_oauth=use_google_oauth)

        mock_request.return_value = _mock_list_environment_call()
        envs = merlin.list_environment()

        assert len(envs) == 2
        assert envs[0].name == env_1.name
        assert envs[1].name == env_2.name
        
def test_get_environment(url, use_google_oauth):
    with patch("urllib3.PoolManager.request") as mock_request:
        merlin.set_url(url, use_google_oauth=use_google_oauth)

        mock_request.return_value = _mock_list_environment_call()
        env = merlin.get_environment(env_1.name)
        assert env is not None
        assert env.name == env_1.name

        env = merlin.get_environment("undefined_env")
        assert env is None
        
def test_get_default_environment(url, use_google_oauth):
    with patch("urllib3.PoolManager.request") as mock_request:
        merlin.set_url(url, use_google_oauth=use_google_oauth)
        
        mock_request.return_value = _mock_list_environment_call()
        env = merlin.get_default_environment()

        assert env is not None
        assert env.name == env_1.name
        assert env.is_default
        
def test_mlflow_methods(url, project, model, version, use_google_oauth):
    with patch("urllib3.PoolManager.request") as mock_request:
        merlin.set_url(url, use_google_oauth=use_google_oauth)
        
        mock_request.return_value = _mock_get_project_call(project)
        merlin.set_project(project.name)
        
        mock_request.side_effect = [_mock_get_project_call(project), _mock_get_model_call(project, model)]
        merlin.set_model(model.name, model.type)
        
        mock_request.side_effect = [_mock_get_project_call(project), _mock_get_model_call(project, model), _mock_new_model_version_call(model, version)]
        with merlin.new_model_version() as v:
            merlin.log_metric("metric", 0.1)
            merlin.log_param("param", "value")
            merlin.set_tag("tag", "value")
        run_id = v.mlflow_run_id
        run = mlflow.get_run(run_id=run_id)

        assert run.data.metrics["metric"] == 0.1
        assert run.data.params["param"] == "value"
        assert run.data.tags["tag"] == "value"

def _mock_get_project_call_empty_result() -> MagicMock:
    mock_response = MagicMock()
    mock_response.method = "GET"
    mock_response.status = 200
    mock_response.path = "/v1/projects"
    mock_response.data = bytes("[]", 'utf-8')
    mock_response.headers = {
        'content-type': 'application/json',
        'charset': 'utf-8'
    }
    
    return mock_response

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

def _mock_get_model_call(project, model) -> MagicMock:
    mock_response = MagicMock()
    mock_response.method = "GET"
    mock_response.status = 200
    mock_response.path = f"/v1/projects/{project.id}/models"
    mock_response.data = bytes(f"""[{{
                        "id": {model.id},
                        "mlflow_experiment_id": {model.mlflow_experiment_id},
                        "name": "{model.name}",
                        "type": "{model.type.value}",
                        "mlflow_url": "{model.mlflow_url}",
                        "created_at": "2019-09-04T03:09:13.842Z",
                        "updated_at": "2019-09-04T03:09:13.842Z"
                      }}]""", 'utf-8')
    mock_response.headers = {
        'content-type': 'application/json',
        'charset': 'utf-8'
    }
    
    return mock_response

def _mock_get_model_call_empty_result(project) -> MagicMock:
    mock_response = MagicMock()
    mock_response.method = "GET"
    mock_response.status = 200
    mock_response.path = f"/v1/projects/{project.id}/models"
    mock_response.data = bytes("[]", 'utf-8')
    mock_response.headers = {
        'content-type': 'application/json',
        'charset': 'utf-8'
    }
    
    return mock_response

def _mock_new_model_version_call(model, version, labels=None) -> MagicMock:
    body = {
        "id": version.id,
        "mlflow_run_id": version.mlflow_run_id,
        "is_serving": False,
        "mlflow_url": version.mlflow_url,
        "created_at": "2019-09-04T03:09:13.842Z",
        "updated_at": "2019-09-04T03:09:13.842Z",
    }
    if labels is not None:
        body["labels"] = labels

    mock_response = MagicMock()
    mock_response.method = "POST"
    mock_response.status = 201
    mock_response.path = f"/v1/models/{model.id}/versions"
    mock_response.data = json.dumps(body).encode('utf-8')
    mock_response.headers = {
        'content-type': 'application/json',
        'charset': 'utf-8'
    }
    
    return mock_response

def _mock_list_environment_call() -> MagicMock:
    mock_response = MagicMock()
    mock_response.method = "GET"
    mock_response.status = 200
    mock_response.path =  "/v1/environments"
    mock_response.data = json.dumps([env_1.to_dict(), env_2.to_dict()]).encode('utf-8')
    mock_response.headers = {
        'content-type': 'application/json',
        'charset': 'utf-8'
    }
    
    return mock_response