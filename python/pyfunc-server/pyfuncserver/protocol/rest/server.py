import asyncio
import logging
import signal

import tornado
from prometheus_client import CollectorRegistry
from pyfuncserver.config import Config
from pyfuncserver.metrics.handler import MetricsHandler
from pyfuncserver.model.model import PyFuncModel
from pyfuncserver.protocol.rest.handler import (HealthHandler, LivenessHandler,
                                                PredictHandler)
from pyfuncserver.publisher.kafka import KafkaProducer
from pyfuncserver.publisher.publisher import Publisher
from pyfuncserver.sampler.sampler import RatioSampling


async def sig_handler(server):
    logging.info("Stop accepting new connection")
    server.stop()
    logging.info("Closing existing connection")
    await server.close_all_connections()
    logging.info("Stopping ioloop")
    tornado.ioloop.IOLoop.current().stop()


class Application(tornado.web.Application):
    def __init__(
        self,
        config: Config,
        metrics_registry: CollectorRegistry,
        registered_models: dict,
    ):
        self.publisher = None
        self.model_manifest = config.model_manifest
        handlers = [
            # Server Liveness API returns 200 if server is alive.
            (r"/", LivenessHandler),
            # Model Health API returns 200 if model is ready to serve.
            (
                r"/v1/models/([a-zA-Z0-9_-]+)",
                HealthHandler,
                dict(models=registered_models),
            ),
            (
                r"/v1/models/([a-zA-Z0-9_-]+):predict",
                PredictHandler,
                dict(models=registered_models),
            ),
            (r"/metrics", MetricsHandler, dict(metrics_registry=metrics_registry)),
        ]
        super().__init__(handlers)  # type: ignore  # noqa


class HTTPServer:
    def __init__(
        self, model: PyFuncModel, config: Config, metrics_registry: CollectorRegistry
    ):
        self.config = config
        self.workers = config.workers
        self.model_manifest = config.model_manifest
        self.http_port = config.http_port
        self.metrics_registry = metrics_registry
        self.registered_models: dict = {}
        self.register_model(model)

    def create_application(self):
        return Application(self.config, self.metrics_registry, self.registered_models)

    def start(self):
        application = self.create_application()
        self._http_server = tornado.httpserver.HTTPServer(application)
        logging.info("Listening on port %s", self.http_port)
        self._http_server.bind(self.http_port)
        logging.info("Will fork %d workers", self.workers)
        self._http_server.start(self.workers)

        # kafka producer must be initialize after fork the process
        if self.config.publisher is not None:
            kafka_producer = KafkaProducer(
                self.config.publisher, self.config.model_manifest
            )
            sampler = RatioSampling(self.config.publisher.sampling_ratio)
            application.publisher = Publisher(kafka_producer, sampler)

        for signame in ("SIGINT", "SIGTERM"):
            asyncio.get_event_loop().add_signal_handler(
                getattr(signal, signame),
                lambda: asyncio.create_task(sig_handler(self._http_server)),
            )

        tornado.ioloop.IOLoop.current().start()

    def register_model(self, model: PyFuncModel):
        self.registered_models[model.full_name] = model
        logging.info(
            "Registering model: name: %s, version: %s, fullname: %s, predict endpoint: /v1/models/%s:predict",
            model.name,
            model.version,
            model.full_name,
            model.full_name,
        )
