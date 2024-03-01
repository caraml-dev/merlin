import abc
from dataclasses import dataclass, field
from enum import Enum, unique
from typing import Dict, List, Optional

import numpy as np
import pandas as pd
from dataclasses_json import DataClassJsonMixin, config, dataclass_json


DEFAULT_SESSION_ID_COLUMN = "session_id"
DEFAULT_ROW_ID_COLUMN = "row_id"


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
    :param session_id_column: Name of the column containing the session id.
    :param row_id_column: Name of the column containing the row id.
    :param observation_types: Types of observations to be included in the output dataframe.
    :return: output dataframe
    """

    @abc.abstractmethod
    def preprocess(
        self, df: pd.DataFrame, session_id_column: str, row_id_column: str, observation_types: List[ObservationType]
    ) -> pd.DataFrame:
        raise NotImplementedError

    """
    Return a dictionary mapping the name of the prediction output column to its value type.
    """

    @abc.abstractmethod
    def prediction_types(self) -> Dict[str, ValueType]:
        raise NotImplementedError

    """
    Return a dictionary mapping the name of the ground truth output column to its value type.
    """
    @abc.abstractmethod
    def ground_truth_types(self) -> Dict[str, ValueType]:
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
        self, df: pd.DataFrame, session_id_column: str, row_id_column: str, observation_types: List[ObservationType]
    ) -> pd.DataFrame:
        return df

    def prediction_types(self) -> Dict[str, ValueType]:
        return {
            self.prediction_score_column: ValueType.FLOAT64,
        }

    def ground_truth_types(self) -> Dict[str, ValueType]:
        return {
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
        actual_score_column: Name of the column containing the actual score.
        positive_class_label: Label for positive class.
        negative_class_label: Label for negative class.
        score_threshold: Score threshold for prediction to be considered as positive class.
    """

    prediction_score_column: str
    actual_score_column: str
    positive_class_label: str
    negative_class_label: str
    score_threshold: float = 0.5

    @property
    def prediction_label_column(self) -> str:
        return "_prediction_label"

    @property
    def actual_label_column(self) -> str:
        return "_actual_label"

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

    def actual_label(self, actual_score: int) -> str:
        """
        Derive actual label from actual score.
        :param actual_score: Actual score.
        :return: actual label
        """
        return (
            self.positive_class_label
            if actual_score >= self.score_threshold
            else self.negative_class_label
        )

    def preprocess(
        self, df: pd.DataFrame, session_id_column: str, row_id_column: str, observation_types: List[ObservationType]
    ) -> pd.DataFrame:
        if ObservationType.PREDICTION in observation_types:
            df[self.prediction_label_column] = df[self.prediction_score_column].apply(
                self.prediction_label
            )
        if ObservationType.GROUND_TRUTH in observation_types:
            df[self.actual_label_column] = df[self.actual_score_column].apply(
                self.actual_label
            )
        return df

    def prediction_types(self) -> Dict[str, ValueType]:
        return {
            self.prediction_score_column: ValueType.FLOAT64,
            self.prediction_label_column: ValueType.STRING
        }

    def ground_truth_types(self) -> Dict[str, ValueType]:
        return {
            self.actual_score_column: ValueType.FLOAT64,
            self.actual_label_column: ValueType.STRING
        }


@dataclass_json
@dataclass
class RankingOutput(PredictionOutput):
    rank_score_column: str
    relevance_score_column: str
    """
    Ranking model prediction output schema.
    Attributes:
        rank_score_column: Name of the column containing the ranking score of the prediction.
        relevance_score_column: Name of the column containing the relevance score of the prediction.
    """

    @property
    def rank_column(self) -> str:
        return "_rank"

    def preprocess(
        self, df: pd.DataFrame, session_id_column: str, row_id_column: str, observation_types: List[ObservationType]
    ) -> pd.DataFrame:
        if ObservationType.PREDICTION in observation_types:
            df[self.rank_column] = (
                df.groupby(session_id_column)[self.rank_score_column]
                .rank(method="first", ascending=False)
                .astype(np.int_)
            )
        return df

    def prediction_types(self) -> Dict[str, ValueType]:
        return {
            self.rank_score_column: ValueType.FLOAT64,
            self.rank_column: ValueType.INT64,
        }

    def ground_truth_types(self) -> Dict[str, ValueType]:
        return {
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
    session_id_column: str = DEFAULT_SESSION_ID_COLUMN
    row_id_column: str = DEFAULT_ROW_ID_COLUMN
    feature_orders: Optional[List[str]] = None
    tag_columns: Optional[List[str]] = None

    @property
    def feature_columns(self) -> List[str]:
        return list(self.feature_types.keys())

    @property
    def prediction_columns(self) -> List[str]:
        return list(self.model_prediction_output.prediction_types().keys())

    @property
    def ground_truth_columns(self) -> List[str]:
        return list(self.model_prediction_output.ground_truth_types().keys())

    def preprocess(
        self, df: pd.DataFrame, observation_types: List[ObservationType]
    ) -> pd.DataFrame:
        return self.model_prediction_output.preprocess(
            df, self.session_id_column, self.row_id_column, observation_types
        )

def add_prediction_id_column(df: pd.DataFrame, session_id_column: str, row_id_column: str, prediction_id_column: str) -> pd.DataFrame:
    """
    Add prediction id column to the dataframe.
    :param df: Input dataframe.
    :param session_id_column: Name of the column containing the session id.
    :param row_id_column: Name of the column containing the row id.
    :param prediction_id_column: Name of the prediction id column.
    :return: Dataframe with prediction id column.
    """
    df[prediction_id_column] = df[session_id_column] + "_" + df[row_id_column]
    return df