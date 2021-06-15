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

from os import listdir
from os.path import isdir


def validate_model_dir(input_model_type, target_model_type, model_dir):
    """
    Validates user-provided model directory based on file structure.
    For tensorflow models, checking is only done on the subdirectory
    with the largest version number.
    
    :param input_model_type: type of given model
    :param target_model_type: type of supposed model, dependent on log_<model type>(...)
    :param model_dir: directory containing serialised model file
    """
    from merlin.model import ModelType

    if target_model_type == None and input_model_type == ModelType.TENSORFLOW:
        path_isdir = [isdir(f'{model_dir}/{path}') for path in listdir(model_dir)]
        if len(listdir(model_dir)) > 0 and all(path_isdir):
            model_dir = f'{model_dir}/{sorted(listdir(model_dir))[-1]}'

    if input_model_type != ModelType.PYFUNC and input_model_type != ModelType.PYFUNC_V2 and input_model_type != ModelType.CUSTOM:
        file_structure_reqs_map = {
            ModelType.XGBOOST: ['model.bst'],
            ModelType.TENSORFLOW: ['saved_model.pb', 'variables'],
            ModelType.SKLEARN: ['model.joblib'],
            ModelType.PYTORCH: ['model.pt'],
            ModelType.ONNX: ['model.onnx']
        }
        input_structure = listdir(model_dir)
        file_structure_req = file_structure_reqs_map[input_model_type]
        if not all([req in input_structure for req in file_structure_req]):
            raise Exception(
                f"Provided {input_model_type.name} model directory should contain all of the following: {file_structure_req}")
