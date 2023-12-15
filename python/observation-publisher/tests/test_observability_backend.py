from datetime import datetime
from typing import List

import pandas as pd
import pyarrow as pa
from arize.pandas.validation.errors import ValidationError
from arize.pandas.validation.validator import Validator
from arize.utils.types import (
    ModelTypes as ArizeModelType,
    Environments as ArizeEnvironment,
    Schema as ArizeSchema,
)
from merlin.observability.inference import (
    BinaryClassificationOutput,
    InferenceSchema,
    ValueType,
    RankingOutput,
    ObservationType,
)
from pandas._testing import assert_frame_equal

from publisher.config import ArizeConfig
from publisher.observability_backend import ArizeSink


def assert_no_validation_errors(errors: List[ValidationError]):
    try:
        assert len(errors) == 0
    except AssertionError:
        print(errors)
        raise


def assert_arize_schema_validity(
    schema: ArizeSchema,
    dataframe: pd.DataFrame,
    environment: ArizeEnvironment,
    model_id: str,
    model_version: str,
    model_type: ArizeModelType,
):
    errors = Validator.validate_required_checks(
        dataframe=dataframe,
        model_id=model_id,
        environment=environment,
        schema=schema,
        model_version=model_version,
    )
    assert_no_validation_errors(errors)
    errors = Validator.validate_params(
        dataframe=dataframe,
        model_id=model_id,
        model_type=model_type,
        environment=environment,
        schema=schema,
        model_version=model_version,
    )
    assert_no_validation_errors(errors)
    Validator.validate_types(
        model_type=model_type,
        schema=schema,
        pyarrow_schema=pa.Schema.from_pandas(dataframe),
    )
    assert_no_validation_errors(errors)


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
    arize_sink = ArizeSink(
        ArizeConfig(api_key="test", space_key="test"),
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
    processed_df = inference_schema.model_prediction_output.preprocess(
        input_df, [ObservationType.FEATURE, ObservationType.PREDICTION]
    )
    model_type, arize_schema = arize_sink.to_arize_schema()
    assert_arize_schema_validity(
        arize_schema,
        processed_df,
        ArizeEnvironment.PRODUCTION,
        "test-model",
        "0.1.0",
        model_type,
    )


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
            [1.0, "1234", "1001", request_timestamp],
            [0.9, "1234", "1001", request_timestamp],
            [0.8, "1234", "1001", request_timestamp],
        ],
        columns=[
            "rank_score",
            "prediction_id",
            "order_id",
            "request_timestamp",
        ],
    )
    arize_sink = ArizeSink(
        ArizeConfig(api_key="test", space_key="test"),
        inference_schema,
        "test-model",
        "0.1.0",
    )
    processed_df = inference_schema.model_prediction_output.preprocess(
        input_df, [ObservationType.FEATURE, ObservationType.PREDICTION]
    )
    model_type, arize_schema = arize_sink.to_arize_schema()
    assert_arize_schema_validity(
        arize_schema,
        processed_df,
        ArizeEnvironment.PRODUCTION,
        "test-model",
        "0.1.0",
        ArizeModelType.RANKING,
    )
