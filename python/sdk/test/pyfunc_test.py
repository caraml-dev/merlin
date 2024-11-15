from unittest.mock import MagicMock

import numpy as np
import pandas as pd
import pytest
from caraml.upi.v1 import table_pb2, type_pb2, upi_pb2
from merlin.model import (
    PYFUNC_EXTRA_ARGS_KEY,
    PYFUNC_MODEL_INPUT_KEY,
    PyFuncModel,
    PyFuncV3Model,
)
from merlin.protocol import Protocol
from merlin.pyfunc import (
    PYFUNC_GRPC_CONTEXT,
    PYFUNC_PROTOCOL_KEY,
    ModelInput,
    ModelOutput,
    PyFuncOutput,
    Values,
)
from mlflow.pyfunc import PythonModelContext

request = {"instances": [[1, 2, 3, 4]]}

response = {"predictions": [[1]]}

headers = {"key": "value"}

upiv1_request = upi_pb2.PredictValuesRequest()
upiv1_response = upi_pb2.PredictValuesResponse()
grpc_context = {}

pyfunc_model_input = ModelInput(
    features=Values(
        columns=["featureA", "featureB", "featureC"],
        data=[[0.1, 0.2, "debit"], [0.12, 0.4, "cc"], [0.42, 0.2, "debit"]],
    ),
    entities=Values(
        columns=["order_id", "customer_id"],
        data=[["order1", "111"], ["order1", "112"], ["order1", "113"]],
    ),
    prediction_ids=["prediction1", "prediction2", "prediction3"],
)
pyfunc_model_output = ModelOutput(
    predictions=Values(
        columns=["prediction_score", "prediction_label"],
        data=[[0.95, "complete"], [0.43, "incomplete"], [0.59, "complete"]],
    ),
    prediction_ids=["prediction1", "prediction2", "prediction3"],
)


def test_http():
    pyfunc_model = PyFuncModel()
    pyfunc_model.infer = MagicMock(return_value=response)

    context = PythonModelContext(artifacts={})

    pyfunc_model.load_context(context)
    model_input = {
        PYFUNC_MODEL_INPUT_KEY: request,
    }

    actual_resp = pyfunc_model.predict(context=context, model_input=model_input)
    assert actual_resp == response
    pyfunc_model.infer.assert_called_with(request)


def test_http_headers():
    pyfunc_model = PyFuncModel()
    pyfunc_model.infer = MagicMock(return_value=response)

    context = PythonModelContext(artifacts={})

    pyfunc_model.load_context(context)
    model_input = {
        PYFUNC_MODEL_INPUT_KEY: request,
        PYFUNC_EXTRA_ARGS_KEY: {"headers": headers},
    }

    actual_resp = pyfunc_model.predict(context=context, model_input=model_input)
    assert actual_resp == response
    pyfunc_model.infer.assert_called_with(request, headers=headers)


def test_explicit_protocol():
    pyfunc_model = PyFuncModel()
    pyfunc_model.infer = MagicMock(return_value=response)

    context = PythonModelContext(artifacts={})

    pyfunc_model.load_context(context)
    model_input = {
        PYFUNC_MODEL_INPUT_KEY: request,
        PYFUNC_PROTOCOL_KEY: Protocol.HTTP_JSON,
    }

    actual_resp = pyfunc_model.predict(context=context, model_input=model_input)
    assert actual_resp == response
    pyfunc_model.infer.assert_called_with(request)


def test_upiv1():
    pyfunc_model = PyFuncModel()
    pyfunc_model.upiv1_infer = MagicMock(return_value=upiv1_response)

    context = PythonModelContext(artifacts={})

    pyfunc_model.load_context(context)
    model_input = {
        PYFUNC_PROTOCOL_KEY: Protocol.UPI_V1,
        PYFUNC_MODEL_INPUT_KEY: upiv1_request,
        PYFUNC_GRPC_CONTEXT: grpc_context,
    }

    actual_resp = pyfunc_model.predict(context=context, model_input=model_input)
    assert actual_resp == upiv1_response
    pyfunc_model.upiv1_infer.assert_called_with(upiv1_request, grpc_context)


def test_pyfuncv3_rest():
    pyfunc_model = PyFuncV3Model()
    http_response = {
        "columns": ["prediction_score", "prediction_label"],
        "data": [[0.95, "complete"], [0.43, "incomplete"], [0.59, "complete"]],
    }
    pyfunc_model.preprocess = MagicMock(return_value=pyfunc_model_input)
    pyfunc_model.infer = MagicMock(return_value=pyfunc_model_output)
    pyfunc_model.postprocess = MagicMock(return_value=http_response)
    context = PythonModelContext(artifacts={})

    pyfunc_model.load_context(context)
    model_input = {
        PYFUNC_MODEL_INPUT_KEY: request,
    }

    exp_response = PyFuncOutput(
        http_response=http_response,
        model_input=pyfunc_model_input,
        model_output=pyfunc_model_output,
    )

    actual_resp = pyfunc_model.predict(context=context, model_input=model_input)
    assert actual_resp == exp_response
    pyfunc_model.preprocess.assert_called_with(model_input[PYFUNC_MODEL_INPUT_KEY])
    pyfunc_model.infer.assert_called_with(pyfunc_model_input)
    pyfunc_model.postprocess.assert_called_with(
        pyfunc_model_output, model_input[PYFUNC_MODEL_INPUT_KEY]
    )


def test_pyfuncv3_upi():
    pyfunc_model = PyFuncV3Model()
    upi_response = upi_pb2.PredictValuesResponse(
        prediction_result_table=table_pb2.Table(
            name="prediction_result_table",
            columns=[
                table_pb2.Column(name="prediction_score", type=type_pb2.TYPE_DOUBLE),
                table_pb2.Column(name="prediction_label", type=type_pb2.TYPE_STRING),
            ],
            rows=[
                table_pb2.Row(
                    row_id="0",
                    values=[
                        table_pb2.Value(double_value=0.95),
                        table_pb2.Value(string_value="complete"),
                    ],
                ),
                table_pb2.Row(
                    row_id="1",
                    values=[
                        table_pb2.Value(double_value=0.43),
                        table_pb2.Value(string_value="incomplete"),
                    ],
                ),
                table_pb2.Row(
                    row_id="2",
                    values=[
                        table_pb2.Value(double_value=0.59),
                        table_pb2.Value(string_value="complete"),
                    ],
                ),
            ],
        )
    )
    pyfunc_model.upiv1_preprocess = MagicMock(return_value=pyfunc_model_input)
    pyfunc_model.infer = MagicMock(return_value=pyfunc_model_output)
    pyfunc_model.upiv1_postprocess = MagicMock(return_value=upi_response)
    context = PythonModelContext(artifacts={})

    pyfunc_model.load_context(context)
    model_input = {
        PYFUNC_PROTOCOL_KEY: Protocol.UPI_V1,
        PYFUNC_MODEL_INPUT_KEY: upiv1_request,
        PYFUNC_GRPC_CONTEXT: grpc_context,
    }

    exp_response = PyFuncOutput(
        upi_response=upi_response,
        model_input=pyfunc_model_input,
        model_output=pyfunc_model_output,
    )

    actual_resp = pyfunc_model.predict(context=context, model_input=model_input)
    assert actual_resp == exp_response
    pyfunc_model.upiv1_preprocess.assert_called_with(
        model_input[PYFUNC_MODEL_INPUT_KEY], grpc_context
    )
    pyfunc_model.infer.assert_called_with(pyfunc_model_input)
    pyfunc_model.upiv1_postprocess.assert_called_with(
        pyfunc_model_output, model_input[PYFUNC_MODEL_INPUT_KEY]
    )


@pytest.mark.parametrize(
    "features,row_ids,use_deprecated_prediction_ids,expected",
    [
        (
            Values(
                columns=["col1", "col2", "col3"],
                data=[[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]],
            ),
            ["prediction_1", "prediction_2", "prediction3"],
            False,
            {
                "columns": ["col1", "col2", "col3"],
                "data": [[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]],
                "row_ids": ["prediction_1", "prediction_2", "prediction3"],
            },
        ),
        (
            Values(
                columns=["col1", "col2", "col3"],
                data=np.array([[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]]),
            ),
            ["prediction_1", "prediction_2", "prediction3"],
            False,
            {
                "columns": ["col1", "col2", "col3"],
                "data": [[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]],
                "row_ids": ["prediction_1", "prediction_2", "prediction3"],
            },
        ),
        (
            pd.DataFrame(
                [[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]], columns=["col1", "col2", "col3"]
            ),
            ["prediction_1", "prediction_2", "prediction3"],
            False,
            {
                "columns": ["col1", "col2", "col3"],
                "data": [[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]],
                "row_ids": ["prediction_1", "prediction_2", "prediction3"],
            },
        ),
        (
            Values(
                columns=["col1", "col2", "col3"],
                data=[[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]],
            ),
            ["prediction_1", "prediction_2", "prediction3"],
            True,
            {
                "columns": ["col1", "col2", "col3"],
                "data": [[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]],
                "row_ids": ["prediction_1", "prediction_2", "prediction3"],
            },
        ),
        (
            Values(
                columns=["col1", "col2", "col3"],
                data=np.array([[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]]),
            ),
            ["prediction_1", "prediction_2", "prediction3"],
            True,
            {
                "columns": ["col1", "col2", "col3"],
                "data": [[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]],
                "row_ids": ["prediction_1", "prediction_2", "prediction3"],
            },
        ),
        (
            pd.DataFrame(
                [[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]], columns=["col1", "col2", "col3"]
            ),
            ["prediction_1", "prediction_2", "prediction3"],
            True,
            {
                "columns": ["col1", "col2", "col3"],
                "data": [[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]],
                "row_ids": ["prediction_1", "prediction_2", "prediction3"],
            },
        ),
    ],
)
def test_model_input_features(
    features, row_ids, use_deprecated_prediction_ids, expected
):
    model_input = ModelInput(features=features, row_ids=row_ids)
    if use_deprecated_prediction_ids:
        model_input = ModelInput(features=features, prediction_ids=row_ids)
    assert model_input.features_dict() == expected


@pytest.mark.parametrize(
    "entities,row_ids,use_deprecated_prediction_ids,expected",
    [
        (
            Values(
                columns=["col1", "col2", "col3"],
                data=[[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]],
            ),
            ["prediction_1", "prediction_2"],
            False,
            {
                "columns": ["col1", "col2", "col3"],
                "data": [[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]],
                "row_ids": ["prediction_1", "prediction_2"],
            },
        ),
        (
            Values(
                columns=["col1", "col2", "col3"],
                data=np.array([[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]]),
            ),
            ["prediction_1", "prediction_2"],
            False,
            {
                "columns": ["col1", "col2", "col3"],
                "data": [[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]],
                "row_ids": ["prediction_1", "prediction_2"],
            },
        ),
        (
            pd.DataFrame(
                [[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]], columns=["col1", "col2", "col3"]
            ),
            ["prediction_1", "prediction_2"],
            False,
            {
                "columns": ["col1", "col2", "col3"],
                "data": [[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]],
                "row_ids": ["prediction_1", "prediction_2"],
            },
        ),
        (None, None, False, None),
        (
            Values(
                columns=["col1", "col2", "col3"],
                data=[[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]],
            ),
            ["prediction_1", "prediction_2"],
            True,
            {
                "columns": ["col1", "col2", "col3"],
                "data": [[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]],
                "row_ids": ["prediction_1", "prediction_2"],
            },
        ),
        (
            Values(
                columns=["col1", "col2", "col3"],
                data=np.array([[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]]),
            ),
            ["prediction_1", "prediction_2"],
            True,
            {
                "columns": ["col1", "col2", "col3"],
                "data": [[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]],
                "row_ids": ["prediction_1", "prediction_2"],
            },
        ),
        (
            pd.DataFrame(
                [[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]], columns=["col1", "col2", "col3"]
            ),
            ["prediction_1", "prediction_2"],
            True,
            {
                "columns": ["col1", "col2", "col3"],
                "data": [[0.1, 0.2, 0.3], [0.2, 0.25, 0.3]],
                "row_ids": ["prediction_1", "prediction_2"],
            },
        ),
        (None, None, True, None),
    ],
)
def test_model_input_entities(
    entities, row_ids, use_deprecated_prediction_ids, expected
):
    model_input = ModelInput(
        features=Values(columns=["col1"], data=[[0.2], [0.3]]),
        entities=entities,
        row_ids=row_ids,
    )
    if use_deprecated_prediction_ids:
        model_input = ModelInput(
            features=Values(columns=["col1"], data=[[0.2], [0.3]]),
            entities=entities,
            prediction_ids=row_ids,
        )
    assert model_input.entities_dict() == expected


@pytest.mark.parametrize(
    "predictions,row_ids,use_deprecated_prediction_ids,expected",
    [
        (
            Values(columns=["prediction_score"], data=[[0.1], [0.2]]),
            ["prediction_1", "prediction_2"],
            False,
            {
                "columns": ["prediction_score"],
                "data": [[0.1], [0.2]],
                "row_ids": ["prediction_1", "prediction_2"],
            },
        ),
        (
            Values(columns=["prediction_score"], data=np.array([[0.1], [0.2]])),
            ["prediction_1", "prediction_2"],
            False,
            {
                "columns": ["prediction_score"],
                "data": [[0.1], [0.2]],
                "row_ids": ["prediction_1", "prediction_2"],
            },
        ),
        (
            pd.DataFrame([[0.1], [0.2]], columns=["prediction_score"]),
            ["prediction_1", "prediction_2"],
            False,
            {
                "columns": ["prediction_score"],
                "data": [[0.1], [0.2]],
                "row_ids": ["prediction_1", "prediction_2"],
            },
        ),
        (None, None, False, None),
        (
            Values(columns=["prediction_score"], data=[[0.1], [0.2]]),
            ["prediction_1", "prediction_2"],
            True,
            {
                "columns": ["prediction_score"],
                "data": [[0.1], [0.2]],
                "row_ids": ["prediction_1", "prediction_2"],
            },
        ),
        (
            Values(columns=["prediction_score"], data=np.array([[0.1], [0.2]])),
            ["prediction_1", "prediction_2"],
            True,
            {
                "columns": ["prediction_score"],
                "data": [[0.1], [0.2]],
                "row_ids": ["prediction_1", "prediction_2"],
            },
        ),
        (
            pd.DataFrame([[0.1], [0.2]], columns=["prediction_score"]),
            ["prediction_1", "prediction_2"],
            True,
            {
                "columns": ["prediction_score"],
                "data": [[0.1], [0.2]],
                "row_ids": ["prediction_1", "prediction_2"],
            },
        ),
        (None, None, True, None),
    ],
)
def test_model_output_predictions(
    predictions, row_ids, use_deprecated_prediction_ids, expected
):
    model_output = ModelOutput(predictions=predictions, row_ids=row_ids)
    if use_deprecated_prediction_ids:
        model_output = ModelOutput(predictions=predictions, prediction_ids=row_ids)
    assert model_output.predictions_dict() == expected
