import os
import pathlib
import re
import shutil
import subprocess
from typing import List

import grpc
import mlflow
import pandas as pd
import requests
from caraml.upi.v1 import upi_pb2, upi_pb2_grpc, value_pb2
from merlin.model import PyFuncModel
from merlin.protocol import Protocol
from prometheus_client import Counter, Gauge

from pyfuncserver.config import GRPC_PORT, HTTP_PORT, MODEL_FULL_NAME, MODEL_NAME, MODEL_VERSION, PROTOCOL, WORKERS
from test.utils import df_to_prediction_rows, wait_server_ready, prediction_result_rows_to_df


class EchoUPIModel(PyFuncModel):
    GAUGE_VALUE = 42

    def __init__(self, model_name, model_version):
        self._model_name = model_name
        self._model_version = model_version

    def initialize(self, artifacts: dict):
        self._req_count = Counter("request_count", "Number of incoming request")
        self._temp = Gauge("some_gauge", "Number of incoming request")

    def upiv1_infer(self, request: upi_pb2.PredictValuesRequest,
                    context: grpc.ServicerContext) -> upi_pb2.PredictValuesResponse:
        self._req_count.inc()
        self._temp.set(EchoUPIModel.GAUGE_VALUE)

        result_rows: List[upi_pb2.PredictionResultRow] = []
        for row in request.prediction_rows:
            result_rows.append(upi_pb2.PredictionResultRow(row_id=row.row_id, values=row.model_inputs))

        return upi_pb2.PredictValuesResponse(
            prediction_result_rows=result_rows,
            target_name=request.target_name,
            prediction_context=request.prediction_context,
            metadata=upi_pb2.ResponseMetadata(
                prediction_id=request.metadata.prediction_id,
                # TODO: allow user to get model name and version from PyFuncModel
                models=[upi_pb2.ModelMetadata(name=self._model_name, version=self._model_version)]
            )
        )


def test_basic_upi():
    model_name = "my-model"
    model_version = "2"
    target_name = "echo"

    model_full_name = f"{model_name}-{model_version}"
    http_port = "8081"
    grpc_port = "9001"

    mlflow.pyfunc.log_model("model", python_model=EchoUPIModel(model_name, model_version))
    model_path = os.path.join(mlflow.get_artifact_uri(), "model")
    env = os.environ.copy()
    mlflow.end_run()
    metrics_path = "metrics_test"

    try:
        pathlib.Path(metrics_path).mkdir(exist_ok=True)

        env[PROTOCOL] = Protocol.UPI_V1.value
        env[HTTP_PORT] = http_port
        env[GRPC_PORT] = grpc_port
        env[MODEL_NAME] = model_name
        env[MODEL_VERSION] = model_version
        env[MODEL_FULL_NAME] = model_full_name
        env[WORKERS] = "1"
        env["PROMETHEUS_MULTIPROC_DIR"] = metrics_path
        c = subprocess.Popen(["python", "-m", "pyfuncserver", "--model_dir", model_path], env=env)

        # wait till the server is up
        wait_server_ready(f"http://localhost:{http_port}/")

        channel = grpc.insecure_channel(f'localhost:{grpc_port}')
        stub = upi_pb2_grpc.UniversalPredictionServiceStub(channel)
        df = pd.DataFrame([[4, 1, "hi"]] * 3,
                          columns=['int_value', 'int_value_2', 'string_value'],
                          index=["0000", "1111", "2222"])
        prediction_id = "12345"

        prediction_context = [
            value_pb2.NamedValue(name="int_context", type=value_pb2.NamedValue.TYPE_INTEGER, integer_value=1),
            value_pb2.NamedValue(name="double_context", type=value_pb2.NamedValue.TYPE_DOUBLE, double_value=1.1),
            value_pb2.NamedValue(name="string_context", type=value_pb2.NamedValue.TYPE_STRING, string_value="hello")
        ]
        response: upi_pb2.PredictValuesResponse = stub.PredictValues(
            request=upi_pb2.PredictValuesRequest(prediction_rows=df_to_prediction_rows(df),
                                                 target_name=target_name,
                                                 prediction_context=prediction_context,
                                                 metadata=upi_pb2.RequestMetadata(
                                                     prediction_id=prediction_id, )
                                                 )
        )

        assert response.metadata.prediction_id == prediction_id
        assert response.metadata.models[0].name == model_name
        assert response.metadata.models[0].version == model_version
        assert list(response.prediction_context) == prediction_context
        assert response.target_name == target_name
        assert df.equals(prediction_result_rows_to_df(response.prediction_result_rows))

        # test metrics
        resp = requests.get(f"http://localhost:{http_port}/metrics")
        assert resp.status_code == 200

        # Check request_count counter
        matches = re.findall(r"request_count_total\s(\d\.\d)", resp.text)
        assert len(matches) == 1
        assert 1 == int(float(matches[0]))

        # Check some_gauge gauge value
        matches = re.findall(r"some_gauge\{pid=\"\d+\"\}\s(\d+.\d+)", resp.text)
        assert len(matches) > 0
        for match in matches:
            gauge_value = int(float(match))
            assert gauge_value == EchoUPIModel.GAUGE_VALUE or gauge_value == 0



    finally:
        c.kill()
        shutil.rmtree(metrics_path)
