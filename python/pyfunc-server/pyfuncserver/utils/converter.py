from datetime import datetime
from typing import Optional

from caraml.upi.v1 import prediction_log_pb2
from google.protobuf import timestamp_pb2
from google.protobuf.struct_pb2 import Struct
from merlin.pyfunc import ModelInput, ModelOutput, PyFuncOutput
from pyfuncserver.config import ModelManifest


def build_prediction_log(
    pyfunc_output: PyFuncOutput, model_manifest: ModelManifest
) -> prediction_log_pb2.PredictionLog:
    model_input = _build_model_input(pyfunc_output.model_input)
    model_output = _build_model_output(pyfunc_output.model_output)

    session_id = pyfunc_output.get_session_id()
    proto_timestamp = timestamp_pb2.Timestamp()
    proto_timestamp.FromDatetime(datetime_now())
    prediction_log = prediction_log_pb2.PredictionLog(
        prediction_id=session_id,
        target_name="",  # TO-DO update this after schema is introduced
        project_name=model_manifest.project,
        model_name=model_manifest.model_name,
        model_version=model_manifest.model_version,
        input=model_input,
        output=model_output,
        table_schema_version=1,
        request_timestamp=proto_timestamp,
    )
    return prediction_log


def datetime_now() -> datetime:
    return datetime.now()


def _new_struct(dictionary: Optional[dict]):
    struct = None
    if dictionary is not None:
        struct = Struct()
        struct.update(dictionary)

    return struct


def _build_model_input(input: ModelInput) -> prediction_log_pb2.ModelInput:
    features_table = _new_struct(input.features_dict()) if input is not None else None
    entities_table = _new_struct(input.entities_dict()) if input is not None else None
    return prediction_log_pb2.ModelInput(
        features_table=features_table, entities_table=entities_table
    )


def _build_model_output(output: ModelOutput) -> prediction_log_pb2.ModelOutput:
    prediction_results_table = (
        _new_struct(output.predictions_dict()) if output is not None else None
    )
    return prediction_log_pb2.ModelOutput(
        prediction_results_table=prediction_results_table
    )
