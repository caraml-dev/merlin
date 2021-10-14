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

import kfserving
from mlflow import pyfunc
from enum import Enum

EXTRA_ARGS_KEY = "__EXTRA_ARGS__"
MODEL_INPUT_KEY = "__INPUT__"


class PyFuncModelVersion(Enum):
    LATEST = 'latest'
    OLD_PYFUNC_LATEST_MLFLOW = 'old_pyfunc_latest_mlflow'
    OLD_PYFUNC_OLD_MLFLOW = 'old_pyfunc_old_mlflow'

class PyFuncModel(kfserving.KFModel):  # pylint:disable=c-extension-no-member

    def __init__(self, name: str, artifact_dir: str):
        super().__init__(name)
        self.name = name
        self.artifact_dir = artifact_dir
        self.ready = False
        self.pyfunc_type = PyFuncModelVersion.LATEST

    def load(self):
        self._model = pyfunc.load_model(self.artifact_dir)
        # pylint:disable=c-extension-no-member
        # pylint:disable=attribute-defined-outside-init
        self.ready = True
        self.pyfunc_type = self._get_pyfunc_model_version()

    def _get_pyfunc_model_version(self):
        try:
            self._model.predict({}) # try to call to determine the model type
            return PyFuncModelVersion.LATEST
        except TypeError as e:
            if "predict() takes 2 positional arguments but 3 were given" in str(e):
                if hasattr(self._model, 'python_model'):
                    return PyFuncModelVersion.OLD_PYFUNC_OLD_MLFLOW
                elif hasattr(self._model, '_model_impl'):
                    return PyFuncModelVersion.OLD_PYFUNC_LATEST_MLFLOW
                else:
                    raise Exception("no compatible predict() found")
            else:
                raise e
        except Exception:
            return PyFuncModelVersion.LATEST

    def predict(self, inputs: dict, **kwargs) -> dict:
        if self.pyfunc_type == PyFuncModelVersion.OLD_PYFUNC_LATEST_MLFLOW:
            # for case user specified old merlin-sdk as dependency and using mlflow without version specified
            return self._model._model_impl.python_model.predict(inputs, **kwargs)
        elif self.pyfunc_type == PyFuncModelVersion.OLD_PYFUNC_OLD_MLFLOW:
            # for case user specified old merlin-sdk as dependency and using old version of mlflow
            # that has `python_model` attribute on the model
            return self._model.python_model.predict(inputs, **kwargs)
        else:
            # for case user doesn't specify merlin-sdk as dependency
            model_inputs = { MODEL_INPUT_KEY: inputs}
            if kwargs is not None:
                model_inputs[EXTRA_ARGS_KEY] = kwargs

            return self._model.predict(model_inputs)
