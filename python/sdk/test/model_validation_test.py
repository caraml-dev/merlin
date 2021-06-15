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

from merlin.validation import validate_model_dir
from merlin.model import ModelType
import pytest


@pytest.mark.unit
def test_validate_model_dir():
    models = [modeltype for modeltype in ModelType]
    functions = [ModelType.PYFUNC, ModelType.PYTORCH, None]
    for input_model in models:
        for target_function in functions:
            valid_model_dir = f'test/{input_model.value}-model'
            invalid_model_dir = f'test/invalid-models/{input_model.value}-model-invalid'
            if target_function == input_model or \
                    (input_model
                     not in [ModelType.PYFUNC, ModelType.PYTORCH,
                             ModelType.PYFUNC_V2]
                     and target_function == None):
                assert validate_model_dir(input_model, target_function,
                                          valid_model_dir) == None
                if input_model != ModelType.PYFUNC and input_model != ModelType.CUSTOM:
                    with pytest.raises(Exception) as exc:
                        validate_model_dir(input_model, target_function,
                                           invalid_model_dir)
                    assert str(exc.value).startswith(
                        f'Provided {input_model.name} model directory should contain all of the following:')
