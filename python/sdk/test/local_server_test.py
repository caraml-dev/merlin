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
import random
import socket
import time
from multiprocessing import Process

import merlin
import pytest
import requests
from merlin.model import ModelType

request_json = {
    "instances": [
        [2.8, 1.0, 6.8, 0.4],
        [3.1, 1.4, 4.5, 1.6]
    ]
}

if os.environ.get("CI_SERVER"):
    host = "172.17.0.1"
else:
    host = "localhost"


@pytest.mark.integration
@pytest.mark.local_server_test
@pytest.mark.dependency(depends=["test/integration_test.py::test_sklearn"], scope='session')
def test_sklearn(integration_test_url, project_name, use_google_oauth):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("sklearn-sample", ModelType.SKLEARN)
    v = _get_latest_version(merlin.active_model())
    port = _get_free_port()
    p = Process(target=v.start_server, kwargs={"port": port, "build_image": True})
    p.start()
    _wait_server_ready(f"http://{host}:{port}")
    resp = requests.post(_get_local_endpoint(v, port), json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()['predictions']) == len(request_json['instances'])
    p.terminate()


@pytest.mark.integration
@pytest.mark.local_server_test
@pytest.mark.dependency(depends=["test/integration_test.py::test_xgboost"], scope='session')
def test_xgboost(integration_test_url, project_name, use_google_oauth):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("xgboost-sample", ModelType.XGBOOST)
    v = _get_latest_version(merlin.active_model())
    port = _get_free_port()
    p = Process(target=v.start_server, kwargs={"port": port, "build_image": True})
    p.start()
    _wait_server_ready(f"http://{host}:{port}")
    resp = requests.post(_get_local_endpoint(v, port), json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()['predictions']) == len(request_json['instances'])
    p.terminate()


@pytest.mark.integration
@pytest.mark.local_server_test
@pytest.mark.dependency(depends=["test/integration_test.py::test_tensorflow"], scope='session')
def test_tensorflow(integration_test_url, project_name, use_google_oauth):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("tensorflow-sample", ModelType.TENSORFLOW)
    v = _get_latest_version(merlin.active_model())
    port = _get_free_port()
    p = Process(target=v.start_server, kwargs={"port": port, "build_image": True})
    p.start()
    _wait_server_ready(f"http://{host}:{port}/v1/models/{v.model.name}-{v.id}")
    request_json = {
        "signature_name": "predict",
        "instances": [
            {"sepal_length": 2.8, "sepal_width": 1.0, "petal_length": 6.8,
             "petal_width": 0.4},
            {"sepal_length": 0.1, "sepal_width": 0.5, "petal_length": 1.8,
             "petal_width": 2.4}
        ]
    }
    resp = requests.post(_get_local_endpoint(v, port), json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()['predictions']) == len(request_json['instances'])
    p.terminate()


@pytest.mark.integration
@pytest.mark.local_server_test
@pytest.mark.dependency(depends=["test/integration_test.py::test_pytorch"], scope='session')
def test_pytorch(integration_test_url, project_name, use_google_oauth):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("pytorch-sample", ModelType.PYTORCH)
    model_dir = "test/pytorch-model"

    with merlin.new_model_version() as v:
        merlin.log_model(model_dir=model_dir)

    port = _get_free_port()
    p = Process(target=v.start_server, kwargs={"port": port, "build_image": True})
    p.start()
    _wait_server_ready(f"http://{host}:{port}")
    resp = requests.post(_get_local_endpoint(v, port), json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()['predictions']) == len(request_json['instances'])
    p.terminate()


@pytest.mark.local_server_test
@pytest.mark.integration
@pytest.mark.dependency(depends=["test/pyfunc_integration_test.py::test_pyfunc"], scope='session')
def test_pyfunc(integration_test_url, project_name, use_google_oauth):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("pyfunc-sample", ModelType.PYFUNC)
    v = _get_latest_version(merlin.active_model())
    port = _get_free_port()
    p = Process(target=v.start_server, kwargs={"port": port})
    p.start()
    _wait_server_ready(f"http://{host}:{port}", timeout_second=600)
    resp = requests.post(_get_local_endpoint(v, port), json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()['predictions']) == len(request_json['instances'])
    p.terminate()


def _get_latest_version(model):
    v_list = model.list_version()
    v_list.sort(key=lambda v: v.id, reverse=True)
    return v_list[0]


def _get_free_port():
    sock = None
    try:
        while (True):
            port = random.randint(8000, 9000)
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            result = sock.connect_ex(('127.0.0.1', port))
            if result != 0:
                return port
    finally:
        if sock is not None:
            sock.close()


def _wait_server_ready(url, timeout_second=300, tick_second=5):
    ellapsed_second = 0
    while (ellapsed_second < timeout_second):
        try:
            resp = requests.get(url)
            if resp.status_code == 200:
                return
        except Exception as e:
            print(f"{url} is not ready: {e}")

        time.sleep(tick_second)
        ellapsed_second += tick_second

    if ellapsed_second >= timeout_second:
        raise TimeoutError("server is not ready within specified timeout duration")


def _get_local_endpoint(model_version, port):
    return f"http://{host}:{port}/v1/models/{model_version.model.name}-{model_version.id}:predict"
