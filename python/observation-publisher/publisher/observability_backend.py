import abc

import pandas as pd
from arize.pandas.logger import Client, Schema
from arize.utils.types import Environments

from publisher.config import (ArizeConfig, ModelSpec, ObservabilityBackend,
                              ObservabilityBackendTypes)


class ObservationSink(abc.ABC):
    @abc.abstractmethod
    def write(self, dataframe: pd.DataFrame):
        raise NotImplementedError


class ArizeSink(ObservationSink):
    def __init__(self, config: ArizeConfig, model_spec: ModelSpec):
        self._client = Client(space_key=config.space_key, api_key=config.api_key)
        self._schema = Schema(
            feature_column_names=model_spec.schema.feature_columns,
            prediction_label_column_name=model_spec.schema.prediction_label_column,
            prediction_score_column_name=model_spec.schema.prediction_score_column,
            timestamp_column_name=model_spec.schema.timestamp_column,
        )
        self._model_id = model_spec.id
        self._model_version = model_spec.version
        self._model_type = model_spec.type

    def write(self, df: pd.DataFrame):
        self._client.log(
            dataframe=df,
            environment=Environments.PRODUCTION,
            schema=self._schema,
            model_id=self._model_id,
            model_type=self._model_type,
            model_version=self._model_version,
        )


def new_observation_sink(
    config: ObservabilityBackend, model_spec: ModelSpec
) -> ObservationSink:
    if config.type == ObservabilityBackendTypes.ARIZE:
        return ArizeSink(config=config.arize_config, model_spec=model_spec)
    else:
        raise ValueError(f"Unknown observability backend type: {config.type}")
