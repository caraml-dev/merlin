import logging
from concurrent import futures

from caraml.upi.v1 import upi_pb2_grpc, upi_pb2
from grpc_reflection.v1alpha import reflection
import grpc

from pyfuncserver.model.model import PyFuncModel


class PredictionService(upi_pb2_grpc.UniversalPredictionServiceServicer):
    def __init__(self, model: PyFuncModel):
        if not model.ready:
            model.load()
        self._model = model

    def PredictValues(self, request, context):
        return self._model.upiv1_predict(request=request, context=context)


class UPIServer:
    def __init__(self, model: PyFuncModel, grpc_port: int):
        self._predict_service = PredictionService(model=model)
        self._port = grpc_port

    def start(self):
        server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
        upi_pb2_grpc.add_UniversalPredictionServiceServicer_to_server(self._predict_service, server)

        # Enable reflection server for debugging
        SERVICE_NAMES = (
            upi_pb2.DESCRIPTOR.services_by_name['UniversalPredictionService'].full_name,
            reflection.SERVICE_NAME,
        )
        reflection.enable_server_reflection(SERVICE_NAMES, server)

        logging.info(f"starting grpc service at port {self._port}")
        server.add_insecure_port(f"[::]:{self._port}")
        server.start()
        server.wait_for_termination()