from datetime import datetime

import pandas as pd
import pyarrow as pa
from arize.pandas.validation.validator import Validator
from arize.utils.types import Environments
from arize.utils.types import ModelTypes as ArizeModelType
from merlin.observability.inference import (BinaryClassificationOutput,
                                            InferenceSchema, InferenceType,
                                            ValueType)

from publisher.observability_backend import map_to_arize_schema


def test_arize_schema_mapping():
    inference_schema = InferenceSchema(
        type=InferenceType.BINARY_CLASSIFICATION,
        feature_types={
            "acceptance_rate": ValueType.FLOAT64,
            "minutes_since_last_order": ValueType.INT64,
            "service_type": ValueType.STRING,
            "prediction_score": ValueType.FLOAT64,
            "prediction_label": ValueType.STRING,
        },
        binary_classification=BinaryClassificationOutput(
            prediction_label_column="prediction_label",
            prediction_score_column="prediction_score",
        ),
    )
    arize_schemas = map_to_arize_schema(inference_schema)
    assert len(arize_schemas) == 1
    input_dataframe = pd.DataFrame.from_records(
        [
            [0.8, 24, "FOOD", 0.9, "non fraud", "1234a", datetime(2021, 1, 1, 0, 0, 0)],
            [0.5, 2, "RIDE", 0.5, "fraud", "1234b", datetime(2021, 1, 1, 0, 0, 0)],
            [1.0, 13, "CAR", 0.4, "non fraud", "5678c", datetime(2021, 1, 1, 0, 0, 0)],
            [0.4, 60, "RIDE", 0.2, "non fraud", "5678d", datetime(2021, 1, 1, 0, 0, 0)],
        ],
        columns=[
            "acceptance_rate",
            "minutes_since_last_order",
            "service_type",
            "prediction_score",
            "prediction_label",
            "prediction_id",
            "request_timestamp",
        ],
    )
    errors = Validator.validate_required_checks(
        dataframe=input_dataframe,
        model_id="test_model",
        schema=arize_schemas[0],
        model_version="0.1.0",
    )
    assert len(errors) == 0
    errors = Validator.validate_params(
        dataframe=input_dataframe,
        model_id="test_model",
        model_type=ArizeModelType.BINARY_CLASSIFICATION,
        environment=Environments.PRODUCTION,
        schema=arize_schemas[0],
        model_version="0.1.0",
    )
    assert len(errors) == 0
    Validator.validate_types(
        model_type=ArizeModelType.BINARY_CLASSIFICATION,
        schema=arize_schemas[0],
        pyarrow_schema=pa.Schema.from_pandas(input_dataframe),
    )
    assert len(errors) == 0
