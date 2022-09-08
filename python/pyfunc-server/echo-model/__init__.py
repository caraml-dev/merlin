from typing import List

import grpc
import mlflow
from caraml.upi.v1 import upi_pb2
from merlin.model import PyFuncModel
from prometheus_client import Counter, Gauge


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


if __name__ == "__main__":
    model_name = "echo-model"
    model_version = "1"
    mlflow.pyfunc.log_model("model", python_model=EchoUPIModel(model_name, model_version))
