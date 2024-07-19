from dataclasses import dataclass
from types import NoneType
from typing import Dict, List, Optional, Union

import numpy as np
from google.protobuf.struct_pb2 import ListValue, Struct
from merlin.observability.inference import InferenceSchema, ValueType, BinaryClassificationOutput, RankingOutput, \
    RegressionOutput
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
        """
        Create a PredictionLogFeatureTable object from a Protobuf Struct object
        :param table_struct: A Protobuf Struct object that represents a feature table.
        :param inference_schema: Model inference schema.
        :return: Instance of PredictionLogFeatureTable.
        """
        if inference_schema.feature_orders is not None and len(inference_schema.feature_orders) > 0:
            columns = inference_schema.feature_orders
        else:
            assert isinstance(table_struct["columns"], ListValue)
            columns = list_value_as_string_list(table_struct["columns"])
        column_types = inference_schema.feature_types
        assert isinstance(table_struct["data"], ListValue)
        rows = list_value_as_rows(table_struct["data"])
        return cls(
            columns=[col for col in columns if column_types.get(col) is not None],
            rows=[list_value_as_numpy_list(row, columns, column_types) for row in rows],
        )


def prediction_columns(inference_schema: InferenceSchema) -> List[str]:
    """
    Get the column name for the prediction output
    :param inference_schema: Model inference schema
    :return: List of column names
    """
    if isinstance(inference_schema.model_prediction_output, BinaryClassificationOutput):
        return [inference_schema.model_prediction_output.prediction_score_column]
    elif isinstance(inference_schema.model_prediction_output, RankingOutput):
        return [inference_schema.model_prediction_output.rank_score_column]
    elif isinstance(inference_schema.model_prediction_output, RegressionOutput):
        return [inference_schema.model_prediction_output.prediction_score_column]
    else:
        raise ValueError(f"Unknown prediction output type: {type(inference_schema.model_prediction_output)}")


@dataclass
class PredictionLogResultsTable:
    columns: List[str]
    rows: List[List[Union[np.int64, np.float64, np.bool_, np.str_]]]
    row_ids: List[str]

    @classmethod
    def from_struct(
        cls, table_struct: Struct, inference_schema: InferenceSchema
    ) -> Self:
        """
        Create a PredictionLogResultsTable object from a Protobuf Struct object
        :param table_struct: Protobuf Struct object that represents a prediction result table.
        :param inference_schema: Model InferenceSchema.
        :return: PredictionLogResultsTable instnace.
        """
        if "columns" in table_struct.keys():
            assert isinstance(table_struct["columns"], ListValue)
            columns = list_value_as_string_list(table_struct["columns"])
        else:
            columns = prediction_columns(inference_schema)
        assert isinstance(table_struct["data"], ListValue)
        assert isinstance(table_struct["row_ids"], ListValue)
        
        rows = list_value_as_rows(table_struct["data"])
        row_ids = list_value_as_string_list(table_struct["row_ids"])
        column_types = inference_schema.model_prediction_output.prediction_types()
        return cls(
            columns=[col for col in columns if column_types.get(col) is not None],
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
    """
    Convert protobuf string list to it's native python type counterpart.
    """
    string_list: List[str] = []
    for v in list_value.items():
        assert isinstance(v, str)
        string_list.append(v)

    return string_list


def list_value_as_rows(list_value: ListValue) -> List[ListValue]:
    """
    Convert a ListValue object to a list of ListValue objects
    :param list_value: ListValue object representing a two dimensional matrix or a vector.
    :return: List of ListValue objects, representing a two dimensional matrix.
    """
    rows: List[ListValue] = []
    for d in list_value.items():
        if isinstance(d, ListValue):
            rows.append(d)
        else:
            nd = ListValue()
            nd.append(d)
            rows.append(nd)

    return rows


def list_value_as_numpy_list(
    list_value: ListValue, column_names: List[str], column_types: Dict[str, ValueType]
) -> List[np.int64 | np.float64 | np.bool_ | np.str_]:
    """
    Convert a ListValue representing a row, to it's native python type counterpart.
    :param list_value: ListValue object representing a row.
    :param column_names: Column names corresponds to each column in a row.
    :param column_types: Map of column name to type.
    :return: List of numpy types.
    """
    column_values: List[int | str | float | bool | None] = []
    for v in list_value.items():
        assert isinstance(v, (int, str, float, bool, NoneType)), f"type of value is {type(v)}"
        column_values.append(v)

    return [
        convert_to_numpy_value(col_value, column_types.get(col_name))
        for col_value, col_name in zip(column_values, column_names) if column_types.get(col_name) is not None
    ]
