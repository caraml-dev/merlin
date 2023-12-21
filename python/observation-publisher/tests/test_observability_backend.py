from datetime import datetime
from typing import Optional

import pandas as pd
import pyarrow as pa
from arize.pandas.logger import Client
from merlin.observability.inference import (
    BinaryClassificationOutput,
    InferenceSchema,
    ValueType,
    RankingOutput,
)
from requests import Response

from publisher.observability_backend import ArizeSink


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
        return MockResponse(pa.ipc.open_stream(pa.OSFile(path)).read_pandas(), "Success", 200)


def test_binary_classification_model_preprocessing_for_arize():
    inference_schema = InferenceSchema(
        feature_types={
            "rating": ValueType.FLOAT64,
        },
        model_prediction_output=BinaryClassificationOutput(
            prediction_score_column="prediction_score",
            actual_label_column="actual_label",
            positive_class_label="fraud",
            negative_class_label="non fraud",
            score_threshold=0.5,
        ),
    )
    arize_client = MockArizeClient(api_key="test", space_key="test")
    arize_sink = ArizeSink(
        arize_client,
        inference_schema,
        "test-model",
        "0.1.0",
    )
    request_timestamp = datetime.now()
    input_df = pd.DataFrame.from_records(
        [
            [0.8, 0.4, "1234a", request_timestamp],
            [0.5, 0.9, "1234b", request_timestamp],
        ],
        columns=[
            "rating",
            "prediction_score",
            "prediction_id",
            "request_timestamp",
        ],
    )
    arize_sink.write(input_df)


def test_ranking_model_preprocessing_for_arize():
    inference_schema = InferenceSchema(
        feature_types={
            "rating": ValueType.FLOAT64,
        },
        model_prediction_output=RankingOutput(
            rank_score_column="rank_score",
            prediction_group_id_column="order_id",
            relevance_score_column="relevance_score_column",
        ),
    )
    request_timestamp = datetime.now()
    input_df = pd.DataFrame.from_records(
        [
            [5.0, 1.0, "1234", "1001", request_timestamp],
            [4.0, 0.9, "1234", "1001", request_timestamp],
            [3.0, 0.8, "1234", "1001", request_timestamp],
        ],
        columns=[
            "rating",
            "rank_score",
            "prediction_id",
            "order_id",
            "request_timestamp",
        ],
    )
    arize_client = MockArizeClient(api_key="test", space_key="test")
    arize_sink = ArizeSink(
        arize_client,
        inference_schema,
        "test-model",
        "0.1.0",
    )
    arize_sink.write(input_df)
