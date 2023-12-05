import abc
from typing import List

import pandas as pd
from arize.pandas.logger import Client
from arize.pandas.logger import Schema as ArizeSchema
from arize.utils.types import Environments
from arize.utils.types import ModelTypes as ArizeModelType
from merlin.observability.inference import InferenceSchema, InferenceType

from publisher.config import (ArizeConfig, ObservabilityBackend,
                              ObservabilityBackendType)
from publisher.prediction_log_parser import PREDICTION_LOG_TIMESTAMP_COLUMN


class ObservationSink(abc.ABC):
    @abc.abstractmethod
    def write(self, dataframe: pd.DataFrame):
        raise NotImplementedError


def map_to_arize_schema(inference_schema: InferenceSchema) -> List[ArizeSchema]:
    # One log will be published per model schema
    match inference_schema.type:
        case InferenceType.BINARY_CLASSIFICATION:
            return [
                ArizeSchema(
                    feature_column_names=inference_schema.feature_columns,
                    prediction_label_column_name=inference_schema.binary_classification.prediction_label_column,
                    prediction_score_column_name=inference_schema.binary_classification.prediction_score_column,
                    prediction_id_column_name=inference_schema.prediction_id_column,
                    timestamp_column_name=PREDICTION_LOG_TIMESTAMP_COLUMN,
                )
            ]
        case InferenceType.MULTICLASS_CLASSIFICATION:
            return [
                ArizeSchema(
                    feature_column_names=inference_schema.feature_columns,
                    prediction_label_column_name=prediction_label_column,
                    prediction_score_column_name=prediction_score_column,
                    prediction_id_column_name=inference_schema.prediction_id_column,
                    timestamp_column_name=PREDICTION_LOG_TIMESTAMP_COLUMN,
                )
                for prediction_label_column, prediction_score_column in zip(
                    inference_schema.multiclass_classification.prediction_label_columns,
                    inference_schema.multiclass_classification.prediction_score_columns,
                )
            ]
        case InferenceType.REGRESSION:
            return [
                ArizeSchema(
                    feature_column_names=inference_schema.feature_columns,
                    prediction_score_column_name=inference_schema.regression.prediction_score_column,
                    prediction_id_column_name=inference_schema.prediction_id_column,
                    timestamp_column_name=PREDICTION_LOG_TIMESTAMP_COLUMN,
                )
            ]
        case InferenceType.RANKING:
            return [
                ArizeSchema(
                    feature_column_names=inference_schema.feature_columns,
                    rank_column_name=inference_schema.ranking.rank_column,
                    prediction_group_id_column_name=inference_schema.ranking.prediction_group_id_column,
                    prediction_id_column_name=inference_schema.prediction_id_column,
                    timestamp_column_name=PREDICTION_LOG_TIMESTAMP_COLUMN,
                )
            ]


class ArizeSink(ObservationSink):
    def __init__(
        self,
        config: ArizeConfig,
        inference_schema: InferenceSchema,
        model_id: str,
        model_version: str,
    ):
        self._client = Client(space_key=config.space_key, api_key=config.api_key)
        self._model_id = model_id
        self._model_version = model_version
        self._inference_schema = inference_schema

    def write(self, df: pd.DataFrame):
        for schema in map_to_arize_schema(self._inference_schema):
            self._client.log(
                dataframe=df,
                environment=Environments.PRODUCTION,
                schema=schema,
                model_id=self._model_id,
                model_type=ArizeModelType(self._inference_schema.type.name),
                model_version=self._model_version,
            )


def new_observation_sink(
    config: ObservabilityBackend,
    inference_schema: InferenceSchema,
    model_id: str,
    model_version: str,
) -> ObservationSink:
    if config.type == ObservabilityBackendType.ARIZE:
        return ArizeSink(
            config=config.arize_config,
            inference_schema=inference_schema,
            model_id=model_id,
            model_version=model_version,
        )
    else:
        raise ValueError(f"Unknown observability backend type: {config.type}")
