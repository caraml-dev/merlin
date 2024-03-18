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
from test.model_test import IrisModel

import mlflow
import pytest
from merlinpyspark.__main__ import start
from merlinpyspark.config import load
from mlflow.pyfunc import log_model

test_config = """
kind: PredictionJob
version: v1
name: integration-test
bigquerySource:
  table: 'project.dataset.table_iris'
  features:
    - sepal_length
    - sepal_width
    - petal_length
    - petal_width
model:
  type: PYFUNC_V2
  uri: {}
  result:
    type: DOUBLE
bigquerySink:
  table: 'project.dataset.table_iris_result'
  result_column: 'prediction'
  save_mode: OVERWRITE
  options:
    project: "project"
    temporaryGcsBucket: "bucket-name"
"""


@pytest.mark.ci
def test_integration(spark_session, bq):
    log_model(
        "model",
        python_model=IrisModel(),
        artifacts={"model_path": "test-model/artifacts/model.joblib"},
    )
    model_path = os.path.join(mlflow.get_artifact_uri(), "model")

    config_path = "test-config/integration_test.yaml"
    with open(config_path, "w") as f:
        f.write(test_config.format(model_path))

    cfg = load(config_path)

    bq.delete_table(table=cfg.sink().table(), not_found_ok=True)

    start(config_path, spark_session)

    result_table = bq.get_table(cfg.sink().table())
    assert len(result_table.schema) == 5
    assert result_table.schema[4].name == cfg.sink().result_column()
    assert result_table.schema[4].field_type == "FLOAT"
