from dataclasses import dataclass
from enum import Enum
from typing import List

from hydra.core.config_store import ConfigStore


class ObservationSinkType(Enum):
    ARIZE = "arize"
    BIGQUERY = "bigquery"
    MAXCOMPUTE = "maxcompute"


@dataclass
class ObservationSinkConfig:
    type: ObservationSinkType
    config: dict


class ObservationSource(Enum):
    KAFKA = "kafka"


@dataclass
class ObservationSourceConfig:
    type: ObservationSource
    config: dict
    buffer_capacity: int = 10
    buffer_max_duration_seconds: int = 60


@dataclass
class Environment:
    project: str
    model_id: str
    model_version: str
    inference_schema: dict
    observation_sinks: List[ObservationSinkConfig]
    observation_source: ObservationSourceConfig
    prometheus_port: int = 8000


@dataclass
class PublisherConfig:
    environment: Environment


cs = ConfigStore.instance()
cs.store(name="base_config", node=PublisherConfig)
