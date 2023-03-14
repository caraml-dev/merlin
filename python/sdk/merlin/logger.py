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
from typing import Optional

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
class PredictionLoggerConfig:
    def __init__(self, enabled: bool, raw_features_table: str, entities_table: str) -> None:
        self._enabled = enabled
        self._raw_features_table = raw_features_table
        self._entities_table = entities_table

    @property
    def enabled(self):
        return self._enabled
    
    @property
    def raw_features_table(self):
        return self._raw_features_table
    
    @property
    def entities_table(self):
        return self._entities_table


@autostr
class Logger:
    logger_mode_mapping = {
        LoggerMode.ALL: client.LoggerMode.ALL,
        LoggerMode.REQUEST: client.LoggerMode.REQUEST,
        LoggerMode.RESPONSE: client.LoggerMode.RESPONSE
    }

    logger_mode_mapping_rev = {
        client.LoggerMode.ALL: LoggerMode.ALL,
        client.LoggerMode.REQUEST: LoggerMode.REQUEST,
        client.LoggerMode.RESPONSE: LoggerMode.RESPONSE,
    }

    def __init__(self, model: LoggerConfig = None, transformer: LoggerConfig = None, prediction: PredictionLoggerConfig = None):
        self._model = model
        self._transformer = transformer
        self._prediction = prediction

    @classmethod
    def from_logger_response(cls, response: client.Logger):
        if response is None:
            return Logger()
        model_config = None
        if response.model is not None:
            model_config = LoggerConfig(enabled=response.model.enabled, mode=cls._get_logger_mode_from_api_response(response.model.mode))
        transformer_config = None
        if response.transformer is not None:
            transformer_config = LoggerConfig(enabled=response.transformer.enabled,
                                              mode=cls._get_logger_mode_from_api_response(response.transformer.mode))
        prediction_config = None
        if response.prediction is not None:
            prediction_config = PredictionLoggerConfig(enabled=response.prediction.enabled,
                                                              raw_features_table=response.prediction.raw_features_table,
                                                              entities_table=response.prediction.entities_table)

        return Logger(model=model_config, transformer=transformer_config, prediction=prediction_config)

    @classmethod
    def _get_logger_mode_from_api_response(cls, mode_from_api_response):
        mode = cls.logger_mode_mapping_rev.get(mode_from_api_response)
        if mode is None:
            mode = LoggerMode.ALL
        return mode


    def to_logger_spec(self) -> Optional[client.Logger]:
        target_logger = None

        model_logger_config = None
        if self.model is not None:
            model_logger_config = client.LoggerConfig(
                enabled=self.model.enabled, mode=Logger.logger_mode_mapping[self.model.mode])

        transformer_logger_config = None
        if self.transformer is not None:
            transformer_logger_config = client.LoggerConfig(
                enabled=self.transformer.enabled, mode=Logger.logger_mode_mapping[self.transformer.mode])

        prediction_logger_config = None
        if self.prediction is not None:
            prediction_logger_config = client.PredictionLoggerConfig(
                enabled=self.prediction.enabled, raw_features_table= self.prediction.raw_features_table, entities_table=self.prediction.entities_table
            )

        if model_logger_config is not None or transformer_logger_config is not None or prediction_logger_config is not None:
            target_logger = client.Logger(model=model_logger_config, transformer=transformer_logger_config, prediction=prediction_logger_config)

        return target_logger


    @property
    def model(self):
        return self._model

    @property
    def transformer(self):
        return self._transformer
    
    @property
    def prediction(self):
        return self._prediction
