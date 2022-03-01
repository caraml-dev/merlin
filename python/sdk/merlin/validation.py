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
from os.path import isdir, isfile, join


def validate_model_dir(model_type, model_dir):
    """
    Validates user-provided model directory based on file structure.
    For tensorflow models, checking is only done on the subdirectory
    with the largest version number.
    
    :param model_type: type of given model
    :param model_dir: directory containing serialised model file
    """
    from merlin.model import ModelType

    if not isdir(model_dir):
        raise ValueError(f"{model_dir} is not a directory")

    if model_type == ModelType.PYFUNC or \
            model_type == ModelType.PYFUNC_V2 or \
            model_type == ModelType.CUSTOM:
        return

    if model_type == ModelType.TENSORFLOW:
        # for tensorflow model it must have following structure
        # <version>
        # |- variables (directory)
        # |- saved_model.pb
        dirs = listdir(model_dir)
        if len(dirs) != 1:
            raise ValueError(f"{model_dir} must contain a version directory")

        version_dir = join(model_dir, dirs[0])
        variable_dir = join(version_dir, "variables")
        saved_model_path = join(version_dir, "saved_model.pb")
        if not isdir(variable_dir):
            raise ValueError(f"{variable_dir} is not found")

        if not isfile(saved_model_path):
            raise ValueError(f"{saved_model_path} is not found")

        return

    if model_type == ModelType.PYTORCH:
        # for Pytorch model it must have following structure
        # |- config
        # |   |- config.properties
        # |- model-store
        #     |- *.mar
        config_path = join(model_dir, "config", "config.properties")
        if not isfile(config_path):
            raise ValueError(f"{config_path} is not found")

        model_store_dir = join(model_dir, "model-store")
        if not any(fname.endswith('.mar') for fname in listdir(model_store_dir)):
            raise ValueError(f".mar file is not found in {model_store_dir}")

        return

    model_file_map = {
        ModelType.XGBOOST: ['model.bst'],
        ModelType.SKLEARN: ['model.joblib'],
        ModelType.ONNX: ['model.onnx']
    }
    files = listdir(model_dir)
    if not all([file in files for file in model_file_map[model_type]]):
        raise ValueError(f"{model_file_map[model_type]} is not found in {model_dir}")

