import os
import random
import socket
import time
from multiprocessing import Process

import merlin
import pytest
import requests
from examples.iris_http.iris_http import IrisModel

request_json = {"instances": [[2.8, 1.0, 6.8, 0.4], [3.1, 1.4, 4.5, 1.6]]}

if os.environ.get("CI_SERVER"):
    host = "172.17.0.1"
else:
    host = "localhost"


def _get_free_port():
    sock = None
    try:
        while True:
            port = random.randint(8000, 9000)
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            result = sock.connect_ex(("127.0.0.1", port))
            if result != 0:
                return port
    finally:
        if sock is not None:
            sock.close()


def _wait_server_ready(proc, url, timeout_second=600, tick_second=10):
    time.sleep(5)

    ellapsed_second = 0
    while ellapsed_second < timeout_second:
        if not proc.is_alive():
            if proc.exitcode is not None and proc.exitcode != 0:
                raise RuntimeError("server failed to start")

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


def _get_local_endpoint(model_full_name, port):
    return f"http://{host}:{port}/v1/models/{model_full_name}:predict"


@pytest.mark.skip(reason="need to release merlin-sdk==0.41.0 first")
@pytest.mark.local_server_test
def test_examples_iris():
    XGB_PATH = os.path.join("examples/iris_http/models/", "model_1.bst")
    SKLEARN_PATH = os.path.join("examples/iris_http/models/", "model_2.joblib")

    port = _get_free_port()

    p = Process(
        target=merlin.run_pyfunc_model,
        kwargs={
            "model_instance": IrisModel(),
            "conda_env": "examples/iris_http/env.yaml",
            "code_dir": ["examples"],
            "artifacts": {
                "xgb_model": XGB_PATH,
                "sklearn_model": SKLEARN_PATH,
            },
            "debug": True,
            "port": port,
        },
    )
    p.start()

    _wait_server_ready(p, f"http://{host}:{port}", timeout_second=600)

    resp = requests.post(_get_local_endpoint("irismodel-dev", port), json=request_json)
    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()["predictions"]) == len(request_json["instances"])
    p.terminate()
