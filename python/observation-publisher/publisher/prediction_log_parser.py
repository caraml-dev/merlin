from dataclasses import dataclass
from types import NoneType
from typing import Dict, List, Optional, Union

import numpy as np
from google.protobuf.struct_pb2 import ListValue, Struct
from merlin.observability.inference import InferenceSchema, ValueType
from typing_extensions import Self

PREDICTION_LOG_TIMESTAMP_COLUMN = "request_timestamp"
PREDICTION_LOG_MODEL_VERSION_COLUMN = "model_version"


@dataclass
class PredictionLogFeatureTable:
    columns: List[str]
    rows: List[List[Union[np.int64, np.float64, np.bool_, np.str_]]]

    @classmethod
    def from_struct(
        cls, table_struct: Struct, inference_schema: InferenceSchema
    ) -> Self:
        assert isinstance(table_struct["columns"], ListValue)
        columns = list_value_as_string_list(table_struct["columns"])
        column_types = inference_schema.feature_types
        assert isinstance(table_struct["data"], ListValue)
        rows = list_value_as_rows(table_struct["data"])
        return cls(
            columns=columns,
            rows=[list_value_as_numpy_list(row, columns, column_types) for row in rows],
        )


@dataclass
class PredictionLogResultsTable:
    columns: List[str]
    rows: List[List[Union[np.int64, np.float64, np.bool_, np.str_]]]
    row_ids: List[str]

    @classmethod
    def from_struct(
        cls, table_struct: Struct, inference_schema: InferenceSchema
    ) -> Self:
        assert isinstance(table_struct["columns"], ListValue)
        assert isinstance(table_struct["data"], ListValue)
        assert isinstance(table_struct["row_ids"], ListValue)
        columns = list_value_as_string_list(table_struct["columns"])
        column_types = inference_schema.model_prediction_output.prediction_types()
        rows = list_value_as_rows(table_struct["data"])
        row_ids = list_value_as_string_list(table_struct["row_ids"])
        return cls(
            columns=columns,
            rows=[list_value_as_numpy_list(row, columns, column_types) for row in rows],
            row_ids=row_ids,
        )


def convert_to_numpy_value(
    col_value: Optional[int | str | float | bool], value_type: Optional[ValueType]
) -> np.int64 | np.float64 | np.bool_ | np.str_:
    if value_type is None:
        if isinstance(col_value, (int, float)):
            return np.float64(col_value)
        if isinstance(col_value, str):
            return np.str_(col_value)
        else:
            raise ValueError(f"Unable to infer numpy type for type: {type(col_value)}")

    match value_type:
        case ValueType.INT64:
            assert isinstance(col_value, (int, float))
            return np.int64(col_value)
        case ValueType.FLOAT64:
            assert isinstance(col_value, (int, float, NoneType))
            return np.float64(col_value)
        case ValueType.BOOLEAN:
            assert isinstance(col_value, bool)
            return np.bool_(col_value)
        case ValueType.STRING:
            assert isinstance(col_value, str)
            return np.str_(col_value)
        case _:
            raise ValueError(f"Unknown value type: {value_type}")


def list_value_as_string_list(list_value: ListValue) -> List[str]:
    string_list: List[str] = []
    for v in list_value.items():
        assert isinstance(v, str)
        string_list.append(v)
    return string_list


def list_value_as_rows(list_value: ListValue) -> List[ListValue]:
    rows: List[ListValue] = []
    for d in list_value.items():
        assert isinstance(d, ListValue)
        rows.append(d)

    return rows


def list_value_as_numpy_list(
    list_value: ListValue, column_names: List[str], column_types: Dict[str, ValueType]
) -> List[np.int64 | np.float64 | np.bool_ | np.str_]:
    column_values: List[int | str | float | bool | None] = []
    for v in list_value.items():
        assert isinstance(v, (int, str, float, bool, NoneType))
        column_values.append(v)

    return [
        convert_to_numpy_value(col_value, column_types.get(col_name))
        for col_value, col_name in zip(column_values, column_names)
    ]
