import json
import os
import subprocess

import mlflow
import requests
from merlin.model import PyFuncModel

from pyfuncserver.config import HTTP_PORT, WORKER, MODEL_NAME, MODEL_VERSION, MODEL_FULL_NAME
from test.utils import wait_server_ready


class EchoModel(PyFuncModel):
    def initialize(self, artifacts: dict):
        pass

    def infer(self, request: dict, **kwargs) -> dict:
        return request


def test_http_protocol():
    mlflow.pyfunc.log_model("model", python_model=EchoModel())
    model_path = os.path.join(mlflow.get_artifact_uri(), "model")
    env = os.environ.copy()
    mlflow.end_run()

    model_name = "my-model"
    model_version = "2"
    model_full_name = f"{model_name}-{model_version}"

    try:
        # use mlruns folder to store prometheus multiprocess files
        env["prometheus_multiproc_dir"] = "mlruns"
        env[HTTP_PORT] = "8081"
        env[MODEL_NAME] = model_name
        env[MODEL_VERSION] = model_version
        env[MODEL_FULL_NAME] = model_full_name
        env[WORKER] = "1"
        c = subprocess.Popen(["python", "-m", "pyfuncserver", "--model_dir", model_path], env=env)

        # wait till the server is up
        wait_server_ready("http://localhost:8081/")

        for request_file_json in os.listdir("benchmark"):
            if not request_file_json.endswith(".json"):
                continue

            with open(os.path.join("benchmark", request_file_json), "r") as f:
                req = json.load(f)

            # test predict endpoint
            resp = requests.post(f"http://localhost:8081/v1/models/{model_full_name}:predict",
                                 json=req)
            assert resp.status_code == 200
            assert req == resp.json()
            assert resp.headers["content-type"] == "application/json; charset=UTF-8"

            # test readiness endpoint
            resp = requests.get(f"http://localhost:8081/")
            assert resp.status_code == 200

            # test readiness endpoint
            resp = requests.get(f"http://localhost:8081/v1/models/{model_full_name}")
            assert resp.status_code == 200
    finally:
        c.kill()
