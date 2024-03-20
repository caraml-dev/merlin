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
from typing import Union

import joblib
import mlflow
import numpy as np
import pandas as pd
import pytest
from merlin.model import PyFuncV2Model
from merlinpyspark.config import ModelConfig
from merlinpyspark.model import create_model_udf
from merlinpyspark.spec.prediction_job_pb2 import Model, ModelType, ResultType
from mlflow.pyfunc import log_model
from pandas.testing import assert_frame_equal
from sklearn.datasets import load_iris


class IrisModel(PyFuncV2Model):
    def initialize(self, artifacts: dict):
        self._model = joblib.load(artifacts["model_path"])

    def infer(
        self, model_input: pd.DataFrame
    ) -> Union[np.ndarray, pd.Series, pd.DataFrame]:
        return self._model.predict(model_input)


@pytest.mark.ci
def test_pyfuncv2_model(spark_session):
    log_model(
        "model",
        python_model=IrisModel(),
        artifacts={"model_path": "test-model/artifacts/model.joblib"},
        code_path=["merlinpyspark", "test"],
        conda_env="test-model/conda.yaml",
    )
    model_path = os.path.join(mlflow.get_artifact_uri(), "model")

    iris = load_iris()
    columns = ["sepal_length", "sepal_width", "petal_length", "petal_width"]
    pdf = pd.DataFrame(iris.data, columns=columns)
    df = spark_session.createDataFrame(pdf)

    model_cfg_proto = Model(
        type=ModelType.PYFUNC_V2,
        uri=model_path,
        result=Model.ModelResult(type=ResultType.DOUBLE),
    )

    model_udf = create_model_udf(spark_session, ModelConfig(model_cfg_proto), columns)
    iris_model = mlflow.pyfunc.load_model(model_path)

    df = df.withColumn(
        "prediction",
        model_udf("sepal_length", "sepal_width", "petal_length", "petal_width"),
    )

    result_pandas = df.toPandas()
    r = result_pandas.drop(columns=["prediction"])

    assert_frame_equal(r, pdf)
    assert np.array_equal(result_pandas["prediction"], iris_model.predict(pdf))
    assert len(df.columns) == len(columns) + 1
