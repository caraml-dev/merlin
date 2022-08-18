import logging
import tornado
from prometheus_client import CollectorRegistry

from pyfuncserver.metrics.handler import MetricsHandler
from pyfuncserver.model.model import PyFuncModel
from pyfuncserver.protocol.rest.handler import HealthHandler, LivenessHandler, PredictHandler


class HTTPServer:
    def __init__(self, port: int, workers: int, metrics_registry: CollectorRegistry):
        self.workers = workers
        self.http_port = port
        self.metrics_registry = metrics_registry
        self.registered_models : dict = {}

    def create_application(self):
        return tornado.web.Application([
            # Server Liveness API returns 200 if server is alive.
            (r"/", LivenessHandler),
            # Model Health API returns 200 if model is ready to serve.
            (r"/v1/models/([a-zA-Z0-9_-]+)",
             HealthHandler, dict(models=self.registered_models)),
            (r"/v1/models/([a-zA-Z0-9_-]+):predict",
             PredictHandler, dict(models=self.registered_models)),
            (r"/metrics", MetricsHandler, dict(metrics_registry=self.metrics_registry))
        ])

    def start(self, model: PyFuncModel):
        self.register_model(model)

        self._http_server = tornado.httpserver.HTTPServer(
            self.create_application())

        logging.info("Listening on port %s", self.http_port)
        self._http_server.bind(self.http_port)
        logging.info("Will fork %d workers", self.workers)
        self._http_server.start(self.workers)
        tornado.ioloop.IOLoop.current().start()

    def register_model(self, model: PyFuncModel):
        self.registered_models[model.full_name] = model
        logging.info("Registering model: name: %s, version: %s, fullname: %s", model.name, model.version,
                     model.full_name)
