import json
import pytest
import os
import pathlib
import re
import shutil
import signal
import subprocess
import random
import pandas as pd
import numpy as np

import mlflow
import requests
from merlin.model import PyFuncModel, PyFuncV3Model
from merlin.pyfunc import ModelInput, ModelOutput, Values, PyFuncOutput
from prometheus_client.metrics import Counter, Gauge

from pyfuncserver.config import HTTP_PORT, WORKERS, MODEL_NAME, MODEL_VERSION, MODEL_FULL_NAME
from test.utils import wait_server_ready


sample_model_input = ModelInput(
            prediction_ids=["prediction_1", "prediction_2", "prediction_3"],
            features=Values(
                columns=["featureA", "featureB", "featureC"],
                data=[[0.1, 0.2, 0.3], [0.2, 0.4, 0.6], [0.8, 0.5, 0.2]]
            ),
            entities=Values(
                columns=["order_id", "customer_id"],
                data=[["order1", "1111"], ["order1", "2222"], ["order1", "3333"]]
            ),
            session_id="randomSessionID"
        )

sample_model_input_np = ModelInput(
            prediction_ids=["prediction_1", "prediction_2", "prediction_3"],
            features=Values(
                columns=["featureA", "featureB", "featureC"],
                data=np.array([[0.1, 0.2, 0.3], [0.2, 0.4, 0.6], [0.8, 0.5, 0.2]])
            ),
            entities=Values(
                columns=["order_id", "customer_id"],
                data=np.array([["order1", "1111"], ["order1", "2222"], ["order1", "3333"]])
            ),
            session_id="randomSessionID"
        )

sample_model_input_df = ModelInput(
            prediction_ids=["prediction_1", "prediction_2", "prediction_3"],
            features = pd.DataFrame([[0.1, 0.2, 0.3], [0.2, 0.4, 0.6], [0.8, 0.5, 0.2]], columns=["featureA", "featureB", "featureC"]),
            entities = pd.DataFrame([["order1", "1111"], ["order1", "2222"], ["order1", "3333"]], columns=["order_id", "customer_id"]),
            session_id="randomSessionID"
        )
sample_model_output = ModelOutput(
            prediction_ids=["prediction_1", "prediction_2", "prediction3"],
            predictions=Values(
                columns=["prediction_score", "prediction_label"],
                data=[[0.9, "fraud"], [0.2, "not fraud"], [0.6, "fraud"]]
            )
        )
sample_model_output_np = ModelOutput(
            prediction_ids=["prediction_1", "prediction_2", "prediction3"],
            predictions=Values(
                columns=["prediction_score", "prediction_label"],
                data=np.array([[0.9, "fraud"], [0.2, "not fraud"], [0.6, "fraud"]])
            )
        )
sample_model_output_df = ModelOutput(
            prediction_ids=["prediction_1", "prediction_2", "prediction3"],
            predictions = pd.DataFrame([[0.9, "fraud"], [0.2, "not fraud"], [0.6, "fraud"]], columns=["prediction_score", "prediction_label"])
        )



class EchoModel(PyFuncModel):
    GAUGE_VALUE = 5

    def initialize(self, artifacts):
        self._req_count = Counter("request_count", "Number of incoming request")
        self._temp = Gauge("some_gauge", "Number of incoming request")

    def infer(self, model_input, **kwargs):
        if model_input == {}:
            raise TypeError("Request could not be empty")
        self._req_count.inc()
        self._temp.set(EchoModel.GAUGE_VALUE)
        return model_input
    
class SamplePyFuncV3Model(PyFuncV3Model):

    def preprocess(self, request: dict, **kwargs) -> ModelInput:
        return sample_model_input
    def infer(self, model_input: ModelInput) -> ModelOutput:
        return sample_model_output
    def postprocess(self, model_output: ModelOutput, request: dict) -> dict:
        return model_output.predictions_dict()
    
class SamplePyFuncV3NumpyModel(PyFuncV3Model):
    def preprocess(self, request: dict, **kwargs) -> ModelInput:
        return sample_model_input_np
    def infer(self, model_input: ModelInput) -> ModelOutput:
        return sample_model_output_np
    def postprocess(self, model_output: ModelOutput, request: dict) -> dict:
        return model_output.predictions_dict()
    
class SamplePyFuncV3DFModel(PyFuncV3Model):
    def preprocess(self, request: dict, **kwargs) -> ModelInput:
        return sample_model_input_df
    def infer(self, model_input: ModelInput) -> ModelOutput:
        return sample_model_output_df
    def postprocess(self, model_output: ModelOutput, request: dict) -> dict:
        return model_output.predictions_dict()


class SamplePyFuncWithModelObservability(PyFuncModel):
    def infer(self, request, **kwargs) -> ModelOutput:
        model_input = sample_model_input
        model_output = sample_model_output
        http_response = model_output.predictions_dict()
        return PyFuncOutput(http_response=http_response, model_input=model_input, model_output=model_output)
    

def test_http_protocol():
    mlflow.pyfunc.log_model("model", python_model=EchoModel())
    model_path = os.path.join(mlflow.get_artifact_uri(), "model")
    env = os.environ.copy()
    mlflow.end_run()

    model_name = "my-model"
    model_version = "2"
    model_full_name = f"{model_name}-{model_version}"
    port = "8081"

    try:
        env[HTTP_PORT[0]] = port
        env[MODEL_NAME[0]] = model_name
        env[MODEL_VERSION[0]] = model_version
        env[MODEL_FULL_NAME[0]] = model_full_name
        env[WORKERS[0]] = "1"
        env["PROMETHEUS_MULTIPROC_DIR"] = "prometheus"
        c = subprocess.Popen(["python", "-m", "pyfuncserver", "--model_dir", model_path], env=env)

        # wait till the server is up
        wait_server_ready(f"http://localhost:{port}/")

        for request_file_json in os.listdir("benchmark"):
            if not request_file_json.endswith(".json"):
                continue

            with open(os.path.join("benchmark", request_file_json), "r") as f:
                req = json.load(f)

            # test predict endpoint
            resp = requests.post(f"http://localhost:{port}/v1/models/{model_full_name}:predict",
                                 json=req)
            assert resp.status_code == 200
            assert req == resp.json()
            assert resp.headers["content-type"] == "application/json; charset=UTF-8"

            # test readiness endpoint
            resp = requests.get(f"http://localhost:{port}/")
            assert resp.status_code == 200

            # test readiness endpoint
            resp = requests.get(f"http://localhost:{port}/v1/models/{model_full_name}")
            assert resp.status_code == 200
    finally:
        c.kill()

@pytest.mark.parametrize("pyfunc_v3_model,expected", [
    (SamplePyFuncV3Model(), sample_model_output.predictions_dict()),
    (SamplePyFuncV3DFModel(), sample_model_output_df.predictions_dict()),
    (SamplePyFuncV3NumpyModel(), sample_model_output_np.predictions_dict())
])
def test_http_pyfuncv3(pyfunc_v3_model, expected):
    mlflow.pyfunc.log_model("model", python_model=pyfunc_v3_model)
    model_path = os.path.join(mlflow.get_artifact_uri(), "model")
    env = os.environ.copy()
    mlflow.end_run()

    model_name = "my-model"
    model_version = "2"
    model_full_name = f"{model_name}-{model_version}"
    port = "8081"

    try:
        env[HTTP_PORT[0]] = port
        env[MODEL_NAME[0]] = model_name
        env[MODEL_VERSION[0]] = model_version
        env[MODEL_FULL_NAME[0]] = model_full_name
        env[WORKERS[0]] = "1"
        env["PROMETHEUS_MULTIPROC_DIR"] = "prometheus"
        c = subprocess.Popen(["python", "-m", "pyfuncserver", "--model_dir", model_path], env=env)

        # wait till the server is up
        wait_server_ready(f"http://localhost:{port}/")

        for request_file_json in os.listdir("benchmark"):
            if not request_file_json.endswith(".json"):
                continue

            with open(os.path.join("benchmark", request_file_json), "r") as f:
                req = json.load(f)

            # test predict endpoint
            resp = requests.post(f"http://localhost:{port}/v1/models/{model_full_name}:predict",
                                 json=req)
            assert resp.status_code == 200
            assert resp.json() == expected
            assert resp.headers["content-type"] == "application/json; charset=UTF-8"

            # test readiness endpoint
            resp = requests.get(f"http://localhost:{port}/")
            assert resp.status_code == 200

            # test readiness endpoint
            resp = requests.get(f"http://localhost:{port}/v1/models/{model_full_name}")
            assert resp.status_code == 200
    finally:
        c.kill()

def test_http_pyfunc_with_model_observability():
    mlflow.pyfunc.log_model("model", python_model=SamplePyFuncWithModelObservability())
    model_path = os.path.join(mlflow.get_artifact_uri(), "model")
    env = os.environ.copy()
    mlflow.end_run()

    model_name = "my-model"
    model_version = "2"
    model_full_name = f"{model_name}-{model_version}"
    port = "8081"

    try:
        env[HTTP_PORT[0]] = port
        env[MODEL_NAME[0]] = model_name
        env[MODEL_VERSION[0]] = model_version
        env[MODEL_FULL_NAME[0]] = model_full_name
        env[WORKERS[0]] = "1"
        env["PROMETHEUS_MULTIPROC_DIR"] = "prometheus"
        c = subprocess.Popen(["python", "-m", "pyfuncserver", "--model_dir", model_path], env=env)

        # wait till the server is up
        wait_server_ready(f"http://localhost:{port}/")

        for request_file_json in os.listdir("benchmark"):
            if not request_file_json.endswith(".json"):
                continue

            with open(os.path.join("benchmark", request_file_json), "r") as f:
                req = json.load(f)

            # test predict endpoint
            resp = requests.post(f"http://localhost:{port}/v1/models/{model_full_name}:predict",
                                 json=req)
            assert resp.status_code == 200
            assert resp.json() == sample_model_output.predictions_dict()
            assert resp.headers["content-type"] == "application/json; charset=UTF-8"

            # test readiness endpoint``
            resp = requests.get(f"http://localhost:{port}/")
            assert resp.status_code == 200

            # test readiness endpoint
            resp = requests.get(f"http://localhost:{port}/v1/models/{model_full_name}")
            assert resp.status_code == 200
    finally:
        c.kill()


def test_metrics():
    mlflow.pyfunc.log_model("model", python_model=EchoModel())
    model_path = os.path.join(mlflow.get_artifact_uri(), "model")
    env = os.environ.copy()
    mlflow.end_run()

    model_name = "my-model"
    model_version = "1"
    model_full_name = f"{model_name}-{model_version}"
    port = "8082"
    metrics_path = "metrics_test"

    try:
        pathlib.Path(metrics_path).mkdir(exist_ok=True)

        env[HTTP_PORT[0]] = port
        env[MODEL_NAME[0]] = model_name
        env[MODEL_VERSION[0]] = model_version
        env[MODEL_FULL_NAME[0]] = model_full_name
        env[WORKERS[0]] = "4"
        env["PROMETHEUS_MULTIPROC_DIR"] = metrics_path
        c = subprocess.Popen(["python", "-m", "pyfuncserver", "--model_dir", model_path], env=env,
                             start_new_session=True)
        # wait till the server is up
        wait_server_ready(f"http://localhost:{port}/")

        predict_count = 0
        for request_file_json in os.listdir("benchmark"):
            if not request_file_json.endswith(".json"):
                continue

            with open(os.path.join("benchmark", request_file_json), "r") as f:
                req = json.load(f)

            # test predict endpoint
            resp = requests.post(f"http://localhost:{port}/v1/models/{model_full_name}:predict",
                                 json=req)
            assert resp.status_code == 200
            assert req == resp.json()
            assert resp.headers["content-type"] == "application/json; charset=UTF-8"
            predict_count += 1

        resp = requests.get(f"http://localhost:{port}/metrics")
        assert resp.status_code == 200

        # Check request_count counter
        matches = re.findall(r"request_count_total\s(\d\.\d)", resp.text)
        assert len(matches) == 1
        assert predict_count == int(float(matches[0]))

        # Check some_gauge gauge value
        matches = re.findall(r"some_gauge\{pid=\"\d+\"\}\s(\d+.\d+)", resp.text)
        assert len(matches) > 0
        for match in matches:
            gauge_value = int(float(match))
            assert gauge_value == EchoModel.GAUGE_VALUE or gauge_value == 0

    finally:
        os.killpg(os.getpgid(c.pid), signal.SIGTERM)
        shutil.rmtree(metrics_path)
