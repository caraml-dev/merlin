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
from time import sleep
from typing import Union

import joblib
import numpy as np
import pandas as pd
import pytest
from sklearn import svm
from sklearn.datasets import load_iris

import merlin
from merlin.batch.config import PredictionJobConfig
from merlin.batch.job import JobStatus
from merlin.batch.sink import BigQuerySink, SaveMode
from merlin.batch.source import BigQuerySource
from merlin.model import ModelType, PyFuncV2Model
from merlin.resource_request import ResourceRequest

MODEL_PATH_ARTIFACT_KEY = "model_path"
MODEL_DIR = "test/batch/model"
MODEL_FILE = "model.joblib"
MODEL_PATH = os.path.join(MODEL_DIR, MODEL_FILE)
ENV_PATH = os.path.join(MODEL_DIR, "env.yaml")


class IrisClassifier(PyFuncV2Model):
    def initialize(self, artifacts: dict):
        self._model = joblib.load(artifacts[MODEL_PATH_ARTIFACT_KEY])

    def infer(
        self, model_input: pd.DataFrame
    ) -> Union[np.ndarray, pd.Series, pd.DataFrame]:
        return self._model.predict(model_input)


@pytest.fixture
def batch_bigquery_source():
    return os.environ.get("E2E_BATCH_BIGQUERY_SOURCE", default="project.dataset.table")


@pytest.fixture
def batch_bigquery_sink():
    return os.environ.get(
        "E2E_BATCH_BIGQUERY_SINK", default="project.dataset.table_result"
    )


@pytest.fixture
def batch_gcs_staging_bucket():
    return os.environ.get("E2E_BATCH_GCS_STAGING_BUCKET", default="bucket-name")


@pytest.mark.batch
@pytest.mark.integration
def test_batch_pyfunc_v2_batch(
    integration_test_url,
    project_name,
    service_account,
    use_google_oauth,
    batch_bigquery_source,
    batch_bigquery_sink,
    batch_gcs_staging_bucket,
):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("batch-iris", ModelType.PYFUNC_V2)
    service_account_name = "merlin-integration-test@project.iam.gserviceaccount.com"
    _create_secret(merlin.active_project(), service_account_name, service_account)

    clf = svm.SVC(gamma="scale")
    iris = load_iris()
    X, y = iris.data, iris.target
    clf.fit(X, y)
    joblib.dump(clf, MODEL_PATH)
    # Create new version of the model
    mdl = merlin.active_model()
    v = mdl.new_model_version()
    v.start()
    # Upload the serialized model to MLP
    v.log_pyfunc_model(
        model_instance=IrisClassifier(),
        conda_env=ENV_PATH,
        code_dir=["test"],
        artifacts={MODEL_PATH_ARTIFACT_KEY: MODEL_PATH},
    )

    v.finish()

    bq_source = BigQuerySource(
        batch_bigquery_source,
        features=["sepal_length", "sepal_width", "petal_length", "petal_width"],
    )
    bq_sink = BigQuerySink(
        batch_bigquery_sink,
        staging_bucket=batch_gcs_staging_bucket,
        result_column="prediction",
        save_mode=SaveMode.OVERWRITE,
    )

    image_builder_resource_request = ResourceRequest(
        cpu_request="2", memory_request="4Gi"
    )
    job_config = PredictionJobConfig(
        source=bq_source,
        sink=bq_sink,
        service_account_name=service_account_name,
        env_vars={"ALPHA": "0.2"},
        image_builder_resource_request=image_builder_resource_request,
    )

    job = v.create_prediction_job(job_config=job_config)

    assert job.status == JobStatus.COMPLETED

    job = v.create_prediction_job(job_config=job_config, sync=False)
    while job.status == JobStatus.PENDING:
        sleep(1)
        job = job.refresh()
    job = job.stop()

    assert job.status == JobStatus.TERMINATED


def _create_secret(project, secret_name, secret_value):
    secrets = project.list_secret()
    for secret in secrets:
        if secret == secret_name:
            return

    if secret_value is None:
        raise ValueError(
            f"{secret_name} secret is not found in the project and E2E_SERVICE_ACCOUNT is empty"
        )
    project.create_secret(secret_name, secret_value)
