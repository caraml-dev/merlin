from datetime import datetime
from typing import Any, List

import numpy as np
import pandas as pd
from caraml.upi.v1.prediction_log_pb2 import PredictionLog
from pandas._testing import assert_frame_equal

from publisher.config import ModelSchema, ModelSpec, ModelType, ValueType
from publisher.prediction_log_consumer import log_to_dataframe


def new_prediction_log(
    model_spec: ModelSpec,
    prediction_id: str,
    row_ids: List[str],
    input_data: List[List[Any]],
    output_data: List[List[Any]],
    request_timestamp: datetime,
) -> PredictionLog:
    input_columns = model_spec.schema.feature_columns
    output_columns = model_spec.schema.prediction_columns
    if len(input_data) == 0 or len(output_data) == 0:
        raise ValueError("input/output data cannot have zero length")
    if len(input_columns) != len(input_data[0]):
        raise ValueError("input columns and input data must have the same length")
    if len(output_columns) != len(output_data[0]):
        raise ValueError("input columns and input data must have the same length")

    prediction_log = PredictionLog()
    prediction_log.prediction_id = prediction_id
    prediction_log.model_name = model_spec.id
    prediction_log.model_version = model_spec.version
    input_column_types = [c.name for c in model_spec.schema.feature_column_types]
    prediction_log.input.features_table.update(
        {
            "column_types": input_column_types,
            "column_names": input_columns,
            "data": input_data,
            "row_ids": row_ids,
        }
    )
    output_column_types = [c.name for c in model_spec.schema.prediction_column_types]
    prediction_log.output.prediction_results_table.update(
        {
            "column_types": output_column_types,
            "column_names": output_columns,
            "data": output_data,
            "row_ids": row_ids,
        }
    )
    prediction_log.request_timestamp.FromDatetime(request_timestamp)
    return prediction_log


def test_log_to_dataframe():
    model_spec = ModelSpec(
        id="test_model",
        version="0.1.0",
        type=ModelType.BINARY_CLASSIFICATION,
        schema=ModelSchema(
            prediction_columns=["prediction_score", "prediction_label"],
            prediction_column_types=[ValueType.BOOLEAN, ValueType.STRING],
            prediction_label_column="prediction_label",
            prediction_score_column="prediction_score",
            timestamp_column="timestamp",
            feature_columns=[
                "acceptance_rate",
                "minutes_since_last_order",
                "service_type",
            ],
            feature_column_types=[
                ValueType.FLOAT64,
                ValueType.INT64,
                ValueType.STRING,
            ],
        ),
    )
    prediction_logs = [
        new_prediction_log(
            prediction_id="1234",
            model_spec=model_spec,
            input_data=[
                [0.8, 24, "FOOD"],
                [0.5, 2, "RIDE"],
            ],
            output_data=[
                [True, "non fraud"],
                [False, "fraud"],
            ],
            request_timestamp=datetime(2021, 1, 1, 0, 0, 0),
            row_ids=["a", "b"],
        ),
        new_prediction_log(
            prediction_id="5678",
            model_spec=model_spec,
            input_data=[
                [1.0, 13, "CAR"],
                [0.4, 60, "RIDE"],
            ],
            output_data=[
                [True, "non fraud"],
                [False, "non fraud"],
            ],
            request_timestamp=datetime(2021, 1, 1, 0, 0, 0),
            row_ids=["c", "d"],
        ),
    ]
    prediction_logs_df = log_to_dataframe(prediction_logs, model_spec.schema)
    expected_df = pd.DataFrame.from_records(
        [
            [
                0.8,
                24,
                "FOOD",
                True,
                "non fraud",
                "1234a",
                datetime(2021, 1, 1, 0, 0, 0),
            ],
            [0.5, 2, "RIDE", False, "fraud", "1234b", datetime(2021, 1, 1, 0, 0, 0)],
            [1.0, 13, "CAR", True, "non fraud", "5678c", datetime(2021, 1, 1, 0, 0, 0)],
            [
                0.4,
                60,
                "RIDE",
                False,
                "non fraud",
                "5678d",
                datetime(2021, 1, 1, 0, 0, 0),
            ],
        ],
        columns=[
            "acceptance_rate",
            "minutes_since_last_order",
            "service_type",
            "prediction_score",
            "prediction_label",
            "prediction_id",
            "timestamp",
        ],
    )
    assert_frame_equal(prediction_logs_df, expected_df)


def test_empty_column_conversion_to_dataframe():
    model_spec = ModelSpec(
        id="test_model",
        version="0.1.0",
        type=ModelType.BINARY_CLASSIFICATION,
        schema=ModelSchema(
            prediction_columns=["prediction_score"],
            prediction_column_types=[ValueType.BOOLEAN],
            prediction_score_column="prediction_score",
            timestamp_column="timestamp",
            feature_columns=[
                "acceptance_rate",
            ],
            feature_column_types=[
                ValueType.FLOAT64,
            ],
        ),
    )
    prediction_logs = [
        new_prediction_log(
            prediction_id="1234",
            model_spec=model_spec,
            input_data=[
                [None],
            ],
            output_data=[
                [True],
            ],
            request_timestamp=datetime(2021, 1, 1, 0, 0, 0),
            row_ids=["a"],
        ),
    ]
    prediction_logs_df = log_to_dataframe(prediction_logs, model_spec.schema)
    expected_df = pd.DataFrame.from_records(
        [
            [
                np.NaN,
                True,
                "1234a",
                datetime(2021, 1, 1, 0, 0, 0),
            ],
        ],
        columns=[
            "acceptance_rate",
            "prediction_score",
            "prediction_id",
            "timestamp",
        ],
    )
    assert_frame_equal(prediction_logs_df, expected_df)
