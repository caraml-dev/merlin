from typing import List

import grpc
import mlflow
from caraml.upi.v1 import upi_pb2
from merlin.model import PyFuncModel


class EchoUPIModel(PyFuncModel):
    def __init__(self, model_name, model_version):
        self._model_name = model_name
        self._model_version = model_version

    def infer(self, request):
        return request

    def upiv1_infer(self, request: upi_pb2.PredictValuesRequest,
                    context: grpc.ServicerContext) -> upi_pb2.PredictValuesResponse:

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


if __name__ == "__main__":
    model_name = "echo-model"
    model_version = "1"
    mlflow.pyfunc.log_model("model", python_model=EchoUPIModel(model_name, model_version))
