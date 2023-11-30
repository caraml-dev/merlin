from dataclasses import dataclass
from enum import unique, Enum
from typing import Dict, Optional, List

from dataclasses_json import dataclass_json


@unique
class ValueType(Enum):
    FLOAT64 = 1
    INT64 = 2
    BOOLEAN = 3
    STRING = 4


@dataclass_json
@dataclass
class RegressionOutput:
    prediction_score_column: str

    @property
    def column_types(self) -> Dict[str, ValueType]:
        return {self.prediction_score_column: ValueType.FLOAT64}


@dataclass_json
@dataclass
class BinaryClassificationOutput:
    prediction_score_column: str
    prediction_label_column: Optional[str] = None

    @property
    def column_types(self) -> Dict[str, ValueType]:
        column_types_mapping = {self.prediction_score_column: ValueType.FLOAT64}
        if self.prediction_label_column is not None:
            column_types_mapping[self.prediction_label_column] = ValueType.STRING
        return column_types_mapping


@dataclass_json
@dataclass
class MulticlassClassificationOutput:
    prediction_score_columns: List[str]
    prediction_label_columns: Optional[List[str]] = None

    @property
    def column_types(self) -> Dict[str, ValueType]:
        column_types_mapping = {
            label_column: ValueType.FLOAT64
            for label_column in self.prediction_score_columns
        }
        if self.prediction_label_columns is not None:
            for column_name in self.prediction_label_columns:
                column_types_mapping[column_name] = ValueType.STRING
        return column_types_mapping


@dataclass_json
@dataclass
class RankingOutput:
    rank_column: str
    prediction_group_id_column: str

    @property
    def column_types(self) -> Dict[str, ValueType]:
        return {
            self.rank_column: ValueType.INT64,
            self.prediction_group_id_column: ValueType.STRING,
        }


@unique
class InferenceType(Enum):
    BINARY_CLASSIFICATION = 1
    MULTICLASS_CLASSIFICATION = 2
    REGRESSION = 3
    RANKING = 4


@dataclass_json
@dataclass
class InferenceSchema:
    feature_types: Dict[str, ValueType]
    type: InferenceType
    binary_classification: Optional[BinaryClassificationOutput] = None
    multiclass_classification: Optional[MulticlassClassificationOutput] = None
    regression: Optional[RegressionOutput] = None
    ranking: Optional[RankingOutput] = None
    prediction_id_column: Optional[str] = "prediction_id"
    tag_columns: Optional[List[str]] = None

    @property
    def feature_columns(self) -> List[str]:
        return list(self.feature_types.keys())

    @property
    def prediction_column_types(self) -> Dict[str, ValueType]:
        if self.type == InferenceType.BINARY_CLASSIFICATION:
            assert self.binary_classification is not None
            return self.binary_classification.column_types
        elif self.type == InferenceType.MULTICLASS_CLASSIFICATION:
            assert self.multiclass_classification is not None
            return self.multiclass_classification.column_types
        elif self.type == InferenceType.REGRESSION:
            assert self.regression is not None
            return self.regression.column_types
        elif self.type == InferenceType.RANKING:
            assert self.ranking is not None
            return self.ranking.column_types
        else:
            raise ValueError(f"Unknown prediction type: {self.type}")
