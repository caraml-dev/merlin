import dataclasses
from datetime import datetime
from typing import Optional

import pandas as pd
import pyarrow as pa
import pytest
from arize.pandas.logger import Client
from dateutil import tz
from google.cloud.bigquery import Client as BigQueryClient
from google.cloud.bigquery import SchemaField
from merlin.observability.inference import (BinaryClassificationOutput,
                                            InferenceSchema, RankingOutput,
                                            ValueType, ObservationType)
from pandas._testing import assert_frame_equal
from requests import Response

from publisher.observation_sink import (ArizeSink, BigQueryConfig,
                                        BigQueryRetryConfig, BigQuerySink)
from tests.common_fixtures import bq_dataset, bq_project


@pytest.fixture
def binary_classification_inference_schema() -> InferenceSchema:
    return InferenceSchema(
        feature_types={
            "rating": ValueType.FLOAT64,
        },
        model_prediction_output=BinaryClassificationOutput(
            prediction_score_column="prediction_score",
            actual_score_column="actual_score",
            positive_class_label="fraud",
            negative_class_label="non fraud",
            score_threshold=0.5,
        ),
    )


@pytest.fixture
def binary_classification_inference_logs() -> pd.DataFrame:
    request_timestamp = datetime(2024, 1, 1, 0, 0, 0).astimezone(tz.UTC)
    return pd.DataFrame.from_records(
        [
            [0.8, 0.4, "1234", "a", request_timestamp, "0.1.0"],
            [0.5, 0.9, "1234", "b", request_timestamp, "0.1.0"],
        ],
        columns=[
            "rating",
            "prediction_score",
            "session_id",
            "row_id",
            "request_timestamp",
            "model_version",
        ],
    )


@pytest.fixture
def ranking_inference_schema() -> InferenceSchema:
    return InferenceSchema(
        feature_types={
            "rating": ValueType.FLOAT64,
        },
        model_prediction_output=RankingOutput(
            rank_score_column="rank_score",
            relevance_score_column="relevance_score_column",
        ),
        session_id_column="order_id",
        row_id_column="driver_id",
    )


@pytest.fixture
def ranking_inference_logs() -> pd.DataFrame:
    request_timestamp = datetime(2024, 1, 1, 0, 0, 0).astimezone(tz.UTC)
    return pd.DataFrame.from_records(
        [
            [5.0, 1.0, "1234", "1001", request_timestamp],
            [4.0, 0.9, "1234", "1002", request_timestamp],
            [3.0, 0.8, "1234", "1003", request_timestamp],
        ],
        columns=[
            "rating",
            "rank_score",
            "order_id",
            "driver_id",
            "request_timestamp"
        ],
    )


class MockResponse(Response):
    def __init__(self, df, reason, status_code):
        super().__init__()
        self.df = df
        self.reason = reason
        self.status_code = status_code


class MockArizeClient(Client):
    def _post_file(
        self,
        path: str,
        schema: bytes,
        sync: Optional[bool],
        timeout: Optional[float] = None,
    ) -> Response:
        return MockResponse(
            pa.ipc.open_stream(pa.OSFile(path)).read_pandas(), "Success", 200
        )


def test_binary_classification_model_arize_schema(
    binary_classification_inference_schema: InferenceSchema,
    binary_classification_inference_logs: pd.DataFrame,
):
    arize_client = MockArizeClient(api_key="test", space_key="test")
    arize_sink = ArizeSink(
        "test-project",
        binary_classification_inference_schema,
        "test-model",
        "0.1.0",
        arize_client,
    )
    df = binary_classification_inference_schema.preprocess(
        binary_classification_inference_logs,
        [ObservationType.PREDICTION]
    )
    arize_sink.write(df)


def test_ranking_model_arize_schema(
    ranking_inference_schema: InferenceSchema,
    ranking_inference_logs: pd.DataFrame,
):
    arize_client = MockArizeClient(api_key="test", space_key="test")
    arize_sink = ArizeSink(
        "test-project",
        ranking_inference_schema,
        "test-model",
        "0.1.0",
        arize_client,
    )
    df = ranking_inference_schema.preprocess(ranking_inference_logs, [ObservationType.PREDICTION])
    arize_sink.write(df)


@pytest.mark.integration
def test_bigquery_sink_schema_migration(
    bq_project: str,
    bq_dataset: str,
    binary_classification_inference_schema: InferenceSchema,
    binary_classification_inference_logs: pd.DataFrame,
):
    client = BigQueryClient()
    client.delete_table(
        f"{bq_project}.{bq_dataset}.prediction_log_test_project_test_model", not_found_ok=True
    )
    bq_sink = BigQuerySink(
        "test-project",
        binary_classification_inference_schema,
        "test-model",
        "0.1.0",
        config=BigQueryConfig(
            project=bq_project,
            dataset=bq_dataset,
            ttl_days=14,
            retry=BigQueryRetryConfig(
                enabled=True, retry_attempts=3, retry_interval_seconds=10
            ),
        ),
    )
    bq_sink.write(binary_classification_inference_logs)
    migrated_schema = dataclasses.replace(binary_classification_inference_schema)
    migrated_schema.feature_types = {
        "rating_v2": ValueType.FLOAT64,
    }
    migrated_bq_sink = BigQuerySink(
        migrated_schema,
        "test-model",
        "0.2.0",
        config=BigQueryConfig(
            project=bq_project,
            dataset=bq_dataset,
            ttl_days=14,
            retry=BigQueryRetryConfig(
                enabled=True, retry_attempts=5, retry_interval_seconds=30
            ),
        ),
    )
    migrated_inference_logs = binary_classification_inference_logs.rename(
        columns={"rating": "rating_v2"}
    )
    migrated_inference_logs["model_version"] = "0.2.0"
    migrated_bq_sink.write(migrated_inference_logs)
    version_update_bq_sink = BigQuerySink(
        migrated_schema,
        "test-model",
        "0.3.0",
        config=BigQueryConfig(
            project=bq_project,
            dataset=bq_dataset,
            ttl_days=14,
        ),
    )
    version_update_inference_logs = migrated_inference_logs.copy()
    version_update_inference_logs["model_version"] = "0.3.0"
    version_update_bq_sink.write(version_update_inference_logs)

    table = client.get_table(f"{bq_project}.{bq_dataset}.prediction_log_test_model")
    assert table.schema == [
        SchemaField(name="session_id", field_type="STRING"),
        SchemaField(name="row_id", field_type="STRING"),
        SchemaField(name="request_timestamp", field_type="TIMESTAMP"),
        SchemaField(name="model_version", field_type="STRING"),
        SchemaField(name="rating", field_type="FLOAT"),
        SchemaField(name="prediction_score", field_type="FLOAT"),
        SchemaField(name="_prediction_label", field_type="STRING"),
        SchemaField(name="rating_v2", field_type="FLOAT"),
    ]
    df = client.query(
        "SELECT * FROM `{}.{}.prediction_log_test_model`".format(bq_project, bq_dataset)
    ).to_dataframe()
    df.reset_index(drop=True, inplace=True)
    event_timestamp = datetime(2024, 1, 1, 0, 0, 0).astimezone(tz.UTC)
    expected_df = pd.DataFrame.from_records(
        [
            [0.8, 0.4, "1234", "a", event_timestamp, "0.1.0", "non fraud", None],
            [0.5, 0.9, "1234", "b", event_timestamp, "0.1.0", "fraud", None],
            [None, 0.4, "1234", "a", event_timestamp, "0.2.0", "non fraud", 0.8],
            [None, 0.9, "1234", "b", event_timestamp, "0.2.0", "fraud", 0.5],
            [None, 0.4, "1234", "a", event_timestamp, "0.3.0", "non fraud", 0.8],
            [None, 0.9, "1234", "b", event_timestamp, "0.3.0", "fraud", 0.5],
        ],
        columns=[
            "rating",
            "prediction_score",
            "session_id",
            "row_id",
            "request_timestamp",
            "model_version",
            "_prediction_label",
            "rating_v2",
        ],
    )
    expected_df.reset_index(drop=True, inplace=True)
    assert_frame_equal(df, expected_df, check_like=True)
