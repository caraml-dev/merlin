import json
import logging
import os

# Following environment variables are expected to be populated by Merlin
from merlin.protocol import Protocol

HTTP_PORT = "CARAML_HTTP_PORT"
GRPC_PORT = "CARAML_GRPC_PORT"
MODEL_NAME = "CARAML_MODEL_NAME"
MODEL_VERSION = "CARAML_MODEL_VERSION"
MODEL_FULL_NAME = "CARAML_MODEL_FULL_NAME"
PROTOCOL = "CARAML_PROTOCOL"
WORKER = "WORKER"
LOG_LEVEL = "LOG_LEVEL"
GRPC_OPTIONS = "GRPC_OPTIONS"

DEFAULT_HTTP_PORT = 8080
DEFAULT_GRPC_PORT = 9000
DEFAULT_MODEL_NAME = "model"
DEFAULT_MODEL_VERSION = "1"
DEFAULT_FULL_NAME = f"{DEFAULT_MODEL_NAME}-{DEFAULT_MODEL_VERSION}"
DEFAULT_LOG_LEVEL = "INFO"
DEFAULT_PROTOCOL = "HTTP_JSON"
DEFAULT_GRPC_OPTIONS = "{}"

class ModelManifest:
    """
    Model Manifest
    """
    def __init__(self, model_name: str, model_version: str, model_full_name: str, model_dir: str):
        self.model_name = model_name
        self.model_version = model_version
        self.model_full_name = model_full_name
        self.model_dir = model_dir


class Config:
    """
    Server Configuration
    """
    def __init__(self, model_dir: str):
        self.protocol = Protocol(os.getenv(PROTOCOL, DEFAULT_PROTOCOL))
        self.http_port = int(os.getenv(HTTP_PORT, DEFAULT_HTTP_PORT))
        self.grpc_port = int(os.getenv(GRPC_PORT, DEFAULT_GRPC_PORT))

        # Model manifest
        model_name = os.getenv(MODEL_NAME, DEFAULT_MODEL_NAME)
        model_version = os.getenv(MODEL_VERSION, DEFAULT_MODEL_VERSION)
        model_full_name = os.getenv(MODEL_FULL_NAME, DEFAULT_FULL_NAME)
        self.model_manifest = ModelManifest(model_name, model_version, model_full_name, model_dir)

        self.workers = int(os.getenv(WORKER, 1))
        self.log_level = self._log_level()

        grpc_options = os.getenv(GRPC_OPTIONS, DEFAULT_GRPC_OPTIONS)
        self.grpc_options = json.loads(grpc_options)

    def _log_level(self):
        log_level = os.getenv(LOG_LEVEL, DEFAULT_LOG_LEVEL)
        numeric_level = getattr(logging, log_level.upper(), None)
        if not isinstance(numeric_level, int):
            logging.warning(f"invalid log level {log_level}")
            return logging.INFO
        return numeric_level

