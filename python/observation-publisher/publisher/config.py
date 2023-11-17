from dataclasses import dataclass
from enum import Enum, unique
from typing import List, Optional, Dict

import arize.utils.types
from hydra.core.config_store import ConfigStore


@unique
class ValueType(Enum):
    FLOAT64 = 1
    INT64 = 2
    BOOLEAN = 3
    STRING = 4


@dataclass
class RegressionPredictionOutput:
    prediction_score_column: str

    @property
    def column_types(self) -> Dict[str, ValueType]:
        return {self.prediction_score_column: ValueType.FLOAT64}


@dataclass
class BinaryClassificationPredictionOutput:
    prediction_label_column: str
    prediction_score_column: Optional[str] = None

    @property
    def column_types(self) -> Dict[str, ValueType]:
        ct = {self.prediction_label_column: ValueType.STRING}
        if self.prediction_score_column is not None:
            ct[self.prediction_score_column] = ValueType.FLOAT64
        return ct


@dataclass
class MulticlassClassificationPredictionOutput:
    prediction_label_column: str
    prediction_score_column: Optional[str] = None

    @property
    def column_types(self) -> Dict[str, ValueType]:
        ct = {self.prediction_label_column: ValueType.STRING}
        if self.prediction_score_column is not None:
            ct[self.prediction_score_column] = ValueType.FLOAT64
        return ct


@dataclass
class RankingPredictionOutput:
    rank_column: str
    prediction_group_id_column: str

    @property
    def column_types(self) -> Dict[str, ValueType]:
        return {
            self.rank_column: ValueType.INT64,
            self.prediction_group_id_column: ValueType.STRING,
        }


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
    feature_types: Dict[str, ValueType]
    timestamp_column: str
    type: ModelType
    binary_classification: Optional[BinaryClassificationPredictionOutput] = None
    multiclass_classification: Optional[MulticlassClassificationPredictionOutput] = None
    regression: Optional[RegressionPredictionOutput] = None
    ranking: Optional[RankingPredictionOutput] = None
    prediction_id_column: Optional[str] = "prediction_id"
    tag_columns: Optional[List[str]] = None

    @property
    def feature_columns(self) -> List[str]:
        return list(self.feature_types.keys())

    @property
    def prediction_types(self) -> Dict[str, ValueType]:
        match self.type:
            case ModelType.BINARY_CLASSIFICATION:
                return self.binary_classification.column_types
            case ModelType.MULTICLASS_CLASSIFICATION:
                return self.multiclass_classification.column_types
            case ModelType.REGRESSION:
                return self.regression.column_types
            case ModelType.RANKING:
                return self.ranking.column_types
            case _:
                raise ValueError(f"Unknown model type: {self.type}")


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
    model: ModelSpec
    observability_backend: ObservabilityBackend
    observation_source: ObservationSourceConfig


@dataclass
class PublisherConfig:
    environment: Environment


cs = ConfigStore.instance()
cs.store(name="base_config", node=PublisherConfig)
