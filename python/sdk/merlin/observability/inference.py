import abc
from dataclasses import dataclass, field
from enum import Enum, unique
from typing import Dict, List, Optional

import numpy as np
import pandas as pd
from dataclasses_json import DataClassJsonMixin, config, dataclass_json


class ObservationType(Enum):
    """
    Supported observation types.
    """

    FEATURE = "feature"
    PREDICTION = "prediction"
    GROUND_TRUTH = "ground_truth"


@unique
class ValueType(Enum):
    """
    Supported feature value types.
    """

    FLOAT64 = "float64"
    INT64 = "int64"
    BOOLEAN = "boolean"
    STRING = "string"


class PredictionOutput(abc.ABC):
    subclass_registry: dict = {}
    discriminator_field: str = "output_class"

    """
    Register subclasses of PredictionOutput.
    """

    def __init_subclass__(cls, **kwargs):
        super().__init_subclass__(**kwargs)
        PredictionOutput.subclass_registry[cls.__name__] = cls

    """
    Given a subclass of PredictionOutput, which is assumed to have dataclass json mix in, encode
    the object with a discriminator field used to differentiate the different subclasses.
    """

    @classmethod
    def encode_with_discriminator(cls, x):
        if type(x) not in cls.subclass_registry.values():
            raise ValueError(
                f"Input must be a subclass of {cls.__name__}, got {type(x)}"
            )
        if not isinstance(x, DataClassJsonMixin):
            raise ValueError(
                f"Input must be a virtual subclass of DataClassJsonMixin, got {type(x)}"
            )

        return {
            cls.discriminator_field: type(x).__name__,
            **x.to_dict(),
        }

    """
    Given a dictionary, encoded using :func:`merlin.observability.inference.PredictionOutput.encode_with_discriminator`,
    decode the dictionary back to the correct subclass of PredictionOutput.
    """

    @classmethod
    def decode(cls, input: Dict):
        return PredictionOutput.subclass_registry[
            input[cls.discriminator_field]
        ].from_dict(input)

    """
    Given an input dataframe, return a new dataframe with the necessary columns for observability,
    along with a schema for the observability backend to parse the dataframe. In place changes might
    be made to the input dataframe.
    
    :param df: Input dataframe.
    :param observation_types: Types of observations to be included in the output dataframe.
    :return: output dataframe
    """

    @abc.abstractmethod
    def preprocess(
        self, df: pd.DataFrame, observation_types: List[ObservationType]
    ) -> pd.DataFrame:
        raise NotImplementedError

    """
    Return a dictionary mapping the name of the prediction output column to its value type.
    """

    @abc.abstractmethod
    def prediction_types(self) -> Dict[str, ValueType]:
        raise NotImplementedError


@dataclass_json
@dataclass
class RegressionOutput(PredictionOutput):
    """
    Regression model prediction output schema.
    Attributes:
        prediction_score_column: Name of the column containing the prediction score.
        actual_score_column: Name of the column containing the actual score.
    """

    prediction_score_column: str
    actual_score_column: str

    def preprocess(
        self, df: pd.DataFrame, observation_types: List[ObservationType]
    ) -> pd.DataFrame:
        return df

    def prediction_types(self) -> Dict[str, ValueType]:
        return {
            self.prediction_score_column: ValueType.FLOAT64,
            self.actual_score_column: ValueType.FLOAT64,
        }


@dataclass_json
@dataclass
class BinaryClassificationOutput(PredictionOutput):
    """
    Binary classification model prediction output schema.

    Attributes:
        prediction_score_column: Name of the column containing the prediction score.
            Prediction score must be between 0.0 and 1.0.
        actual_label_column: Name of the column containing the actual class.
        positive_class_label: Label for positive class.
        negative_class_label: Label for negative class.
        score_threshold: Score threshold for prediction to be considered as positive class.
    """

    prediction_score_column: str
    actual_label_column: str
    positive_class_label: str
    negative_class_label: str
    score_threshold: float = 0.5

    @property
    def prediction_label_column(self) -> str:
        return "_prediction_label"

    @property
    def actual_score_column(self) -> str:
        return "_actual_score"

    def prediction_label(self, prediction_score: float) -> str:
        """
        Returns either positive or negative class label based on prediction score.
        :param prediction_score: Probability of positive class, between 0.0 and 1.0.
        :return: prediction label
        """
        return (
            self.positive_class_label
            if prediction_score >= self.score_threshold
            else self.negative_class_label
        )

    def actual_score(self, actual_label: str) -> float:
        """
        Derive actual score from actual label.
        :param actual_label: Actual label.
        :return: actual score. Either 0.0 for negative class or 1.0 for positive class.
        """
        if actual_label not in [self.positive_class_label, self.negative_class_label]:
            raise ValueError(
                f"Actual label must be one of the classes, got {actual_label}"
            )
        return 1.0 if actual_label == self.positive_class_label else 0.0

    def preprocess(
        self, df: pd.DataFrame, observation_types: List[ObservationType]
    ) -> pd.DataFrame:
        if ObservationType.PREDICTION in observation_types:
            df[self.prediction_label_column] = df[self.prediction_score_column].apply(
                self.prediction_label
            )
        if ObservationType.GROUND_TRUTH in observation_types:
            df[self.actual_score_column] = df[self.actual_label_column].apply(
                self.actual_score
            )
        return df

    def prediction_types(self) -> Dict[str, ValueType]:
        return {
            self.prediction_score_column: ValueType.FLOAT64,
            self.prediction_label_column: ValueType.STRING,
            self.actual_score_column: ValueType.FLOAT64,
            self.actual_label_column: ValueType.STRING,
        }


@dataclass_json
@dataclass
class RankingOutput(PredictionOutput):
    rank_score_column: str
    prediction_group_id_column: str
    relevance_score_column: str
    """
    Ranking model prediction output schema.
    Attributes:
        rank_score_column: Name of the column containing the ranking score of the prediction.
        prediction_group_id_column: Name of the column containing the prediction group id.
        relevance_score_column: Name of the column containing the relevance score of the prediction.
    """

    @property
    def rank_column(self) -> str:
        return "_rank"

    def preprocess(
        self, df: pd.DataFrame, observation_types: List[ObservationType]
    ) -> pd.DataFrame:
        if ObservationType.PREDICTION in observation_types:
            df[self.rank_column] = df.groupby(self.prediction_group_id_column)[
                self.rank_score_column
            ].rank(method="first", ascending=False).astype(np.int_)
        return df

    def prediction_types(self) -> Dict[str, ValueType]:
        return {
            self.rank_column: ValueType.INT64,
            self.prediction_group_id_column: ValueType.STRING,
            self.relevance_score_column: ValueType.FLOAT64,
        }


@dataclass_json
@dataclass
class InferenceSchema:
    feature_types: Dict[str, ValueType]
    model_prediction_output: PredictionOutput = field(
        metadata=config(
            encoder=PredictionOutput.encode_with_discriminator,
            decoder=PredictionOutput.decode,
        )
    )
    prediction_id_column: str = "prediction_id"
    tag_columns: Optional[List[str]] = None

    @property
    def feature_columns(self) -> List[str]:
        return list(self.feature_types.keys())
