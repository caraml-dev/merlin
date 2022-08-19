import time
from typing import List

from caraml.upi.v1 import value_pb2, upi_pb2

import pandas as pd
import requests
import pandas.testing


def wait_server_ready(url, timeout_second=10, tick_second=2):
    ellapsed_second = 0
    while (ellapsed_second < timeout_second):
        try:
            resp = requests.get(url)
            if resp.status_code == 200:
                return
        except Exception as e:
            print(f"{url} is not ready: {e}")

        time.sleep(tick_second)
        ellapsed_second += tick_second

    if ellapsed_second >= timeout_second:
        raise TimeoutError("server is not ready within specified timeout duration")


def get_value(val: value_pb2.NamedValue):
    if val.type == val.TYPE_STRING:
        return val.string_value
    elif val.type == val.TYPE_DOUBLE:
        return val.double_value
    elif val.type == val.TYPE_INTEGER:
        return val.integer_value
    else:
        raise ValueError(f"unknown type {val.type}")


def df_to_prediction_rows(df: pd.DataFrame) -> List[upi_pb2.PredictionRow]:
    """
    Utilitiy function to convert pandas dataframe into upi prediction rows
    :param df: DataFrame to convert
    :return: Representation of the dataframe as prediction rows
    """
    rows: List[upi_pb2.PredictionRow] = []

    def process_row(row: pd.Series):
        model_inputs: List[value_pb2.NamedValue] = []
        for index, value in row.items():
            if isinstance(value, int):
                model_inputs.append(
                    value_pb2.NamedValue(name=str(index), type=value_pb2.NamedValue.TYPE_INTEGER, integer_value=value))
            elif isinstance(value, float):
                model_inputs.append(
                    value_pb2.NamedValue(name=str(index), type=value_pb2.NamedValue.TYPE_DOUBLE, double_value=value))
            elif isinstance(value, str):
                model_inputs.append(
                    value_pb2.NamedValue(name=str(index), type=value_pb2.NamedValue.TYPE_STRING, string_value=value))
            else:
                raise ValueError(f"unknown type {type(value)}")
        rows.append(upi_pb2.PredictionRow(row_id=str(row.name), model_inputs=model_inputs))

    df.apply(func=process_row, axis=1)
    return rows


def df_to_prediction_result_rows(df: pd.DataFrame) -> List[upi_pb2.PredictionResultRow]:
    """
    Utility function to convert pandas dataframe into upi prediction rows
    :param df:  DataFrame to convert
    :return: representation of the dataframe as prediction result rows
    """
    rows: List[upi_pb2.PredictionResultRow] = []

    def process_row(row: pd.Series):
        values: List[value_pb2.NamedValue] = []
        for index, value in row.items():
            if isinstance(value, int):
                values.append(
                    value_pb2.NamedValue(name=str(index), type=value_pb2.NamedValue.TYPE_INTEGER, integer_value=value))
            elif isinstance(value, float):
                values.append(
                    value_pb2.NamedValue(name=str(index), type=value_pb2.NamedValue.TYPE_DOUBLE, double_value=value))
            elif isinstance(value, str):
                values.append(
                    value_pb2.NamedValue(name=str(index), type=value_pb2.NamedValue.TYPE_STRING, string_value=value))
            else:
                raise ValueError(f"unknown type {type(value)}")
        rows.append(upi_pb2.PredictionResultRow(row_id=str(row.name), values=values))

    df.apply(func=process_row, axis=1)
    return rows


def prediction_result_rows_to_df(prediction_result_rows: List[upi_pb2.PredictionResultRow]) -> pd.DataFrame:
    """
    Utility function to convert prediction result rows into equivalent pandas dataframe

    :param prediction_result_rows: prediction result rows
    :return: equivalent dataframe representation
    """

    records = []
    indices = []
    for row in prediction_result_rows:
        indices.append(row.row_id)

        record = {}
        for value in row.values:
            record[value.name] = get_value(value)

        records.append(record)

    return pd.DataFrame(records, index=indices)


def test_df_to_prediction_rows():
    df = pd.DataFrame([[4, 9.2, "hi"]] * 3, columns=['int_value', 'double_value', 'string_value'])
    print(df_to_prediction_rows(df))


def test_prediction_rows_to_df():
    df = pd.DataFrame([[4, 1, "hi"]] * 3,
                      columns=['int_value', 'int_value_2', 'string_value'],
                      index=["0000", "1111", "2222"])
    pred_rows = df_to_prediction_result_rows(df)
    new_df = prediction_result_rows_to_df(pred_rows)
    assert new_df.equals(df)
