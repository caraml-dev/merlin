from dataclasses import dataclass
from enum import Enum, unique
from typing import List, Optional

import arize.utils.types
from hydra.core.config_store import ConfigStore


@unique
class ValueType(Enum):
    FLOAT64 = 1
    INT64 = 2
    BOOLEAN = 3
    STRING = 4


@dataclass
class ModelSchema:
    prediction_columns: List[str]
    prediction_column_types: List[ValueType]
    prediction_score_column: str
    timestamp_column: str
    feature_columns: List[str]
    feature_column_types: List[ValueType]
    prediction_label_column: Optional[str] = None
    prediction_id_column = "prediction_id"


@unique
class ModelType(Enum):
    BINARY_CLASSIFICATION = 1
    MULTICLASS_CLASSIFICATION = 2
    REGRESSION = 3
    RANKING = 4

    def as_arize_model_type(self) -> arize.utils.types.ModelTypes:
        return arize.utils.types.ModelTypes(self.name)


@dataclass
class ModelSpec:
    id: str
    version: str
    type: ModelType
    schema: ModelSchema


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
class ConsumerType(Enum):
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
class PredictionLogConsumerConfig:
    type: ConsumerType
    kafka_config: Optional[KafkaConsumerConfig] = None


@dataclass
class Environment:
    model: ModelSpec
    observability_backend: ObservabilityBackend
    consumer: PredictionLogConsumerConfig


@dataclass
class PublisherConfig:
    environment: Environment


cs = ConfigStore.instance()
cs.store(name="base_config", node=PublisherConfig)
