import abc
from dataclasses import dataclass
from datetime import datetime, timedelta
from typing import List, Tuple

import pandas as pd
from arize.pandas.logger import Client as ArizeClient
from arize.pandas.logger import Schema as ArizeSchema
from arize.pandas.validation.errors import ValidationFailure
from arize.utils.types import Environments
from arize.utils.types import ModelTypes as ArizeModelType
from dataclasses_json import dataclass_json
from google.cloud.bigquery import Client as BigQueryClient
from google.cloud.bigquery import (
    SchemaField,
    Table,
    TimePartitioning,
    TimePartitioningType,
)
from merlin.observability.inference import (
    BinaryClassificationOutput,
    InferenceSchema,
    ObservationType,
    RankingOutput,
    RegressionOutput,
    ValueType,
)

from publisher.config import ObservationSinkConfig, ObservationSinkType
from publisher.prediction_log_parser import PREDICTION_LOG_TIMESTAMP_COLUMN


class ObservationSink(abc.ABC):
    """
    An abstract class for writing prediction logs to an observability backend.
    """

    def __init__(
        self,
        inference_schema: InferenceSchema,
        model_id: str,
        model_version: str,
    ):
        self._inference_schema = inference_schema
        self._model_id = model_id
        self._model_version = model_version

    @abc.abstractmethod
    def write(self, dataframe: pd.DataFrame):
        """
        Convert a given pandas dataframe to PredictionLog protobuf, then send them to the observability backend.
        :param dataframe:
        :return:
        """
        raise NotImplementedError


@dataclass_json
@dataclass
class ArizeConfig:
    api_key: str
    space_key: str


class ArizeSink(ObservationSink):
    """
    Writes prediction logs to Arize AI.
    """

    def __init__(
        self,
        inference_schema: InferenceSchema,
        model_id: str,
        model_version: str,
        arize_client: ArizeClient,
    ):
        super().__init__(inference_schema, model_id, model_version)
        self._client = arize_client

    def _common_arize_schema_attributes(self) -> dict:
        return dict(
            feature_column_names=self._inference_schema.feature_columns,
            prediction_id_column_name=self._inference_schema.prediction_id_column,
            timestamp_column_name=PREDICTION_LOG_TIMESTAMP_COLUMN,
            tag_column_names=self._inference_schema.tag_columns,
        )

    def _to_arize_schema(self) -> Tuple[ArizeModelType, ArizeSchema]:
        prediction_output = self._inference_schema.model_prediction_output
        if isinstance(prediction_output, BinaryClassificationOutput):
            schema_attributes = self._common_arize_schema_attributes() | dict(
                prediction_label_column_name=prediction_output.prediction_label_column,
                prediction_score_column_name=prediction_output.prediction_score_column,
            )
            model_type = ArizeModelType.BINARY_CLASSIFICATION
        elif isinstance(prediction_output, RegressionOutput):
            schema_attributes = self._common_arize_schema_attributes() | dict(
                prediction_score_column_name=prediction_output.prediction_score_column,
            )
            model_type = ArizeModelType.REGRESSION
        elif isinstance(prediction_output, RankingOutput):
            schema_attributes = self._common_arize_schema_attributes() | dict(
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
        model_type, arize_schema = self._to_arize_schema()
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
            error_message = "\n".join([err.error_message() for err in e.errors])
            print(f"Failed to log to Arize: {error_message}")
            raise e
        except Exception as e:
            print(f"Failed to log to Arize: {e}")
            raise e


@dataclass_json
@dataclass
class BigQueryConfig:
    project: str
    dataset: str
    ttl_days: int


class BigQuerySink(ObservationSink):
    """
    Writes prediction logs to BigQuery. If the destination table doesn't exist, it will be created based on the inference schema..
    """

    def __init__(
        self,
        inference_schema: InferenceSchema,
        model_id: str,
        model_version: str,
        project: str,
        dataset: str,
        ttl_days: int,
    ):
        super().__init__(inference_schema, model_id, model_version)
        self._client = BigQueryClient()
        self._inference_schema = inference_schema
        self._model_id = model_id
        self._model_version = model_version
        self._project = project
        self._dataset = dataset
        table = Table(self.write_location, schema=self.schema_fields)
        table.time_partitioning = TimePartitioning(type_=TimePartitioningType.DAY)
        table.expires = datetime.now() + timedelta(days=ttl_days)
        self._table: Table = self._client.create_table(exists_ok=True, table=table)

    @property
    def schema_fields(self) -> List[SchemaField]:
        value_type_to_bq_type = {
            ValueType.INT64: "INTEGER",
            ValueType.FLOAT64: "FLOAT",
            ValueType.BOOLEAN: "BOOLEAN",
            ValueType.STRING: "STRING",
        }

        schema_fields = [
            SchemaField(
                name=self._inference_schema.prediction_id_column,
                field_type="STRING",
            ),
            SchemaField(
                name=PREDICTION_LOG_TIMESTAMP_COLUMN,
                field_type="TIMESTAMP",
            ),
        ]
        for feature, feature_type in self._inference_schema.feature_types.items():
            schema_fields.append(
                SchemaField(
                    name=feature, field_type=value_type_to_bq_type[feature_type]
                )
            )
        for (
            prediction,
            prediction_type,
        ) in self._inference_schema.model_prediction_output.prediction_types().items():
            schema_fields.append(
                SchemaField(
                    name=prediction, field_type=value_type_to_bq_type[prediction_type]
                )
            )

        return schema_fields

    @property
    def write_location(self) -> str:
        table_name = f"prediction_log_{self._model_id}_{self._model_version}".replace(
            "-", "_"
        ).replace(".", "_")
        return f"{self._project}.{self._dataset}.{table_name}"

    def write(self, dataframe: pd.DataFrame):
        self._client.insert_rows_from_dataframe(dataframe=dataframe, table=self._table)


def new_observation_sink(
    sink_config: ObservationSinkConfig,
    inference_schema: InferenceSchema,
    model_id: str,
    model_version: str,
) -> ObservationSink:
    match sink_config.type:
        case ObservationSinkType.BIGQUERY:
            bq_config: BigQueryConfig = BigQueryConfig.from_dict(sink_config.config)  # type: ignore[attr-defined]

            return BigQuerySink(
                inference_schema=inference_schema,
                model_id=model_id,
                model_version=model_version,
                project=bq_config.project,
                dataset=bq_config.dataset,
                ttl_days=bq_config.ttl_days,
            )
        case ObservationSinkType.ARIZE:
            arize_config: ArizeConfig = ArizeConfig.from_dict(sink_config.config)  # type: ignore[attr-defined]
            client = ArizeClient(
                space_key=arize_config.space_key, api_key=arize_config.api_key
            )
            return ArizeSink(
                inference_schema=inference_schema,
                model_id=model_id,
                model_version=model_version,
                arize_client=client,
            )
        case _:
            raise ValueError(f"Unknown observability backend type: {sink_config.type}")
