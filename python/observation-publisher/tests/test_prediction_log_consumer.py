from datetime import datetime
from typing import Any, List

import numpy as np
import pandas as pd
from caraml.upi.v1.prediction_log_pb2 import PredictionLog
from merlin.observability.inference import (
    BinaryClassificationOutput,
    InferenceSchema,
    ValueType,
)
from pandas._testing import assert_frame_equal

from publisher.prediction_log_consumer import log_batch_to_dataframe


def new_prediction_log(
    model_id: str,
    model_version: str,
    prediction_id: str,
    row_ids: List[str],
    input_columns: List[str],
    input_data: List[List[Any]],
    output_columns: List[str],
    output_data: List[List[Any]],
    request_timestamp: datetime,
) -> PredictionLog:
    if len(input_data) == 0 or len(output_data) == 0:
        raise ValueError("input/output data cannot have zero length")
    if len(input_columns) != len(input_data[0]):
        raise ValueError("input columns and input data must have the same length")
    if len(output_columns) != len(output_data[0]):
        raise ValueError("input columns and input data must have the same length")

    prediction_log = PredictionLog()
    prediction_log.prediction_id = prediction_id
    prediction_log.model_name = model_id
    prediction_log.model_version = model_version
    prediction_log.input.features_table.update(
        {
            "columns": input_columns,
            "data": input_data,
            "row_ids": row_ids,
        }
    )
    prediction_log.output.prediction_results_table.update(
        {
            "columns": output_columns,
            "data": output_data,
            "row_ids": row_ids,
        }
    )
    prediction_log.request_timestamp.FromDatetime(request_timestamp)
    return prediction_log


def test_log_to_dataframe():
    model_id = "test_model"
    model_version = "0.1.0"
    inference_schema = InferenceSchema(
        feature_types={
            "acceptance_rate": ValueType.FLOAT64,
            "minutes_since_last_order": ValueType.INT64,
            "service_type": ValueType.STRING,
        },
        model_prediction_output=BinaryClassificationOutput(
            prediction_score_column="prediction_score",
            actual_label_column="actual_label",
            positive_class_label="fraud",
            negative_class_label="non fraud",
            score_threshold=0.5,
        ),
    )
    input_columns = [
        "acceptance_rate",
        "minutes_since_last_order",
        "service_type",
    ]
    output_columns = ["prediction_score"]
    prediction_logs = [
        new_prediction_log(
            prediction_id="1234",
            model_id=model_id,
            model_version=model_version,
            input_columns=input_columns,
            input_data=[
                [0.8, 24, "FOOD"],
                [0.5, 2, "RIDE"],
            ],
            output_columns=output_columns,
            output_data=[
                [0.9],
                [0.5],
            ],
            request_timestamp=datetime(2021, 1, 1, 0, 0, 0),
            row_ids=["a", "b"],
        ),
        new_prediction_log(
            prediction_id="5678",
            model_id=model_id,
            model_version=model_version,
            input_columns=input_columns,
            input_data=[
                [1.0, 13, "CAR"],
                [0.4, 60, "RIDE"],
            ],
            output_columns=output_columns,
            output_data=[
                [0.4],
                [0.2],
            ],
            request_timestamp=datetime(2021, 1, 1, 0, 0, 0),
            row_ids=["c", "d"],
        ),
    ]
    prediction_logs_df = log_batch_to_dataframe(prediction_logs, inference_schema)
    expected_df = pd.DataFrame.from_records(
        [
            [0.8, 24, "FOOD", 0.9, "1234a", datetime(2021, 1, 1, 0, 0, 0)],
            [0.5, 2, "RIDE", 0.5, "1234b", datetime(2021, 1, 1, 0, 0, 0)],
            [1.0, 13, "CAR", 0.4, "5678c", datetime(2021, 1, 1, 0, 0, 0)],
            [0.4, 60, "RIDE", 0.2, "5678d", datetime(2021, 1, 1, 0, 0, 0)],
        ],
        columns=[
            "acceptance_rate",
            "minutes_since_last_order",
            "service_type",
            "prediction_score",
            "prediction_id",
            "request_timestamp",
        ],
    )
    assert_frame_equal(prediction_logs_df, expected_df)


def test_empty_column_conversion_to_dataframe():
    model_id = "test_model"
    model_version = "0.1.0"
    inference_schema = InferenceSchema(
        feature_types={
            "acceptance_rate": ValueType.FLOAT64,
        },
        model_prediction_output=BinaryClassificationOutput(
            prediction_score_column="prediction_score",
            actual_label_column="actual_label",
            positive_class_label="fraud",
            negative_class_label="non fraud",
            score_threshold=0.5,
        ),
    )
    prediction_logs = [
        new_prediction_log(
            prediction_id="1234",
            model_id=model_id,
            model_version=model_version,
            input_columns=["acceptance_rate"],
            input_data=[
                [None],
            ],
            output_columns=["prediction_score"],
            output_data=[
                [0.5],
            ],
            request_timestamp=datetime(2021, 1, 1, 0, 0, 0),
            row_ids=["a"],
        ),
    ]
    prediction_logs_df = log_batch_to_dataframe(prediction_logs, inference_schema)
    expected_df = pd.DataFrame.from_records(
        [
            [
                np.NaN,
                0.5,
                "1234a",
                datetime(2021, 1, 1, 0, 0, 0),
            ],
        ],
        columns=[
            "acceptance_rate",
            "prediction_score",
            "prediction_id",
            "request_timestamp",
        ],
    )
    assert_frame_equal(prediction_logs_df, expected_df)
