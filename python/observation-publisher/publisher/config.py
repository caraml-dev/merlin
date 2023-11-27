from dataclasses import dataclass
from enum import Enum, auto
from typing import Optional

from hydra.core.config_store import ConfigStore


@dataclass
class ArizeConfig:
    api_key: str
    space_key: str


class ObservabilityBackendType(Enum):
    ARIZE = "arize"


@dataclass
class ObservabilityBackend:
    type: ObservabilityBackendType
    arize_config: Optional[ArizeConfig] = None

    def __post_init__(self):
        if self.type == ObservabilityBackendType.ARIZE:
            assert (
                self.arize_config is not None
            ), "Arize config must be set for Arize observability backend"


class ObservationSource(Enum):
    KAFKA = "kafka"


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

    def __post_init__(self):
        if self.type == ObservationSource.KAFKA:
            assert (
                self.kafka_config is not None
            ), "Kafka config must be set for Kafka observation source"


@dataclass
class Environment:
    model_id: str
    model_version: str
    inference_schema: dict
    observability_backend: ObservabilityBackend
    observation_source: ObservationSourceConfig


@dataclass
class PublisherConfig:
    environment: Environment


cs = ConfigStore.instance()
cs.store(name="base_config", node=PublisherConfig)
