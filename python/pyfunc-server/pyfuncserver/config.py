import json
import logging
import os
from typing import Optional

from merlin.protocol import Protocol
from dataclasses import dataclass
from merlin.model_schema import ModelSchema

# Following environment variables are expected to be populated by Merlin
HTTP_PORT = ("CARAML_HTTP_PORT", 8080)
MODEL_NAME = ("CARAML_MODEL_NAME", "model")
MODEL_VERSION = ("CARAML_MODEL_VERSION", "1")
MODEL_FULL_NAME = ("CARAML_MODEL_FULL_NAME", "model-1")
PROJECT = ("CARAML_PROJECT", "project")
PROTOCOL = ("CARAML_PROTOCOL", "HTTP_JSON")
MODEL_SCHEMA = "CARAML_MODEL_SCHEMA"

WORKERS = ("WORKERS", 1)
GRPC_PORT = ("CARAML_GRPC_PORT", 9000)
LOG_LEVEL = ("LOG_LEVEL", "INFO")
GRPC_OPTIONS = ("GRPC_OPTIONS", "{}")
GRPC_CONCURRENCY = ("GRPC_CONCURRENCY", 10)

PUSHGATEWAY_ENABLED = ("PUSHGATEWAY_ENABLED", "false")
PUSHGATEWAY_URL = ("PUSHGATEWAY_URL", "localhost:9091")
PUSHGATEWAY_PUSH_INTERVAL_SEC = ("PUSHGATEWAY_PUSH_INTERVAL_SEC", 30)

PUBLISHER_KAFKA_TOPIC = ("PUBLISHER_KAFKA_TOPIC", "")
PUBLISHER_KAFKA_BROKERS = ("PUBLISHER_KAFKA_BROKERS", "")
PUBLISHER_KAKFA_LINGER_MS = ("PUBLISHER_KAFKA_LINGER_MS", 1000)
PUBLISHER_KAFKA_ACKS = ("PUBLISHER_KAFKA_ACKS", 0)
PUBLISHER_KAFKA_NUM_PARTITIONS = ("PUBLISHER_KAFKA_NUM_PARTITIONS", 24)
PUBLISHER_KAFKA_REPLICATION_FACTOR = ("PUBLISHER_KAFKA_REPLICATION_FACTOR", 3)
PUBLISHER_KAFKA_CONFIG = ("PUBLISHER_KAFKA_CONFIG", "{}")
PUBLISHER_SAMPLING_RATIO = ("PUBLISHER_SAMPLING_RATIO", 0.01)
PUBLISHER_ENABLED = ("PUBLISHER_ENABLED", "false")


@dataclass
class ModelManifest:
    """
    Model Manifest
    """
    model_name: str
    model_version: str
    model_full_name: str
    model_dir: str
    project: str
    model_schema: ModelSchema = None


class PushGateway:
    def __init__(self, enabled, url, push_interval_sec):
        self.url = url
        self.enabled = enabled
        self.push_interval_sec = push_interval_sec

@dataclass
class Kafka:
    """
    Kafka configuration
    """
    topic: str
    brokers: str
    linger_ms: int
    acks: int
    configuration: dict

@dataclass
class Publisher:
    """
    Publisher configuration
    """
    # sampling ratio of data that needs to be published
    sampling_ratio: float
    enabled: bool
    kafka: Kafka
    num_partitions: int
    replication_factor: int


class Config:
    """
    Server Configuration
    """

    def __init__(self, model_dir: str):
        self.protocol = Protocol(os.getenv(*PROTOCOL))
        self.http_port = int(os.getenv(*HTTP_PORT))
        self.grpc_port = int(os.getenv(*GRPC_PORT))

        # Model manifest
        model_name = os.getenv(*MODEL_NAME)
        model_version = os.getenv(*MODEL_VERSION)
        model_full_name = os.getenv(*MODEL_FULL_NAME)
        project = os.getenv(*PROJECT)
        model_schema = None
        model_schema_from_env_var = os.getenv(MODEL_SCHEMA) 
        if model_schema_from_env_var is not None:
            model_schema = ModelSchema.from_json(os.getenv(MODEL_SCHEMA))

        self.model_manifest = ModelManifest(model_name=model_name, 
                                            model_version=model_version, 
                                            model_full_name=model_full_name, 
                                            model_dir=model_dir, 
                                            project=project,
                                            model_schema=model_schema)

        self.workers = int(os.getenv(*WORKERS))
        self.log_level = self._log_level()

        self.grpc_options = self._grpc_options()
        self.grpc_concurrency = int(os.getenv(*GRPC_CONCURRENCY))

        push_enabled = str_to_bool(os.getenv(*PUSHGATEWAY_ENABLED))
        push_url = os.getenv(*PUSHGATEWAY_URL)
        push_interval = os.getenv(*PUSHGATEWAY_PUSH_INTERVAL_SEC)
        self.push_gateway = PushGateway(push_enabled,
                                        push_url,
                                        push_interval)
        
            
        # Publisher
        self.publisher = None
        publisher_enabled = str_to_bool(os.getenv(*PUBLISHER_ENABLED))
        if publisher_enabled:
            sampling_ratio = float(os.getenv(*PUBLISHER_SAMPLING_RATIO))
            kafka_topic = os.getenv(*PUBLISHER_KAFKA_TOPIC)
            kafka_brokers = os.getenv(*PUBLISHER_KAFKA_BROKERS)
            if kafka_topic == "":
                raise ValueError("kafka topic must be set")
            if kafka_brokers == "":
                raise ValueError("kafka brokers must be set")
            kafka_linger_ms = int(os.getenv(*PUBLISHER_KAKFA_LINGER_MS))
            kafka_acks = int(os.getenv(*PUBLISHER_KAFKA_ACKS))
            num_partitions = int(os.getenv(*PUBLISHER_KAFKA_NUM_PARTITIONS))
            replication_factor = int(os.getenv(*PUBLISHER_KAFKA_REPLICATION_FACTOR))
            kafka_cfgs = self._kafka_config()
            kafka = Kafka(
                kafka_topic, 
                kafka_brokers,
                kafka_linger_ms, 
                kafka_acks, 
                kafka_cfgs)
            self.publisher = Publisher(sampling_ratio, publisher_enabled, kafka, num_partitions, replication_factor)

    
    def _kafka_config(self):
        raw_cfg = os.getenv(*PUBLISHER_KAFKA_CONFIG)
        cfg = json.loads(raw_cfg)
        return cfg

    def _log_level(self):
        log_level = os.getenv(*LOG_LEVEL)
        numeric_level = getattr(logging, log_level.upper(), None)
        if not isinstance(numeric_level, int):
            logging.warning(f"invalid log level {log_level}")
            return logging.INFO
        return numeric_level

    def _grpc_options(self):
        raw_options = os.getenv(*GRPC_OPTIONS)
        options = json.loads(raw_options)
        grpc_options = []
        for k, v in options.items():
            grpc_options.append((k, v))
        return grpc_options

def str_to_bool(str: str)->bool:
    return str.lower() in ("true", "1")