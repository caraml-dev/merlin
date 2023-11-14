import pytest
from test.test_http import sample_model_input, sample_model_output
from pyfuncserver.config import ModelManifest
from pyfuncserver.utils.converter import build_prediction_log, _build_model_input, _build_model_output
from merlin.pyfunc import PyFuncOutput

from caraml.upi.v1 import prediction_log_pb2

def test_build_prediction_log():
    pyfunc_output = PyFuncOutput(
        http_response = {"response": "ok"},
        model_input=sample_model_input,
        model_output=sample_model_output
    )
    manifest = ModelManifest(model_name="model_name", model_version="1", model_full_name="model_name_1", model_dir="/", project="project")
    got_prediction_log = build_prediction_log(pyfunc_output, manifest)
    exp_prediction_log = prediction_log_pb2.PredictionLog(
            prediction_id=sample_model_input.session_id,
            target_name="",
            project_name="project",
            model_name="model_name",
            model_version="1",
            input= _build_model_input(pyfunc_output.model_input),
            output= _build_model_output(pyfunc_output.model_output),
            table_schema_version=1
    )
    assert got_prediction_log == exp_prediction_log
