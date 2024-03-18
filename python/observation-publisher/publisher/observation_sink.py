import abc
import time
from dataclasses import dataclass
from typing import List, Tuple

import pandas as pd
from arize.pandas.logger import Client as ArizeClient
from arize.pandas.logger import Schema as ArizeSchema
from arize.pandas.validation.errors import ValidationFailure
from arize.utils.types import Environments
from arize.utils.types import ModelTypes as ArizeModelType
from dataclasses_json import dataclass_json
from google.api_core.exceptions import NotFound
from google.cloud.bigquery import Client as BigQueryClient
from google.cloud.bigquery import (SchemaField, Table, TimePartitioning,
                                   TimePartitioningType)
from merlin.observability.inference import (BinaryClassificationOutput,
                                            InferenceSchema,
                                            RankingOutput, RegressionOutput,
                                            ValueType, add_prediction_id_column)

from publisher.config import ObservationSinkConfig, ObservationSinkType
from publisher.prediction_log_parser import (PREDICTION_LOG_MODEL_VERSION_COLUMN,
                                             PREDICTION_LOG_TIMESTAMP_COLUMN)


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

    ARIZE_PREDICTION_ID_COLUMN = "_prediction_id"

    def __init__(
        self,
        inference_schema: InferenceSchema,
        model_id: str,
        model_version: str,
        arize_client: ArizeClient,
    ):
        """
        :param inference_schema: Inference schema for the ingested model
        :param model_id: Merlin model id
        :param model_version: Merlin model version
        :param arize_client: Arize Pandas Logger client
        """
        super().__init__(inference_schema, model_id, model_version)
        self._client = arize_client

    def _common_arize_schema_attributes(self) -> dict:
        return dict(
            feature_column_names=self._inference_schema.feature_columns,
            prediction_id_column_name=ArizeSink.ARIZE_PREDICTION_ID_COLUMN,
            timestamp_column_name=PREDICTION_LOG_TIMESTAMP_COLUMN,
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
                prediction_score_column_name=prediction_output.rank_score_column,
                rank_column_name=prediction_output.rank_column,
                prediction_group_id_column_name=self._inference_schema.session_id_column,
            )
            model_type = ArizeModelType.RANKING
        else:
            raise ValueError(
                f"Unknown prediction output type: {type(prediction_output)}"
            )

        return model_type, ArizeSchema(**schema_attributes)

    def write(self, df: pd.DataFrame):
        model_type, arize_schema = self._to_arize_schema()
        df = add_prediction_id_column(df, self._inference_schema.session_id_column, self._inference_schema.row_id_column, ArizeSink.ARIZE_PREDICTION_ID_COLUMN)
        try:
            self._client.log(
                dataframe=df,
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
class BigQueryRetryConfig:
    """
    Configuration for retrying failed write attempts. Write could fail due to BigQuery
    taking time to update the table schema / create new table.
    Attributes:
        enabled: Whether to retry failed write attempts
        retry_attempts: Number of retry attempts
        retry_interval_seconds: Interval between retry attempts
    """

    enabled: bool = False
    retry_attempts: int = 4
    retry_interval_seconds: int = 30


@dataclass_json
@dataclass
class BigQueryConfig:
    """
    Configuration for writing to BigQuery
    Attributes:
        project: GCP project id
        dataset: BigQuery dataset name
        ttl_days: Time to live for the date partition
        retry: Configuration for retrying failed write attempts
    """

    project: str
    dataset: str
    ttl_days: int
    retry: BigQueryRetryConfig = BigQueryRetryConfig()


class BigQuerySink(ObservationSink):
    """
    Writes prediction logs to BigQuery. If the destination table doesn't exist, it will be created based on the inference schema.
    """

    def __init__(
        self,
        inference_schema: InferenceSchema,
        model_id: str,
        model_version: str,
        config: BigQueryConfig,
    ):
        """
        :param inference_schema: Inference schema for the ingested model
        :param model_id: Merlin model id
        :param model_version: Merlin model version
        :param config: Configuration to write to bigquery sink
        """
        super().__init__(inference_schema, model_id, model_version)
        self._client = BigQueryClient()
        self._inference_schema = inference_schema
        self._model_id = model_id
        self._model_version = model_version
        self._config = config
        self._table = self.create_or_update_table()

    @property
    def project(self) -> str:
        return self._config.project

    @property
    def dataset(self) -> str:
        return self._config.dataset

    @property
    def retry(self) -> BigQueryRetryConfig:
        return self._config.retry

    def create_or_update_table(self) -> Table:
        try:
            original_table = self._client.get_table(self.write_location)
            original_schema = original_table.schema
            migrated_schema = original_schema[:]
            for field in self.schema_fields:
                if field not in original_schema:
                    migrated_schema.append(field)
            if migrated_schema == original_schema:
                return original_table
            original_table.schema = migrated_schema
            return self._client.update_table(original_table, ["schema"])
        except NotFound:
            table = Table(self.write_location, schema=self.schema_fields)
            table.time_partitioning = TimePartitioning(
                type_=TimePartitioningType.DAY,
                field=PREDICTION_LOG_TIMESTAMP_COLUMN,
                expiration_ms=self._config.ttl_days * 24 * 60 * 60 * 1000,
            )
            return self._client.create_table(table=table)

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
                name=self._inference_schema.session_id_column,
                field_type="STRING",
            ),
            SchemaField(
                name=self._inference_schema.row_id_column,
                field_type="STRING",
            ),
            SchemaField(
                name=PREDICTION_LOG_TIMESTAMP_COLUMN,
                field_type="TIMESTAMP",
            ),
            SchemaField(
                name=PREDICTION_LOG_MODEL_VERSION_COLUMN,
                field_type="STRING",
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
        table_name = f"prediction_log_{self._model_id}".replace("-", "_").replace(
            ".", "_"
        )
        return f"{self.project}.{self.dataset}.{table_name}"

    def write(self, dataframe: pd.DataFrame):
        for i in range(0, self.retry.retry_attempts + 1):
            try:
                response = self._client.insert_rows_from_dataframe(
                    dataframe=dataframe, table=self._table
                )
                errors = [error for error_chunk in response for error in error_chunk]
                if len(errors) > 0:
                    if not self.retry.enabled:
                        print("Errors when inserting rows to BigQuery")
                        return
                    else:
                        print(
                            f"Errors when inserting rows to BigQuery, retrying attempt {i}/{self.retry.retry_attempts}"
                        )
                        time.sleep(self.retry.retry_interval_seconds)
                else:
                    return
            except NotFound as e:
                print(
                    f"Table not found: {e}, retrying attempt {i}/{self.retry.retry_attempts}"
                )
                time.sleep(self.retry.retry_interval_seconds)
        print(f"Failed to write to BigQuery after {self.retry.retry_attempts} attempts")


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
                config=bq_config,
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
