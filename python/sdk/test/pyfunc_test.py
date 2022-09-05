from unittest.mock import MagicMock

import grpc
from caraml.upi.v1 import upi_pb2
from mlflow.pyfunc import PythonModelContext


from merlin.model import PyFuncModel, PYFUNC_MODEL_INPUT_KEY, PYFUNC_EXTRA_ARGS_KEY
from merlin.protocol import Protocol
from merlin.pyfunc import PYFUNC_PROTOCOL_KEY, PYFUNC_GRPC_CONTEXT

request = {
    "instances": [[1, 2, 3, 4]]
}

response = {
    "predictions": [[1]]
}

headers = {
    "key" : "value"
}

upiv1_request = upi_pb2.PredictValuesRequest()
upiv1_response = upi_pb2.PredictValuesResponse()
grpc_context = {}

def test_http():
    pyfunc_model = PyFuncModel()
    pyfunc_model.infer = MagicMock(return_value=response)

    context = PythonModelContext(artifacts = {})

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

    context = PythonModelContext(artifacts = {})

    pyfunc_model.load_context(context)
    model_input = {
        PYFUNC_MODEL_INPUT_KEY: request,
        PYFUNC_EXTRA_ARGS_KEY: {
            "headers" : headers
        }
    }

    actual_resp = pyfunc_model.predict(context=context, model_input=model_input)
    assert actual_resp == response
    pyfunc_model.infer.assert_called_with(request, headers=headers)


def test_explicit_protocol():
    pyfunc_model = PyFuncModel()
    pyfunc_model.infer = MagicMock(return_value=response)

    context = PythonModelContext(artifacts = {})

    pyfunc_model.load_context(context)
    model_input = {
        PYFUNC_MODEL_INPUT_KEY: request,
        PYFUNC_PROTOCOL_KEY: Protocol.HTTP_JSON
    }

    actual_resp = pyfunc_model.predict(context=context, model_input=model_input)
    assert actual_resp == response
    pyfunc_model.infer.assert_called_with(request)


def test_upiv1():
    pyfunc_model = PyFuncModel()
    pyfunc_model.upiv1_infer = MagicMock(return_value=upiv1_response)

    context = PythonModelContext(artifacts = {})

    pyfunc_model.load_context(context)
    model_input = {
        PYFUNC_PROTOCOL_KEY: Protocol.UPI_V1,
        PYFUNC_MODEL_INPUT_KEY: upiv1_request,
        PYFUNC_GRPC_CONTEXT: grpc_context,
    }

    actual_resp = pyfunc_model.predict(context=context, model_input=model_input)
    assert actual_resp == upiv1_response
    pyfunc_model.upiv1_infer.assert_called_with(upiv1_request, grpc_context)