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

import mlflow
import os
import pytest

import client as cl
from client import ApiClient, Configuration
from merlin.model import Model, ModelType, ModelVersion, Project


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
    r = mlflow.start_run()
    mlflow.end_run()
    v = cl.Version(
        id=1,
        model_id=model.id,
        mlflow_run_id=r.info.run_id,
        mlflow_url=mlflow_url,
        artifact_uri="gs://artifacts",
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
