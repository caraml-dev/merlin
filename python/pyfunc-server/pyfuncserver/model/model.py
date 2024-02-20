# Copyright 2020 The Merlin Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import inspect
from enum import Enum

import grpc
from caraml.upi.v1 import upi_pb2

from merlin.protocol import Protocol
from merlin.pyfunc import PYFUNC_EXTRA_ARGS_KEY, PYFUNC_GRPC_CONTEXT, PYFUNC_MODEL_INPUT_KEY, PYFUNC_PROTOCOL_KEY, PyFuncOutput
from mlflow import pyfunc

from pyfuncserver.config import ModelManifest

NUM_OF_LATEST_PREDICT_FUNC_ARGS = 3


class PyFuncModelVersion(Enum):
    """
    PyFunc model version based on the merlin-sdk and mlflow version used in PyFunc model
    * LATEST -> using latest merlin-sdk or user doesn't specify merlin-sdk as dependency in PyFunc model
    * OLD_PYFUNC_LATEST_MLFLOW -> using older merlin-sdk (< 0.16) which doesn't have `predict(context, model_input)` signature,
        using latest version of mlflow (> 1.8).
    * OLD_PYFUNC_OLD_MLFLOW -> using older merlin-sdk (< 0.16) which doesn't have `predict(context, model_input)` signature,
        using older version of mlflow (<= 1.8)

    older version of merlin-sdk, merlin-sdk < 0.16
    older version of mlflow,  mlflow <= 1.8
    """
    LATEST = 'latest'
    OLD_PYFUNC_LATEST_MLFLOW = 'old_pyfunc_latest_mlflow'
    OLD_PYFUNC_OLD_MLFLOW = 'old_pyfunc_old_mlflow'


def _is_latest_pyfunc_model_version(func):
    full_args = inspect.getfullargspec(func)
    return len(full_args.args) == NUM_OF_LATEST_PREDICT_FUNC_ARGS


class PyFuncModel:  # pylint:disable=c-extension-no-member

    def __init__(self, model_manifest: ModelManifest):
        self.name = model_manifest.model_name
        self.version = model_manifest.model_version
        self.full_name = model_manifest.model_full_name
        self.artifact_dir = model_manifest.model_dir
        self.ready = False
        self.pyfunc_type = PyFuncModelVersion.LATEST

    def load(self):
        self._model = pyfunc.load_model(self.artifact_dir)
        # pylint:disable=c-extension-no-member
        # pylint:disable=attribute-defined-outside-init
        self.ready = True
        self.pyfunc_type = self._get_pyfunc_model_version()

    def _get_pyfunc_model_version(self):
        if hasattr(self._model, 'python_model'):
            if not _is_latest_pyfunc_model_version(self._model.python_model.predict):
                return PyFuncModelVersion.OLD_PYFUNC_OLD_MLFLOW
        elif hasattr(self._model, '_model_impl'):
            if not _is_latest_pyfunc_model_version(self._model._model_impl.python_model.predict):
                return PyFuncModelVersion.OLD_PYFUNC_LATEST_MLFLOW

        return PyFuncModelVersion.LATEST

    def predict(self, inputs: dict, **kwargs) -> PyFuncOutput:
        if self.pyfunc_type == PyFuncModelVersion.OLD_PYFUNC_LATEST_MLFLOW:
            # for case user specified old merlin-sdk as dependency and using mlflow without version specified
            return self._model._model_impl.python_model.predict(inputs, **kwargs)
        elif self.pyfunc_type == PyFuncModelVersion.OLD_PYFUNC_OLD_MLFLOW:
            # for case user specified old merlin-sdk as dependency and using old version of mlflow
            # that has `python_model` attribute on the model
            return self._model.python_model.predict(inputs, **kwargs)
        else:
            # for case user doesn't specify merlin-sdk as dependency
            model_inputs = {PYFUNC_MODEL_INPUT_KEY: inputs}
            if kwargs is not None:
                model_inputs[PYFUNC_EXTRA_ARGS_KEY] = kwargs

            return self._model.predict(model_inputs)

    def upiv1_predict(self, request: upi_pb2.PredictValuesRequest,
                      context: grpc.ServicerContext) -> PyFuncOutput:
        model_inputs = {
            PYFUNC_PROTOCOL_KEY: Protocol.UPI_V1,
            PYFUNC_MODEL_INPUT_KEY: request,
            PYFUNC_GRPC_CONTEXT: context,
        }
        return self._model.predict(model_inputs)
