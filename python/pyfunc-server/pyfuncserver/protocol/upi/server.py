import logging
import multiprocessing
from concurrent import futures

import grpc
from grpc import aio
import asyncio
from typing import Optional
from caraml.upi.v1 import upi_pb2, upi_pb2_grpc
from grpc_reflection.v1alpha import reflection
from grpc_health.v1.health import HealthServicer
from grpc_health.v1 import health_pb2_grpc

from pyfuncserver.config import Config
from pyfuncserver.model.model import PyFuncModel
from pyfuncserver.publisher.publisher import Publisher
from pyfuncserver.publisher.kafka import KafkaProducer
from pyfuncserver.sampler.sampler import RatioSampling
from merlin.pyfunc import PyFuncOutput


class PredictionService(upi_pb2_grpc.UniversalPredictionServiceServicer):
    def __init__(self, model: PyFuncModel):
        if not model.ready:
            model.load()
        self._model = model
        self._publisher: Optional[Publisher] = None

    def set_publisher(self, publisher: Publisher):
        if self._publisher is None:
            self._publisher = publisher

    def PredictValues(self, request, context):
        output = self._model.upiv1_predict(request=request, context=context)
        upi_response = output
        output_is_pyfunc_output = isinstance(upi_response, PyFuncOutput) 
        if output_is_pyfunc_output:
            upi_response = output.upi_response

        if self._publisher is not None and output_is_pyfunc_output and output.contains_prediction_log():
        # need to also check whether the output contains prediction log since the pyfunc server doesn't know which model that is used
            asyncio.create_task(self._publisher.publish(output))

        return upi_response


class UPIServer:
    def __init__(self, model: PyFuncModel, config: Config):
        self._predict_service = PredictionService(model=model)
        self._config = config
        self._health_service = HealthServicer()

    def start(self):
        logging.info(f"Starting {self._config.workers} workers")

        if self._config.workers > 1:
            # multiprocessing based on https://github.com/grpc/grpc/tree/master/examples/python/multiprocessing
            workers = []
            for _ in range(self._config.workers - 1):
                worker = multiprocessing.Process(target=self._run_server)
                worker.start()
                workers.append(worker)
        
        if self._config.publisher is not None:
            kafka_producer = KafkaProducer(self._config.publisher, self._config.model_manifest)
            sampler = RatioSampling(self._config.publisher.sampling_ratio)
            publisher = Publisher(kafka_producer, sampler)
            self._predict_service.set_publisher(publisher)

        asyncio.get_event_loop().run_until_complete(self._run_server())

    async def _run_server(self):
        """
            Start a server in a subprocess.

        """
        options = self._config.grpc_options
        options.append(('grpc.so_reuseport', 1))

        server = aio.server(futures.ThreadPoolExecutor(max_workers=self._config.grpc_concurrency),
                             options=options)
        upi_pb2_grpc.add_UniversalPredictionServiceServicer_to_server(self._predict_service, server)
        health_pb2_grpc.add_HealthServicer_to_server(self._health_service, server)

        # Enable reflection server for debugging
        SERVICE_NAMES = (
            upi_pb2.DESCRIPTOR.services_by_name['UniversalPredictionService'].full_name,
            reflection.SERVICE_NAME,
        )
        reflection.enable_server_reflection(SERVICE_NAMES, server)

        logging.info(
            f"Starting grpc service at port {self._config.grpc_port} with options {self._config.grpc_options}")
        server.add_insecure_port(f"[::]:{self._config.grpc_port}")
        await server.start()
        await server.wait_for_termination()
