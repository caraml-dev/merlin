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
from merlin.model import PYFUNC_EXTRA_ARGS_KEY, PYFUNC_MODEL_INPUT_KEY

class PyFuncModel(kfserving.KFModel):  # pylint:disable=c-extension-no-member

    def __init__(self, name: str, artifact_dir: str):
        super().__init__(name)
        self.name = name
        self.artifact_dir = artifact_dir
        self.ready = False

    def load(self):
        self._model = pyfunc.load_model(self.artifact_dir)
        # pylint:disable=c-extension-no-member
        # pylint:disable=attribute-defined-outside-init
        self.ready = True

    def predict(self, inputs: dict, **kwargs) -> dict:
        model_inputs = { PYFUNC_MODEL_INPUT_KEY: inputs}
        if kwargs is not None:
            model_inputs[PYFUNC_EXTRA_ARGS_KEY] = kwargs

        return self._model.predict(model_inputs)
