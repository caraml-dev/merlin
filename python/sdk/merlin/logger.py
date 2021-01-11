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

from enum import Enum
import client
from merlin.util import autostr


class LoggerMode(Enum):
    ALL = 'all'
    REQUEST = 'request'
    RESPONSE = 'response'


@autostr
class LoggerConfig:
    def __init__(self, enabled: bool, mode: LoggerMode):
        self._enabled = enabled
        self._mode = mode

    @property
    def enabled(self):
        return self._enabled

    @property
    def mode(self):
        return self._mode

@autostr
class Logger:
    def __init__(self, model: LoggerConfig = None, transformer: LoggerConfig = None):
        self._model = model
        self._transformer = transformer

    @classmethod
    def from_logger_response(response: client.Logger):
        model_config = None
        if response.model is not None:
            model_config = LoggerConfig(enabled=response.model.enabled, mode=LoggerMode[response.model.mode])
        transformer_config = None
        if response.transformer is not None:
            transformer_config = LoggerConfig(enabled=response.transformer.enabled,
                                              mode=LoggerMode[response.transformer.mode])

        return Logger(model=model_config, transformer=transformer_config)

    @property
    def model(self):
        return self._model

    @property
    def transformer(self):
        return self._transformer



