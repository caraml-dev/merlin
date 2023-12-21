import abc
from typing import Tuple

import pandas as pd
from arize.pandas.logger import Client
from arize.pandas.logger import Schema as ArizeSchema
from arize.pandas.validation.errors import ValidationFailure
from arize.utils.types import Environments
from arize.utils.types import ModelTypes as ArizeModelType
from merlin.observability.inference import (
    InferenceSchema,
    RegressionOutput,
    BinaryClassificationOutput,
    RankingOutput,
    ObservationType,
)

from publisher.config import ObservabilityBackend, ObservabilityBackendType
from publisher.prediction_log_parser import PREDICTION_LOG_TIMESTAMP_COLUMN


class ObservationSink(abc.ABC):
    @abc.abstractmethod
    def write(self, dataframe: pd.DataFrame):
        raise NotImplementedError


class ArizeSink(ObservationSink):
    def __init__(
        self,
        arize_client: Client,
        inference_schema: InferenceSchema,
        model_id: str,
        model_version: str,
    ):
        self._client = arize_client
        self._model_id = model_id
        self._model_version = model_version
        self._inference_schema = inference_schema

    def common_arize_schema_attributes(self) -> dict:
        return dict(
            feature_column_names=self._inference_schema.feature_columns,
            prediction_id_column_name=self._inference_schema.prediction_id_column,
            timestamp_column_name=PREDICTION_LOG_TIMESTAMP_COLUMN,
            tag_column_names=self._inference_schema.tag_columns,
        )

    def to_arize_schema(self) -> Tuple[ArizeModelType, ArizeSchema]:
        prediction_output = self._inference_schema.model_prediction_output
        if isinstance(prediction_output, BinaryClassificationOutput):
            schema_attributes = self.common_arize_schema_attributes() | dict(
                prediction_label_column_name=prediction_output.prediction_label_column,
                prediction_score_column_name=prediction_output.prediction_score_column,
            )
            model_type = ArizeModelType.BINARY_CLASSIFICATION
        elif isinstance(prediction_output, RegressionOutput):
            schema_attributes = self.common_arize_schema_attributes() | dict(
                prediction_score_column_name=prediction_output.prediction_score_column,
            )
            model_type = ArizeModelType.REGRESSION
        elif isinstance(prediction_output, RankingOutput):
            schema_attributes = self.common_arize_schema_attributes() | dict(
                rank_column_name=prediction_output.rank_column,
                prediction_group_id_column_name=prediction_output.prediction_group_id_column,
            )
            model_type = ArizeModelType.RANKING
        else:
            raise ValueError(
                f"Unknown prediction output type: {type(prediction_output)}"
            )

        return model_type, ArizeSchema(**schema_attributes)

    def write(self, df: pd.DataFrame):
        processed_df = self._inference_schema.model_prediction_output.preprocess(
            df, [ObservationType.FEATURE, ObservationType.PREDICTION]
        )
        model_type, arize_schema = self.to_arize_schema()
        try:
            self._client.log(
                dataframe=processed_df,
                environment=Environments.PRODUCTION,
                schema=arize_schema,
                model_id=self._model_id,
                model_type=model_type,
                model_version=self._model_version,
            )
        except ValidationFailure as e:
            error_mesage = "\n".join([err.error_message() for err in e.errors])
            print(f"Failed to log to Arize: {error_mesage}")
            raise e
        except Exception as e:
            print(f"Failed to log to Arize: {e}")
            raise e


def new_observation_sink(
    config: ObservabilityBackend,
    inference_schema: InferenceSchema,
    model_id: str,
    model_version: str,
) -> ObservationSink:
    if config.type == ObservabilityBackendType.ARIZE:
        client = Client(space_key=config.arize_config.space_key, api_key=config.arize_config.api_key)
        return ArizeSink(
            arize_client = client,
            inference_schema=inference_schema,
            model_id=model_id,
            model_version=model_version,
        )
    else:
        raise ValueError(f"Unknown observability backend type: {config.type}")
