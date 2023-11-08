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

import http
import json
import os
import subprocess
import time
from abc import abstractmethod

import mlflow.pyfunc
import pytest
import requests
import tornado.web
from prometheus_client import Counter, Gauge

from pyfuncserver.config import HTTP_PORT, ModelManifest, WORKERS
from pyfuncserver.model.model import PyFuncModel
from test.utils import wait_server_ready
from merlin.pyfunc import PYFUNC_MODEL_INPUT_KEY, PYFUNC_EXTRA_ARGS_KEY, PyFuncOutput


class NewModelImpl(mlflow.pyfunc.PythonModel):
    def load_context(self, context):
        self.initialize(context.artifacts)
        self._use_kwargs_infer = True

    def predict(self, context, model_input):
        extra_args = model_input.get(PYFUNC_EXTRA_ARGS_KEY, {})
        input = model_input.get(PYFUNC_MODEL_INPUT_KEY, {})
        if extra_args is not None:
            response = self._do_predict(input, **extra_args)
            return PyFuncOutput(http_response=response)

        response = self._do_predict(input)
        return PyFuncOutput(http_response=response)

    def _do_predict(self, model_input, **kwargs):
        if self._use_kwargs_infer:
            try:
                return self.infer(model_input, **kwargs)
            except TypeError as e:
                if "infer() got an unexpected keyword argument" in str(e):
                    print(
                        'Fallback to the old infer() method, got TypeError exception: {}'.format(e))
                    self._use_kwargs_infer = False
                else:
                    raise e

        return self.infer(model_input)

    @abstractmethod
    def initialize(self, artifacts: dict):
        """
        Implementation of PyFuncModel can specify initialization step which
        will be called one time during model initialization.
        :param artifacts: dictionary of artifacts passed to log_model method
        """
        pass

    @abstractmethod
    def infer(self, request: dict, **kwargs) -> dict:
        """
        Do inference
        This method MUST be implemented by concrete implementation of
        PyFuncModel.
        This method accept 'request' which is the body content of incoming
        request.
        Implementation should return inference a json object of response.
        :param request: Dictionary containing incoming request body content
        :return: Dictionary containing response body
        """
        pass


class ModelImpl(mlflow.pyfunc.PythonModel):
    def load_context(self, context):
        self.initialize(context.artifacts)
        self._use_kwargs_infer = True

    def predict(self, model_input, **kwargs):
        if self._use_kwargs_infer:
            try:
                response = self.infer(model_input, **kwargs)
                return PyFuncOutput(http_response=response)
            except TypeError as e:
                if "infer() got an unexpected keyword argument" in str(e):
                    print(
                        'Fallback to the old infer() method, got TypeError exception: {}'.format(e))
                    self._use_kwargs_infer = False
                else:
                    raise e

        response = self.infer(model_input)
        return PyFuncOutput(http_response=response)

    @abstractmethod
    def initialize(self, artifacts: dict):
        """
        Implementation of PyFuncModel can specify initialization step which
        will be called one time during model initialization.
        :param artifacts: dictionary of artifacts passed to log_model method
        """
        pass

    @abstractmethod
    def infer(self, request: dict, **kwargs) -> dict:
        """
        Do inference
        This method MUST be implemented by concrete implementation of
        PyFuncModel.
        This method accept 'request' which is the body content of incoming
        request.
        Implementation should return inference a json object of response.
        :param request: Dictionary containing incoming request body content
        :return: Dictionary containing response body
        """
        pass


class EchoModel(ModelImpl):
    def initialize(self, artifacts):
        self._req_count = Counter("my_req_count", "Number of incoming request")
        self._temp = Gauge("my_gauge", "Number of incoming request")

    def infer(self, model_input):
        self._req_count.inc()
        self._temp.set(10)
        return model_input


class NewEchoModel(NewModelImpl):
    def initialize(self, artifacts):
        self._req_count = Counter("new_my_req_count", "Number of incoming request")
        self._temp = Gauge("new_my_gauge", "Number of incoming request")

    def infer(self, model_input):
        self._req_count.inc()
        self._temp.set(10)
        return model_input


class NonEmptyEchoModel(NewModelImpl):
    def initialize(self, artifacts):
        self._req_count = Counter("non_empty_my_req_count", "Number of incoming request")
        self._temp = Gauge("non_empty_my_gauge", "Number of incoming request")

    def infer(self, model_input):
        if model_input == {}:
            raise TypeError("Request could not be empty")
        self._req_count.inc()
        self._temp.set(10)
        return model_input


class HeadersModel(ModelImpl):
    def initialize(self, artifacts):
        pass

    def infer(self, model_input, **kwargs):
        return kwargs.get('headers', {})


class NewHeadersModel(NewModelImpl):
    def initialize(self, artifacts):
        pass

    def infer(self, model_input, **kwargs):
        return kwargs.get('headers', {})


class HttpErrorModel(ModelImpl):
    def initialize(self, artifacts):
        pass

    def infer(self, model_input):
        raise tornado.web.HTTPError(
            status_code=model_input["status_code"],
            reason=model_input["reason"]
        )


class NewHttpErrorModel(NewModelImpl):
    def initialize(self, artifacts):
        pass

    def infer(self, model_input):
        raise tornado.web.HTTPError(
            status_code=model_input["status_code"],
            reason=model_input["reason"]
        )

@pytest.mark.parametrize(
    "model",
    [EchoModel(), NewEchoModel(), NonEmptyEchoModel()]
)
def test_model(model):
    mlflow.pyfunc.log_model("model", python_model=model)
    model_path = os.path.join(mlflow.get_artifact_uri(), "model")
    model_manifest = ModelManifest(model_name="echo-model",
                                   model_version="1",
                                   model_full_name="echo-model-1",
                                   model_dir=model_path,
                                   project="sample")

    model = PyFuncModel(model_manifest)
    model.load()

    assert model.ready

    inputs = [[1, 2, 3], [4, 5, 6]]
    outputs = model.predict(inputs)
    assert outputs == PyFuncOutput(http_response=inputs)

@pytest.mark.parametrize(
    "model",
    [EchoModel(), NewEchoModel(), NonEmptyEchoModel()]
)
def test_model_int(model):
    mlflow.pyfunc.log_model("model", python_model=model)
    model_path = os.path.join(mlflow.get_artifact_uri(), "model")
    env = os.environ.copy()
    mlflow.end_run()

    try:
        env["PROMETHEUS_MULTIPROC_DIR"] = "prometheus"
        env[HTTP_PORT[0]] = "8081"
        env[WORKERS[0]] = "1"
        c = subprocess.Popen(["python", "-m", "pyfuncserver", "--model_dir", model_path], env=env)

        # wait till the server is up
        wait_server_ready("http://localhost:8081/")

        for request_file_json in os.listdir("benchmark"):
            if not request_file_json.endswith(".json"):
                continue

            with open(os.path.join("benchmark", request_file_json), "r") as f:
                req = json.load(f)

            resp = requests.post("http://localhost:8081/v1/models/model-1:predict",
                                 json=req)
            assert resp.status_code == 200
            assert req == resp.json()
            assert resp.headers["content-type"] == "application/json; charset=UTF-8"
    finally:
        c.kill()


@pytest.mark.parametrize(
    "model",
    [HeadersModel(), NewHeadersModel()]
)
def test_model_headers(model):
    mlflow.pyfunc.log_model("model", python_model=model)
    model_path = os.path.join(mlflow.get_artifact_uri(), "model")
    env = os.environ.copy()
    mlflow.end_run()

    try:
        env["PROMETHEUS_MULTIPROC_DIR"] = "prometheus"
        env[HTTP_PORT[0]] = "8081"
        env[WORKERS[0]] = "1"
        c = subprocess.Popen(["python", "-m", "pyfuncserver", "--model_dir", model_path], env=env)

        # wait till the server is up
        wait_server_ready("http://localhost:8081/")

        with open(os.path.join("benchmark", "small.json"), "r") as f:
            req = json.load(f)

        resp = requests.post("http://localhost:8081/v1/models/model-1:predict",
                             json=req, headers={'Foo': 'bar'})
        assert resp.status_code == 200
        assert resp.json()["Foo"] == "bar"
        assert resp.headers["content-type"] == "application/json; charset=UTF-8"
    finally:
        c.kill()


@pytest.mark.parametrize(
    "error_core,message,model",
    [
        (http.HTTPStatus.BAD_REQUEST, "invalid request", HttpErrorModel()),
        (http.HTTPStatus.NOT_FOUND, "not found", HttpErrorModel()),
        (http.HTTPStatus.INTERNAL_SERVER_ERROR, "server error", HttpErrorModel()),
(http.HTTPStatus.BAD_REQUEST, "invalid request", NewHttpErrorModel()),
        (http.HTTPStatus.NOT_FOUND, "not found", NewHttpErrorModel()),
        (http.HTTPStatus.INTERNAL_SERVER_ERROR, "server error", NewHttpErrorModel()),
    ]
)
def test_error_model_int(error_core, message, model):
    mlflow.pyfunc.log_model("model", python_model=model)
    model_path = os.path.join(mlflow.get_artifact_uri(), "model")
    env = os.environ.copy()
    mlflow.end_run()

    try:
        env["PROMETHEUS_MULTIPROC_DIR"] = "prometheus"
        env[HTTP_PORT[0]] = "8081"
        env[WORKERS[0]] = "1"
        c = subprocess.Popen(["python", "-m", "pyfuncserver", "--model_dir", model_path], env=env)

        # wait till the server is up
        wait_server_ready("http://localhost:8081/")
        
        req = {"status_code": error_core.value, "reason": message}
        resp = requests.post("http://localhost:8081/v1/models/model-1:predict",
                             json=req)
        assert resp.status_code == error_core.value
        assert resp.json() == {"status_code": error_core.value, "reason": message}
        assert resp.headers["content-type"] == "application/json; charset=UTF-8"
    finally:
        c.kill()


