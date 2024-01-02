from dataclasses import dataclass
from typing import Dict, List, Optional, Union

import numpy as np
from google.protobuf.internal.well_known_types import ListValue, Struct
from merlin.observability.inference import InferenceSchema, ValueType

PREDICTION_LOG_TIMESTAMP_COLUMN = "request_timestamp"


@dataclass
class PredictionLogFeatureTable:
    columns: List[str]
    rows: List[List[Union[np.int64, np.float64, np.bool_, np.str_]]]


@dataclass
class PredictionLogResultsTable:
    columns: List[str]
    rows: List[List[Union[np.int64, np.float64, np.bool_, np.str_]]]
    row_ids: List[str]


def convert_to_numpy_value(
    col_value: Optional[int | str | float | bool], value_type: ValueType
) -> np.int64 | np.float64 | np.bool_ | np.str_:
    match value_type:
        case ValueType.INT64:
            return np.int64(col_value)
        case ValueType.FLOAT64:
            return np.float64(col_value)
        case ValueType.BOOLEAN:
            return np.bool_(col_value)
        case ValueType.STRING:
            return np.str_(col_value)
        case _:
            raise ValueError(f"Unknown value type: {value_type}")


def convert_list_value(
    list_value: ListValue, column_names: List[str], column_types: Dict[str, ValueType]
) -> List[np.int64 | np.float64 | np.bool_ | np.str_]:
    return [
        convert_to_numpy_value(col_value, column_types[col_name])
        for col_value, col_name in zip([v for v in list_value], column_names)
    ]


def parse_struct_to_feature_table(
    table_struct: Struct, inference_schema: InferenceSchema
) -> PredictionLogFeatureTable:
    columns = [c for c in table_struct["columns"]]
    column_types = inference_schema.feature_types
    return PredictionLogFeatureTable(
        columns=columns,
        rows=[
            convert_list_value(d, columns, column_types) for d in table_struct["data"]
        ],
    )


def parse_struct_to_result_table(
    table_struct: Struct, inference_schema: InferenceSchema
) -> PredictionLogResultsTable:
    columns = [c for c in table_struct["columns"]]
    column_types = inference_schema.model_prediction_output.prediction_types()
    return PredictionLogResultsTable(
        columns=columns,
        rows=[
            convert_list_value(d, columns, column_types) for d in table_struct["data"]
        ],
        row_ids=[r for r in table_struct["row_ids"]],
    )
