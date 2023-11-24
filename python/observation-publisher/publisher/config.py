from dataclasses import dataclass
from enum import Enum, unique
from typing import Optional

from hydra.core.config_store import ConfigStore
from merlin.observability.inference import InferenceSchema


@dataclass
class ArizeConfig:
    api_key: str
    space_key: str


@unique
class ObservabilityBackendType(Enum):
    ARIZE = 1


@dataclass
class ObservabilityBackend:
    type: ObservabilityBackendType
    arize_config: Optional[ArizeConfig] = None


@unique
class ObservationSource(Enum):
    KAFKA = 1


@dataclass
class KafkaConsumerConfig:
    topic: str
    bootstrap_servers: str
    group_id: str
    batch_size: int = 100
    poll_timeout_seconds: float = 1.0
    additional_consumer_config: Optional[dict] = None


@dataclass
class ObservationSourceConfig:
    type: ObservationSource
    kafka_config: Optional[KafkaConsumerConfig] = None


@dataclass
class Environment:
    model_id: str
    model_version: str
    inference_schema: InferenceSchema
    observability_backend: ObservabilityBackend
    observation_source: ObservationSourceConfig


@dataclass
class PublisherConfig:
    environment: Environment


cs = ConfigStore.instance()
cs.store(name="base_config", node=PublisherConfig)
