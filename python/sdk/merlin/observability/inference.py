import abc
from dataclasses import dataclass, field
from enum import Enum
from typing import Dict, Optional, List

from dataclasses_json import dataclass_json, DataClassJsonMixin, config

_pandas_installed = False

try:
    import pandas as pd

    _pandas_installed = True
except ImportError:
    pass


def require_pandas(func):
    def wrapper(*args, **kwargs):
        if not _pandas_installed:
            raise ImportError("Pandas is required for this function")
        return func(*args, **kwargs)

    return wrapper


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

    def __init_subclass__(cls, **kwargs):
        super().__init_subclass__(**kwargs)
        cls.subclass_registry[cls.__name__] = cls

    @classmethod
    def encode_with_discriminator(cls, x: DataClassJsonMixin) -> Dict:
        if type(x) not in cls.subclass_registry.values():
            raise ValueError(
                f"Input must be a subclass of {cls.__name__}, got {type(x)}"
            )

        return {
            cls.discriminator_field: type(x).__name__,
            **x.to_dict(),
        }

    @classmethod
    def decode(cls, input: Dict):
        return cls.subclass_registry[input[cls.discriminator_field]].from_dict(input)

    """
    Given an input dataframe, return a new dataframe with the necessary columns for observability,
    along with a schema for the observability backend to parse the dataframe. In place changes might
    be made to the input dataframe.
    """

    @require_pandas
    @abc.abstractmethod
    def preprocess(self, df: pd.DataFrame) -> pd.DataFrame:
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

    @require_pandas
    def preprocess(self, df: pd.DataFrame) -> pd.DataFrame:
        return df


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

    @require_pandas
    def preprocess(self, df: pd.DataFrame) -> pd.DataFrame:
        df[self.prediction_label_column] = df[self.prediction_score_column].apply(
            self.prediction_label
        )
        return df


@dataclass_json
@dataclass
class RankingOutput(PredictionOutput):
    """
    Ranking model prediction output schema.

    Attributes:
        rank_column: Name of the column containing the rank of the prediction. Rank value is greater or equals to 1.
        prediction_group_id_column: Name of the column containing the prediction group id.
        relevance_score_column: Name of the column containing the relevance score, which serves as the ground truth
            for the ranking model.
    """

    rank_column: str
    prediction_group_id_column: str
    relevance_score_column: str

    @require_pandas
    def preprocess(self, df: pd.DataFrame) -> pd.DataFrame:
        return df


@dataclass_json
@dataclass
class InferenceSchema:
    """
    Inference schema for a Merlin model, which are required to extract the necessary model observability metrics
    from either a prediction log or a Pandas dataframe, and send the metrics to an observability backend such as
    Arize AI.

    Attributes:
        feature_types: Dictionary of feature names and their types.
        model_prediction_output: Model output, supported types are regression, binary classification and ranking.
        prediction_id_column: Name of the column containing the prediction id, so that the ground truth can be joined
            to the predictions.
        tag_columns: List of column names containing tags for grouping the predictions.
    """

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
