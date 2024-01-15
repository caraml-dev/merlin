import logging
import os

import grpc
import merlin
from caraml.upi.v1 import upi_pb2
from merlin.model import PyFuncModel
from merlin.protocol import Protocol
from prometheus_client import Counter, Gauge


class EchoUPIModel(PyFuncModel):
    GAUGE_VALUE = 42

    def __init__(self, model_name, model_version):
        self._model_name = model_name
        self._model_version = model_version

    def initialize(self, artifacts: dict):
        self._req_count = Counter("request_count", "Number of incoming request")
        self._temp = Gauge("some_gauge", "Number of incoming request")

    def upiv1_infer(
        self, request: upi_pb2.PredictValuesRequest, context: grpc.ServicerContext
    ) -> upi_pb2.PredictValuesResponse:
        logging.info(f"PID: {os.getpid()}")
        return upi_pb2.PredictValuesResponse(
            prediction_result_table=request.prediction_table,
            target_name=request.target_name,
            prediction_context=request.prediction_context,
            metadata=upi_pb2.ResponseMetadata(
                prediction_id=request.metadata.prediction_id,
                # TODO: allow user to get model name and version from PyFuncModel
                models=[
                    upi_pb2.ModelMetadata(
                        name=self._model_name, version=self._model_version
                    )
                ],
            ),
        )


if __name__ == "__main__":
    model_name = "echo-model"
    model_version = "1"

    merlin.run_pyfunc_model(
        model_instance=EchoUPIModel(model_name, model_version),
        conda_env="env.yaml",
        protocol=Protocol.UPI_V1,
    )
