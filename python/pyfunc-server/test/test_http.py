import json
import os
import pathlib
import re
import shutil
import signal
import subprocess
import random

import mlflow
import requests
from merlin.model import PyFuncModel, PyFuncV3Model
from merlin.pyfunc import ModelInput, ModelOutput, Values
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
sample_model_output = ModelOutput(
            prediction_ids=["prediction_1", "prediction_2", "prediction3"],
            predictions=Values(
                columns=["prediction_score", "prediction_label"],
                data=[[0.9, "fraud"], [0.2, "not fraud"], [0.6, "fraud"]]
            )
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
    def ml_predict(self, model_input: ModelInput) -> ModelOutput:
        return sample_model_output
    def postprocess(self, model_output: ModelOutput, request: dict) -> dict:
        return model_output.predictions_dict()


class LoadTestModelV3(PyFuncV3Model):
    def preprocess(self, request: dict, **kwargs) -> ModelInput:
        features_data = request["features"]
        columns  = list(features_data[0].keys())
        data = []
        row_ids = []
        for feature in features_data:
            row_ids.append(feature["prediction_id"])
            row = []
            for name in feature:
                row.append(feature[name])
            data.append(row)
        model_input = ModelInput(
            features=Values(
                columns=columns,
                data = data
            ),
            prediction_ids = row_ids
        )
        return model_input
    
    def ml_predict(self, model_input: ModelInput) -> ModelOutput:
        predictions = [random.uniform(0,1) for _ in range(len(model_input.prediction_ids))]
        data = []
        for p in predictions:
            data.append([p])
        
        return ModelOutput(
            predictions=Values(
                columns=["prediction_score"],
                data = data
            ),
            prediction_ids = model_input.prediction_ids
        )
    def postprocess(self, model_output: ModelOutput, request: dict) -> dict:
        return model_output.predictions_dict()

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

def test_http_pyfuncv3():
    mlflow.pyfunc.log_model("model", python_model=SamplePyFuncV3Model())
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

            # test readiness endpoint
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
