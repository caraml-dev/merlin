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

import os

import mlflow
import pytest
import uuid
import requests as requests_lib
from requests.adapters import HTTPAdapter
from requests.packages.urllib3.util.retry import Retry
from urllib3_mock import Responses

import client as cl
from client import ApiClient, Configuration
from merlin.model import Model, ModelType, ModelVersion, Project


# From the documentation (https://docs.pytest.org/en/7.1.x/reference/reference.html#pytest.hookspec.pytest_configure):
# Allow plugins and conftest files to perform initial configuration.
# This hook is called for every plugin and initial conftest file after command line options have been parsed.
# After that, the hook is called for other conftest files as they are imported.
def pytest_configure():
    # share mock responses as global variable so it can be reused and called as decorator on other files.
    pytest.responses = Responses("requests.packages.urllib3")


@pytest.fixture
def url():
    return "http://127.0.0.1:8080"


@pytest.fixture
def api_client(url):
    config = Configuration()
    config.host = url + "/v1"
    return ApiClient(config)


@pytest.fixture
def mlflow_url():
    return "mlflow"


@pytest.fixture
def project(url, mlflow_url, api_client):
    prj = cl.Project(
        id=1,
        name="my-project",
        mlflow_tracking_url=mlflow_url,
        created_at="2019-08-29T08:13:12.377Z",
        updated_at="2019-08-29T08:13:12.377Z",
    )
    return Project(prj, url, api_client)


@pytest.fixture
def model(project, mlflow_url, api_client):
    mdl = cl.Model(
        id=1,
        project_id=project.id,
        mlflow_experiment_id=1,
        name="my-model",
        type=ModelType.PYFUNC_V2.value,
        mlflow_url=mlflow_url,
        created_at="2019-08-29T08:13:12.377Z",
        updated_at="2019-08-29T08:13:12.377Z",
    )
    return Model(mdl, project, api_client)


@pytest.fixture
def version(project, model, mlflow_url, api_client):
    mlflow.set_tracking_uri(mlflow_url)
    run_id = mlflow.create_experiment(
        name=f"unit_test_experiment_{uuid.uuid4()}"
    )
    r = mlflow.start_run(experiment_id=run_id)
    mlflow.end_run()
    v = cl.Version(
        id=1,
        model_id=model.id,
        mlflow_run_id=r.info.run_id,
        mlflow_url=mlflow_url,
        artifact_uri="gs://artifacts",
        python_version="3.7.*",
        created_at="2019-08-29T08:13:12.377Z",
        updated_at="2019-08-29T08:13:12.377Z",
    )
    return ModelVersion(v, model, api_client)


@pytest.fixture
def integration_test_url():
    return os.environ.get("E2E_MERLIN_URL", default="http://127.0.0.1:8080")


@pytest.fixture
def feast_serving_redis_url():
    return os.environ.get(
        "E2E_FEAST_SERVING_REDIS_URL", default="http://127.0.0.1:6566"
    )


@pytest.fixture
def feast_serving_bigtable_url():
    return os.environ.get(
        "E2E_FEAST_SERVING_BIGTABLE_URL", default="http://127.0.0.1:6567"
    )


@pytest.fixture
def project_name():
    return os.environ.get("E2E_PROJECT_NAME", default="integration-test")


@pytest.fixture
def service_account():
    return os.environ.get("E2E_SERVICE_ACCOUNT", default=None)


@pytest.fixture
def use_google_oauth():
    return os.environ.get("E2E_USE_GOOGLE_OAUTH", default=True) == "true"


@pytest.fixture
def requests():
    retry_strategy = Retry(
        total=5,
        status_forcelist=[429, 500, 502, 503, 504, 404],
        allowed_methods=["POST"],
        backoff_factor=0.5,
    )
    adapter = HTTPAdapter(max_retries=retry_strategy)
    req = requests_lib.Session()
    req.mount("http://", adapter)

    return req

@pytest.fixture
def gpu_config():
    gpu_request = os.environ.get("E2E_GPU_REQUEST", default="1")
    gpu_name = os.environ.get("E2E_GPU_NAME", default="nvidia-tesla-p4")

    return {"gpu_request": gpu_request, "gpu_name": gpu_name}


@pytest.fixture
def mock_oauth():
    responses = pytest.responses

    responses.add(
        "GET",
        "/computeMetadata/v1/instance/service-accounts/default/",
        body="""
    {
        "email": "computeengine@google.com",
        "scopes": ["scope"],
        "aliases": ["default"]
    }
    """,
        status=200,
        content_type="application/json",
    )

    responses.add(
        "GET",
        "/computeMetadata/v1/instance/service-accounts/computeengine@google.com/token",
        body="""
    {
        "access_token": "ya29.ImCpB6BS2mdOMseaUjhVlHqNfAOz168XjuDrK7Sd33glPd7XvtMLIngi1-V52ReytFSUluE-iBV88OlDkjtraggB_qc-LN2JlGtQ3sHZq_MuTxrU0-oK_kpq-1wsvniFFGQ",
        "expires_in": 3600,
        "scope": "openid https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/userinfo.email",
        "token_type": "Bearer",
        "id_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6IjhhNjNmZTcxZTUzMDY3NTI0Y2JiYzZhM2E1ODQ2M2IzODY0YzA3ODciLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20iLCJhenAiOiI3NjQwODYwNTE4NTAtNnFyNHA2Z3BpNmhuNTA2cHQ4ZWp1cTgzZGkzNDFodXIuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJhdWQiOiI3NjQwODYwNTE4NTAtNnFyNHA2Z3BpNmhuNTA2cHQ4ZWp1cTgzZGkzNDFodXIuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJzdWIiOiIxMDM5ODg0MzM2OTY3NzI1NDkzNjAiLCJoZCI6ImdvLWplay5jb20iLCJlbWFpbCI6InByYWRpdGh5YS5wdXJhQGdvLWplay5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiYXRfaGFzaCI6ImdrbXIxY0dPTzNsT0dZUDhtYjNJRnciLCJpYXQiOjE1NzE5Njg3NDUsImV4cCI6MTU3MTk3MjM0NX0.FIY5xvySNVxt1cbw-QXdDfiwollxcqupz1YDJuP14obKRyDwFG9ZcC_j-mTDZF5_dzpYeNMMK-LPTq9QIaM-blSKm2Eh9LeMvQGUk_S-9y_r2jKCmBlrEeHM8DXk3xyKf65LEoBA8cwMPdgb2s8AMIxxN9JJ09fjou20yLDI84Q4BFMriMIBBYLFgBW0wcg2PQ1hy5hrV1PdZj-ZNKNWmouh0lOjLLYmVFZPCzD9ENWo1N52ZLaLODdI2gDcpbyTUbeAh81sacdtJd0pLf-FuBLdfuktvP4MVvdmIhXv98Zb0dFBzRtmiqlQusSjoG5VEaBc6o2gkM5rHR0ozby0Fg"
    }
    """,
        status=200,
        content_type="application/json",
    )

    responses.add(
        "GET",
        "/computeMetadata/v1/instance/service-accounts/default/?recursive=true",
        body="""
    {
        "aliases":["default"],
        "email":"merlin@caraml.dev",
        "scopes":["https://www.googleapis.com/auth/cloud-platform","https://www.googleapis.com/auth/userinfo.email"]
    }
    """,
        status=200,
        content_type="application/json",
    )

    responses.add(
        "GET",
        "/computeMetadata/v1/instance/service-accounts/default/identity?audience=sdk.caraml&format=full",
        body="""eyJhbGciOiJSUzI1NiIsImtpZCI6IjFlOWdkazcifQ.ewogImlzcyI6ICJodHRwOi8vc2VydmVyLmV4YW1wbGUuY29tIiwKICJzdWIiOiAiMjQ4Mjg5NzYxMDAxIiwKICJhdWQiOiAiczZCaGRSa3F0MyIsCiAibm9uY2UiOiAibi0wUzZfV3pBMk1qIiwKICJleHAiOiAxMzExMjgxOTcwLAogImlhdCI6IDEzMTEyODA5NzAKfQ.ggW8hZ1EuVLuxNuuIJKX_V8a_OMXzR0EHR9R6jgdqrOOF4daGU96Sr_P6qJp6IcmD3HP99Obi1PRs-cwh3LO-p146waJ8IhehcwL7F09JdijmBqkvPeB2T9CJNqeGpe-gccMg4vfKjkM8FcGvnzZUN4_KSP0aAp1tOJ1zZwgjxqGByKHiOtX7TpdQyHE5lcMiKPXfEIQILVq0pc_E2DzL7emopWoaoZTF_m0_N0YzFC6g6EJbOEoRoSK5hoDalrcvRYLSrQAZZKflyuVCyixEoV9GfNQC3_osjzw2PAithfubEEBLuVVk4XUVrWOLrLl0nx7RkKU8NXNHq-rvKMzqg""",
        status=200,
        content_type="application/text",
        match_querystring=True,
    )

    responses.add(
        "GET",
        "/computeMetadata/api/v1/instance/service-accounts/default/",
        body="""
    {
        "email": "computeengine@google.com",
        "scopes": ["scope"],
        "aliases": ["default"]
    }
    """,
        status=200,
        content_type="application/json",
    )

    responses.add(
        "GET",
        "/computeMetadata/api/v1/instance/service-accounts/computeengine@google.com/token",
        body="""
    {
        "access_token": "thisisnotarealtoken",
        "expires_in": 3600,
        "scope": "openid https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/userinfo.email",
        "token_type": "Bearer",
        "id_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6IjFlOWdkazcifQ.ewogImlzcyI6ICJodHRwOi8vc2VydmVyLmV4YW1wbGUuY29tIiwKICJzdWIiOiAiMjQ4Mjg5NzYxMDAxIiwKICJhdWQiOiAiczZCaGRSa3F0MyIsCiAibm9uY2UiOiAibi0wUzZfV3pBMk1qIiwKICJleHAiOiAxMzExMjgxOTcwLAogImlhdCI6IDEzMTEyODA5NzAKfQ.ggW8hZ1EuVLuxNuuIJKX_V8a_OMXzR0EHR9R6jgdqrOOF4daGU96Sr_P6qJp6IcmD3HP99Obi1PRs-cwh3LO-p146waJ8IhehcwL7F09JdijmBqkvPeB2T9CJNqeGpe-gccMg4vfKjkM8FcGvnzZUN4_KSP0aAp1tOJ1zZwgjxqGByKHiOtX7TpdQyHE5lcMiKPXfEIQILVq0pc_E2DzL7emopWoaoZTF_m0_N0YzFC6g6EJbOEoRoSK5hoDalrcvRYLSrQAZZKflyuVCyixEoV9GfNQC3_osjzw2PAithfubEEBLuVVk4XUVrWOLrLl0nx7RkKU8NXNHq-rvKMzqg"
    }
    """,
        status=200,
        content_type="application/json",
    )

    responses.add(
        "POST",
        "/token",
        body="""
    {
        "access_token": "thisisnotarealtoken",
        "expires_in": 3600,
        "scope": "openid https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/userinfo.email",
        "token_type": "Bearer",
        "id_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6IjFlOWdkazcifQ.ewogImlzcyI6ICJodHRwOi8vc2VydmVyLmV4YW1wbGUuY29tIiwKICJzdWIiOiAiMjQ4Mjg5NzYxMDAxIiwKICJhdWQiOiAiczZCaGRSa3F0MyIsCiAibm9uY2UiOiAibi0wUzZfV3pBMk1qIiwKICJleHAiOiAxMzExMjgxOTcwLAogImlhdCI6IDEzMTEyODA5NzAKfQ.ggW8hZ1EuVLuxNuuIJKX_V8a_OMXzR0EHR9R6jgdqrOOF4daGU96Sr_P6qJp6IcmD3HP99Obi1PRs-cwh3LO-p146waJ8IhehcwL7F09JdijmBqkvPeB2T9CJNqeGpe-gccMg4vfKjkM8FcGvnzZUN4_KSP0aAp1tOJ1zZwgjxqGByKHiOtX7TpdQyHE5lcMiKPXfEIQILVq0pc_E2DzL7emopWoaoZTF_m0_N0YzFC6g6EJbOEoRoSK5hoDalrcvRYLSrQAZZKflyuVCyixEoV9GfNQC3_osjzw2PAithfubEEBLuVVk4XUVrWOLrLl0nx7RkKU8NXNHq-rvKMzqg"
    }
    """,
        status=200,
        content_type="application/json",
    )
