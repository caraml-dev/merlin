import logging
import os
import pathlib
import re
import shutil
import signal
import subprocess
import time

import grpc
import mlflow
import pandas as pd
import pytest
import requests
from caraml.upi.utils import df_to_table
from caraml.upi.v1 import upi_pb2, upi_pb2_grpc, variable_pb2, type_pb2, table_pb2
from grpc_health.v1 import health_pb2_grpc, health_pb2
from merlin.model import PyFuncModel, PyFuncV3Model
from merlin.pyfunc import ModelInput, ModelOutput, Values
from merlin.protocol import Protocol
from prometheus_client import Counter, Gauge

from pyfuncserver.config import GRPC_PORT, HTTP_PORT, MODEL_FULL_NAME, MODEL_NAME, MODEL_VERSION, PROTOCOL, WORKERS
from test.utils import wait_server_ready
from test.test_http import sample_model_input, sample_model_output


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

        return upi_pb2.PredictValuesResponse(
            prediction_result_table=request.prediction_table,
            target_name=request.target_name,
            prediction_context=request.prediction_context,
            metadata=upi_pb2.ResponseMetadata(
                prediction_id=request.metadata.prediction_id,
                # TODO: allow user to get model name and version from PyFuncModel
                models=[upi_pb2.ModelMetadata(name=self._model_name, version=self._model_version)]
            )
        )

class SampleUPIPyFuncV3(PyFuncV3Model):

    def __init__(self, model_name, model_version):
        self._model_name = model_name
        self._model_version = model_version

    def upiv1_preprocess(self, request: upi_pb2.PredictValuesRequest,
                    context: grpc.ServicerContext) -> ModelInput:
        return sample_model_input
    
    def infer(self, model_input: ModelInput) -> ModelOutput:
        return sample_model_output
    
    def upiv1_postprocess(self, model_output: ModelOutput, request: upi_pb2.PredictValuesRequest) -> upi_pb2.PredictValuesResponse:
        return upi_pb2.PredictValuesResponse(
            target_name=request.target_name,
            prediction_result_table=table_pb2.Table(
                name="prediction_result_table",
                columns=[
                    table_pb2.Column(name="prediction_score", type=type_pb2.TYPE_DOUBLE),
                    table_pb2.Column(name="prediction_label", type=type_pb2.TYPE_STRING)
                ],
                rows=[
                    table_pb2.Row(row_id="0", values=[table_pb2.Value(double_value=0.95), table_pb2.Value(string_value="complete")]),
                    table_pb2.Row(row_id="1", values=[table_pb2.Value(double_value=0.43), table_pb2.Value(string_value="incomplete")]),
                    table_pb2.Row(row_id="2", values=[table_pb2.Value(double_value=0.59), table_pb2.Value(string_value="complete")]),
                ]
            ),
            prediction_context=request.prediction_context,
            metadata=upi_pb2.ResponseMetadata(
                models=[upi_pb2.ModelMetadata(name=self._model_name, version=self._model_version)],
                prediction_id=request.metadata.prediction_id
        )
    )

@pytest.mark.parametrize("workers", [(1), (4), (8)])
@pytest.mark.benchmark
def test_benchmark_multiprocess(workers, benchmark):
    model_name = "my-model"
    model_version = "1"
    grpc_port = 9001
    http_port = 8081
    target_name = "echo"
    metrics_path = "metrics_test"

    try:
        c = start_upi_server(EchoUPIModel(model_name, model_version), model_name, model_version, http_port, grpc_port, workers, metrics_path)

        channel = grpc.insecure_channel(f'localhost:{grpc_port}')
        
        health_check_stub = health_pb2_grpc.HealthStub(channel)
        response: health_pb2.HealthCheckResponse = health_check_stub.Check(health_pb2.HealthCheckRequest())
        assert response.status == health_pb2.HealthCheckResponse.SERVING

        stub = upi_pb2_grpc.UniversalPredictionServiceStub(channel)
        df = pd.DataFrame([[4, 1, "hi"]] * 3,
                          columns=['int_value', 'int_value_2', 'string_value'],
                          index=["0000", "1111", "2222"])
        prediction_id = "12345"

        prediction_context = [
            variable_pb2.Variable(name="int_context", type=type_pb2.TYPE_INTEGER, integer_value=1),
            variable_pb2.Variable(name="double_context", type=type_pb2.TYPE_DOUBLE, double_value=1.1),
            variable_pb2.Variable(name="string_context", type=type_pb2.TYPE_STRING, string_value="hello")
        ]

        def send_request():
            return stub.PredictValues(
                request=upi_pb2.PredictValuesRequest(prediction_table=df_to_table(df, "predict"),
                                                     target_name=target_name,
                                                     prediction_context=prediction_context,
                                                     metadata=upi_pb2.RequestMetadata(
                                                         prediction_id=prediction_id, )
                                                     )
            )

        response = benchmark(send_request)

        assert response.metadata.prediction_id == prediction_id
        assert response.metadata.models[0].name == model_name
        assert response.metadata.models[0].version == model_version
        assert list(response.prediction_context) == prediction_context
        assert response.target_name == target_name
        assert df_to_table(df, "predict") == response.prediction_result_table

    finally:
        os.killpg(os.getpgid(c.pid), signal.SIGTERM)
        shutil.rmtree(metrics_path)
        # Wait until the previous server have been terminated completely
        time.sleep(5)

@pytest.mark.parametrize("workers", [(1), (4)])
def test_upi(workers):
    model_name = "my-model"
    model_version = "1"
    grpc_port = 9001
    http_port = 8081
    target_name = "echo"
    metrics_path = "metrics_test"

    try:
        
        c = start_upi_server(EchoUPIModel(model_name, model_version), model_name, model_version, http_port, grpc_port, workers, metrics_path)

        channel = grpc.insecure_channel(f'localhost:{grpc_port}')

        health_check_stub = health_pb2_grpc.HealthStub(channel)
        response: health_pb2.HealthCheckResponse = health_check_stub.Check(health_pb2.HealthCheckRequest())
        assert response.status == health_pb2.HealthCheckResponse.SERVING

        stub = upi_pb2_grpc.UniversalPredictionServiceStub(channel)
        df = pd.DataFrame([[4, 1, "hi"]] * 3,
                          columns=['int_value', 'int_value_2', 'string_value'],
                          index=["0000", "1111", "2222"])
        prediction_id = "12345"

        prediction_context = [
            variable_pb2.Variable(name="int_context", type=type_pb2.TYPE_INTEGER, integer_value=1),
            variable_pb2.Variable(name="double_context", type=type_pb2.TYPE_DOUBLE, double_value=1.1),
            variable_pb2.Variable(name="string_context", type=type_pb2.TYPE_STRING, string_value="hello")
        ]
        response: upi_pb2.PredictValuesResponse = stub.PredictValues(
            request=upi_pb2.PredictValuesRequest(prediction_table=df_to_table(df, "predict"),
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
        assert df_to_table(df, "predict") == response.prediction_result_table

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
        os.killpg(os.getpgid(c.pid), signal.SIGTERM)
        shutil.rmtree(metrics_path)
        # Wait until the previous server have been terminated completely
        time.sleep(5)


def test_pyfuncv3_upi():
    model_name = "my-model"
    model_version = "1"
    grpc_port = 9001
    http_port = 8081
    target_name = "echo"
    metrics_path = "metrics_test"

    try:
        c = start_upi_server(SampleUPIPyFuncV3(model_name, model_version),model_name, model_version, http_port, grpc_port, 1, metrics_path)

        channel = grpc.insecure_channel(f'localhost:{grpc_port}')

        health_check_stub = health_pb2_grpc.HealthStub(channel)
        response: health_pb2.HealthCheckResponse = health_check_stub.Check(health_pb2.HealthCheckRequest())
        assert response.status == health_pb2.HealthCheckResponse.SERVING

        stub = upi_pb2_grpc.UniversalPredictionServiceStub(channel)
        df = pd.DataFrame([[4, 1, "hi"]] * 3,
                          columns=['int_value', 'int_value_2', 'string_value'],
                          index=["0000", "1111", "2222"])
        prediction_id = "12345"

        prediction_context = [
            variable_pb2.Variable(name="int_context", type=type_pb2.TYPE_INTEGER, integer_value=1),
            variable_pb2.Variable(name="double_context", type=type_pb2.TYPE_DOUBLE, double_value=1.1),
            variable_pb2.Variable(name="string_context", type=type_pb2.TYPE_STRING, string_value="hello")
        ]
        response: upi_pb2.PredictValuesResponse = stub.PredictValues(
            request=upi_pb2.PredictValuesRequest(prediction_table=df_to_table(df, "predict"),
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

        # test metrics
        resp = requests.get(f"http://localhost:{http_port}/metrics")
        assert resp.status_code == 200


    finally:
        os.killpg(os.getpgid(c.pid), signal.SIGTERM)
        shutil.rmtree(metrics_path)
        # Wait until the previous server have been terminated completely
        time.sleep(5)


def start_upi_server(python_model, model_name="my-model", model_version="1", http_port=8080, grpc_port=8081, workers=1, metrics_path="prometheus"):
    model_full_name = f"{model_name}-{model_version}"

    mlflow.pyfunc.log_model("model", python_model=python_model)
    model_path = os.path.join(mlflow.get_artifact_uri(), "model")
    env = os.environ.copy()
    mlflow.end_run()
    pathlib.Path(metrics_path).mkdir(exist_ok=True)

    env[PROTOCOL[0]] = Protocol.UPI_V1.value
    env[HTTP_PORT[0]] = str(http_port)
    env[GRPC_PORT[0]] = str(grpc_port)
    env[MODEL_NAME[0]] = model_name
    env[MODEL_VERSION[0]] = model_version
    env[MODEL_FULL_NAME[0]] = model_full_name
    env[WORKERS[0]] = str(workers)
    env["PROMETHEUS_MULTIPROC_DIR"] = metrics_path
    pid = subprocess.Popen(["python", "-m", "pyfuncserver", "--model_dir", model_path], env=env, start_new_session=True)

    # wait till the server is up
    wait_server_ready(f"http://localhost:{http_port}/")
    return pid