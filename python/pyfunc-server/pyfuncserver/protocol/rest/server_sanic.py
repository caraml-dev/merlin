import logging

from merlin.pyfunc import PyFuncOutput
from prometheus_client import CollectorRegistry
from prometheus_client.exposition import choose_encoder
from sanic import Sanic
from sanic import json as sanic_json
from sanic import raw, text
from sanic.request import Request

from pyfuncserver.config import Config
from pyfuncserver.model.model import PyFuncModel


class HTTPServer:
    def __init__(
        self, config: Config, metrics_registry: CollectorRegistry, model: PyFuncModel
    ):

        self.config = config
        self.workers = config.workers
        self.model_manifest = config.model_manifest
        self.http_port = config.http_port
        self.metrics_registry = metrics_registry
        self.registered_models: dict = {}
        self.register_model(model)

        Sanic.start_method = "fork"
        self.app = Sanic("pyfunc")
        self.app.update_config({"ACCESS_LOG": False})
        self.setup_routes()

    def setup_routes(self):
        @self.app.post(r"/v1/models/<model:([a-zA-Z0-9_-]+):predict>")
        async def predict_handler(request: Request, model):
            model = self.get_model(model)
            output = model.predict(request.json, headers=request.headers)
            response_json = output
            output_is_pyfunc_output = isinstance(response_json, PyFuncOutput)
            if output_is_pyfunc_output:
                response_json = output.http_response
            return sanic_json(response_json)

        @self.app.get("/")
        async def liveness_handler(request: Request):
            return text("alive")

        @self.app.get(r"/v1/models/<model:([a-zA-Z0-9_-]+)>")
        async def readiness_handler(request: Request, model):
            model = self.get_model(model)
            return sanic_json({"name": model.full_name, "ready": model.ready})

        @self.app.get("/metrics")
        async def metrics_handler(request: Request):
            encoder, content_type = choose_encoder(request.headers.get("accept"))
            return raw(
                encoder(self.metrics_registry), headers={"Content-Type": content_type}
            )

    def register_model(self, model: PyFuncModel):
        self.registered_models[model.full_name] = model
        logging.info(
            "Registering model: name: %s, version: %s, fullname: %s, predict endpoint: /v1/models/%s:predict",
            model.name,
            model.version,
            model.full_name,
            model.full_name,
        )

    def get_model(self, full_name: str) -> PyFuncModel:
        if full_name not in self.registered_models:
            raise Exception(
                reason="Model with full name %s does not exist." % full_name
            )
        model = self.registered_models[full_name]
        if not model.ready:
            model.load()
        return model

    def start(self):
        self.app.run(host="0.0.0.0", port=self.http_port, workers=self.workers)
