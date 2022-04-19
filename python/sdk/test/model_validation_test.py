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
import re

@pytest.mark.parametrize(
    "model_type,model_dir,message",
    [
        (ModelType.PYTORCH, "test/invalid-models/pytorch-model-invalid", "test/invalid-models/pytorch-model-invalid/config/config.properties is not found"),
        (ModelType.TENSORFLOW, "test/invalid-models/tensorflow-model-invalid", "test/invalid-models/tensorflow-model-invalid/1/saved_model.pb is not found"),
        (ModelType.SKLEARN, "test/invalid-models/sklearn-model-invalid", "['model.joblib'] is not found in test/invalid-models/sklearn-model-invalid"),
        (ModelType.XGBOOST, "test/invalid-models/xgboost-model-invalid", "['model.bst'] is not found in test/invalid-models/xgboost-model-invalid"),
        (ModelType.ONNX, "test/invalid-models/onnx-model-invalid", "['model.onnx'] is not found in test/invalid-models/onnx-model-invalid")
    ]
)
@pytest.mark.unit
def test_invalid_model_dir(model_type, model_dir, message):
    with pytest.raises(ValueError, match=re.escape(message)):
        validate_model_dir(model_type, model_dir)


@pytest.mark.parametrize(
    "model_type,model_dir",
    [
        (ModelType.PYTORCH, "test/pytorch-model/pytorch-sample"),
        (ModelType.TENSORFLOW, "test/tensorflow-model"),
        (ModelType.SKLEARN, "test/sklearn-model"),
        (ModelType.XGBOOST, "test/xgboost-model"),
        (ModelType.ONNX, "test/onnx-model")
    ]
)
@pytest.mark.unit
def test_valid_model_dir(model_type, model_dir):
    validate_model_dir(model_type, model_dir)
