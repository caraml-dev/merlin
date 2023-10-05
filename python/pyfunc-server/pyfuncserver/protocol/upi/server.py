import asyncio
from concurrent import futures
import logging
import multiprocessing

from grpc import aio
from caraml.upi.v1 import upi_pb2, upi_pb2_grpc
from grpc_reflection.v1alpha import reflection
from grpc_health.v1.health import HealthServicer
from grpc_health.v1 import health_pb2_grpc

from pyfuncserver.config import Config
from pyfuncserver.model.model import PyFuncModel

class PredictionService(upi_pb2_grpc.UniversalPredictionServiceServicer):
    def __init__(self, model: PyFuncModel):
        if not model.ready:
            model.load()
        self._model = model

    def PredictValues(self, request, context):
        return self._model.upiv1_predict(request=request, context=context)


class UPIServer:
    def __init__(self, model: PyFuncModel, config: Config):
        self._predict_service = PredictionService(model=model)
        self._config = config
        self._health_service = HealthServicer()
        self._upi_server = None

    def start(self):
        logging.info(f"Starting {self._config.workers} workers")

        if self._config.workers > 1:
            # multiprocessing based on https://github.com/grpc/grpc/tree/master/examples/python/multiprocessing
            workers = []
            for _ in range(self._config.workers - 1):
                worker = multiprocessing.Process(target=self._run_server)
                worker.start()
                workers.append(worker)

        asyncio.get_event_loop().run_until_complete(self._run_server())

    async def stop(self, after_termination):
        logging.info(f"Stopping server") 
        await self._upi_server.stop(grace=None)
        after_termination()

    async def _run_server(self):
        """
            Start a server in a subprocess.

        """
        options = self._config.grpc_options
        options.append(('grpc.so_reuseport', 1))

        self._upi_server = aio.server(futures.ThreadPoolExecutor(max_workers=self._config.grpc_concurrency),
                            options=options)
        upi_pb2_grpc.add_UniversalPredictionServiceServicer_to_server(self._predict_service, self._upi_server)
        health_pb2_grpc.add_HealthServicer_to_server(self._health_service, self._upi_server)

        # Enable reflection server for debugging
        SERVICE_NAMES = (
            upi_pb2.DESCRIPTOR.services_by_name['UniversalPredictionService'].full_name,
            reflection.SERVICE_NAME,
        )
        reflection.enable_server_reflection(SERVICE_NAMES, self._upi_server)

        logging.info(
            f"Starting grpc service at port {self._config.grpc_port} with options {self._config.grpc_options}")
        self._upi_server.add_insecure_port(f"[::]:{self._config.grpc_port}")
        
        await self._upi_server.start()
        await self._upi_server.wait_for_termination()
