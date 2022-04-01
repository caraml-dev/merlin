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
import pytest
from urllib3_mock import Responses

import client as cl
import merlin
from merlin.model import ModelVersion

responses = Responses('requests.packages.urllib3')

default_resource_request = cl.ResourceRequest(1, 1, "100m", "128Mi")
env_1 = cl.Environment(1, "dev", "cluster-1", True, default_resource_request=default_resource_request)
env_2 = cl.Environment(2, "dev-2", "cluster-2", False, default_resource_request=default_resource_request)


def test_set_url(url, use_google_oauth):
    merlin.set_url(url, use_google_oauth=use_google_oauth)
    assert url == merlin.get_url()


@responses.activate
def test_set_project(url, project, mock_oauth, use_google_oauth):
    # expect exception when setting project but client is not set
    with pytest.raises(Exception):
        merlin.set_project(project.name)

    _mock_get_project_call(project)

    merlin.set_url(url, use_google_oauth=use_google_oauth)
    merlin.set_project(project.name)

    assert merlin.active_project().name == project.name
    assert merlin.active_project().id == project.id
    assert merlin.active_project().mlflow_tracking_url == project.mlflow_tracking_url


@responses.activate
def test_set_model(url, project, model, mock_oauth, use_google_oauth):
    # expect exception when setting model but client and project is not set
    with pytest.raises(Exception):
        merlin.set_model(model.name, model.type)

    merlin.set_url(url, use_google_oauth=use_google_oauth)

    with pytest.raises(Exception):
        merlin.set_model(model.name, model.type)

    _mock_get_project_call(project)
    merlin.set_project(project.name)

    _mock_get_model_call(project, model)
    merlin.set_model(model.name, model.type)

    assert merlin.active_model().name == model.name
    assert merlin.active_model().type == model.type
    assert merlin.active_model().id == model.id
    assert merlin.active_model().mlflow_experiment_id == model.mlflow_experiment_id


@responses.activate
def test_new_model_version(url, project, model, version, mock_oauth, use_google_oauth):
    # expect exception when creating new model  version but client and
    # project is not set
    with pytest.raises(Exception):
        with merlin.new_model_version() as v:
            print(v)

    merlin.set_url(url, use_google_oauth=use_google_oauth)

    with pytest.raises(Exception):
        with merlin.new_model_version() as v:
            print(v)

    _mock_get_project_call(project)
    merlin.set_project(project.name)

    with pytest.raises(Exception):
        with merlin.new_model_version() as v:
            print(v)

    _mock_get_model_call(project, model)
    merlin.set_model(model.name, model.type)

    _mock_new_model_version_call(model, version)
    with merlin.new_model_version() as v:
        assert v is not None
        assert isinstance(v, ModelVersion)

        assert v.mlflow_run_id == version.mlflow_run_id


@responses.activate
def test_new_model_version_with_labels(url, project, model, version, mock_oauth, use_google_oauth):
    merlin.set_url(url, use_google_oauth=use_google_oauth)
    _mock_get_project_call(project)
    merlin.set_project(project.name)
    _mock_get_model_call(project, model)
    merlin.set_model(model.name, model.type)

    # Insert labels
    labels = {"model": "T-800", "software": "skynet"}
    _mock_new_model_version_call(model, version, labels)

    with merlin.new_model_version(labels=labels) as v:
        assert v is not None
        assert isinstance(v, ModelVersion)

        assert v.mlflow_run_id == version.mlflow_run_id
        for key, value in v.labels.items():
            assert labels[key] == value


@responses.activate
def test_list_environment(url, mock_oauth, use_google_oauth):
    merlin.set_url(url, use_google_oauth=use_google_oauth)

    _mock_list_environment_call()

    envs = merlin.list_environment()

    assert len(envs) == 2
    assert envs[0].name == env_1.name
    assert envs[1].name == env_2.name


@responses.activate
def test_get_environment(url, mock_oauth, use_google_oauth):
    merlin.set_url(url, use_google_oauth=use_google_oauth)

    _mock_list_environment_call()

    env = merlin.get_environment(env_1.name)
    assert env is not None
    assert env.name == env_1.name

    env = merlin.get_environment("undefined_env")
    assert env is None


@responses.activate
def test_get_default_environment(url, mock_oauth, use_google_oauth):
    merlin.set_url(url, use_google_oauth=use_google_oauth)

    _mock_list_environment_call()

    env = merlin.get_default_environment()

    assert env is not None
    assert env.name == env_1.name
    assert env.is_default


@responses.activate
def test_mlflow_methods(url, project, model, version, mock_oauth, use_google_oauth):
    _mock_get_project_call(project)
    _mock_get_model_call(project, model)
    _mock_new_model_version_call(model, version)

    merlin.set_url(url, use_google_oauth=use_google_oauth)
    merlin.set_project(project.name)
    merlin.set_model(model.name, model.type)
    with merlin.new_model_version() as v:
        merlin.log_metric("metric", 0.1)
        merlin.log_param("param", "value")
        merlin.set_tag("tag", "value")
    run_id = v.mlflow_run_id
    run = mlflow.get_run(run_id=run_id)

    assert run.data.metrics["metric"] == 0.1
    assert run.data.params["param"] == "value"
    assert run.data.tags["tag"] == "value"


def _mock_get_project_call(project):
    responses.add('GET', '/v1/projects',
                  body=f"""[{{
                        "id": {project.id},
                        "name": "{project.name}",
                        "mlflow_tracking_url": "{project.mlflow_tracking_url}",
                        "created_at": "{project.created_at}",
                        "updated_at": "{project.updated_at}"
                      }}]""",
                  status=200,
                  content_type='application/json'
                  )


def _mock_get_model_call(project, model):
    responses.add('GET', f"/v1/projects/{project.id}/models",
                  body=f"""[{{
                        "id": {model.id},
                        "mlflow_experiment_id": "{model.mlflow_experiment_id}",
                        "name": "{model.name}",
                        "type": "{model.type.value}",
                        "mlflow_url": "{model.mlflow_url}",
                        "created_at": "2019-09-04T03:09:13.842Z",
                        "updated_at": "2019-09-04T03:09:13.842Z"
                      }}]""",
                  status=200,
                  content_type='application/json'
                  )


def _mock_new_model_version_call(model, version, labels=None):
    body = {
        "id": version.id,
        "mlflow_run_id": version.mlflow_run_id,
        "is_serving": False,
        "mlflow_url": version.mlflow_url,
        "created_at": "2019-09-04T03:09:13.842Z",
        "updated_at": "2019-09-04T03:09:13.842Z"
    }
    if labels is not None:
        body["labels"] = labels

    responses.add('POST', f"/v1/models/{model.id}/versions",
                  body=json.dumps(body),
                  status=200,
                  content_type='application/json')


def _mock_list_environment_call():
    responses.add("GET", "/v1/environments", body=json.dumps([env_1.to_dict(),
                                                              env_2.to_dict()]),
                  status=200,
                  content_type='application/json')


@pytest.fixture
def mock_oauth():
    responses.add('GET',
                  '/computeMetadata/v1/instance/service-accounts/default/',
                  body="""
    {
        "email": "computeengine@google.com",
        "scopes": ["scope"],
        "aliases": ["default"]
    }
    """, status=200, content_type='application/json')
    responses.add('GET',
                  '/computeMetadata/v1/instance/service-accounts/computeengine@google.com/token',
                  body="""
    {
        "access_token": "ya29.ImCpB6BS2mdOMseaUjhVlHqNfAOz168XjuDrK7Sd33glPd7XvtMLIngi1-V52ReytFSUluE-iBV88OlDkjtraggB_qc-LN2JlGtQ3sHZq_MuTxrU0-oK_kpq-1wsvniFFGQ",
        "expires_in": 3600,
        "scope": "openid https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/userinfo.email",
        "token_type": "Bearer",
        "id_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6IjhhNjNmZTcxZTUzMDY3NTI0Y2JiYzZhM2E1ODQ2M2IzODY0YzA3ODciLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20iLCJhenAiOiI3NjQwODYwNTE4NTAtNnFyNHA2Z3BpNmhuNTA2cHQ4ZWp1cTgzZGkzNDFodXIuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJhdWQiOiI3NjQwODYwNTE4NTAtNnFyNHA2Z3BpNmhuNTA2cHQ4ZWp1cTgzZGkzNDFodXIuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJzdWIiOiIxMDM5ODg0MzM2OTY3NzI1NDkzNjAiLCJoZCI6ImdvLWplay5jb20iLCJlbWFpbCI6InByYWRpdGh5YS5wdXJhQGdvLWplay5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiYXRfaGFzaCI6ImdrbXIxY0dPTzNsT0dZUDhtYjNJRnciLCJpYXQiOjE1NzE5Njg3NDUsImV4cCI6MTU3MTk3MjM0NX0.FIY5xvySNVxt1cbw-QXdDfiwollxcqupz1YDJuP14obKRyDwFG9ZcC_j-mTDZF5_dzpYeNMMK-LPTq9QIaM-blSKm2Eh9LeMvQGUk_S-9y_r2jKCmBlrEeHM8DXk3xyKf65LEoBA8cwMPdgb2s8AMIxxN9JJ09fjou20yLDI84Q4BFMriMIBBYLFgBW0wcg2PQ1hy5hrV1PdZj-ZNKNWmouh0lOjLLYmVFZPCzD9ENWo1N52ZLaLODdI2gDcpbyTUbeAh81sacdtJd0pLf-FuBLdfuktvP4MVvdmIhXv98Zb0dFBzRtmiqlQusSjoG5VEaBc6o2gkM5rHR0ozby0Fg"
    }
    """, status=200, content_type='application/json')
    responses.add('POST', '/token', body="""
    {
        "access_token": "ya29.ImCpB6BS2mdOMseaUjhVlHqNfAOz168XjuDrK7Sd33glPd7XvtMLIngi1-V52ReytFSUluE-iBV88OlDkjtraggB_qc-LN2JlGtQ3sHZq_MuTxrU0-oK_kpq-1wsvniFFGQ",
        "expires_in": 3600,
        "scope": "openid https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/userinfo.email",
        "token_type": "Bearer",
        "id_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6IjhhNjNmZTcxZTUzMDY3NTI0Y2JiYzZhM2E1ODQ2M2IzODY0YzA3ODciLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20iLCJhenAiOiI3NjQwODYwNTE4NTAtNnFyNHA2Z3BpNmhuNTA2cHQ4ZWp1cTgzZGkzNDFodXIuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJhdWQiOiI3NjQwODYwNTE4NTAtNnFyNHA2Z3BpNmhuNTA2cHQ4ZWp1cTgzZGkzNDFodXIuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJzdWIiOiIxMDM5ODg0MzM2OTY3NzI1NDkzNjAiLCJoZCI6ImdvLWplay5jb20iLCJlbWFpbCI6InByYWRpdGh5YS5wdXJhQGdvLWplay5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiYXRfaGFzaCI6ImdrbXIxY0dPTzNsT0dZUDhtYjNJRnciLCJpYXQiOjE1NzE5Njg3NDUsImV4cCI6MTU3MTk3MjM0NX0.FIY5xvySNVxt1cbw-QXdDfiwollxcqupz1YDJuP14obKRyDwFG9ZcC_j-mTDZF5_dzpYeNMMK-LPTq9QIaM-blSKm2Eh9LeMvQGUk_S-9y_r2jKCmBlrEeHM8DXk3xyKf65LEoBA8cwMPdgb2s8AMIxxN9JJ09fjou20yLDI84Q4BFMriMIBBYLFgBW0wcg2PQ1hy5hrV1PdZj-ZNKNWmouh0lOjLLYmVFZPCzD9ENWo1N52ZLaLODdI2gDcpbyTUbeAh81sacdtJd0pLf-FuBLdfuktvP4MVvdmIhXv98Zb0dFBzRtmiqlQusSjoG5VEaBc6o2gkM5rHR0ozby0Fg"
    }
    """, status=200, content_type='application/json')
