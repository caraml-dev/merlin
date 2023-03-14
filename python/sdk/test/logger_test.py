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

import pytest
import client
from merlin.logger import Logger,LoggerMode,LoggerConfig,PredictionLoggerConfig


@pytest.mark.unit
def test_from_logger_response():
    logger_response = client.Logger(model=client.LoggerConfig(enabled=True, mode=client.LoggerMode.RESPONSE))
    result = Logger.from_logger_response(logger_response)
    expected_result = Logger(model=LoggerConfig(enabled=True, mode=LoggerMode.RESPONSE))
    assert result.model is not None
    assert result.model.enabled == expected_result.model.enabled
    assert result.model.mode == expected_result.model.mode
    assert result.transformer is None
    assert result.prediction is None

    logger_response = client.Logger(model=client.LoggerConfig(enabled=False, mode=""))
    result = Logger.from_logger_response(logger_response)
    expected_result = Logger(model=LoggerConfig(enabled=False, mode=LoggerMode.ALL))
    assert result.model is not None
    assert result.model.enabled == expected_result.model.enabled
    assert result.model.mode == expected_result.model.mode
    assert result.transformer is None
    assert result.prediction is None

    logger_response = client.Logger(transformer=client.LoggerConfig(enabled=True, mode=client.LoggerMode.REQUEST))
    result = Logger.from_logger_response(logger_response)
    expected_result = Logger(transformer=LoggerConfig(enabled=True, mode=LoggerMode.REQUEST))
    assert result.transformer is not None
    assert result.transformer.enabled == expected_result.transformer.enabled
    assert result.transformer.mode == expected_result.transformer.mode
    assert result.model is None
    assert result.prediction is None

    logger_response = client.Logger(model=client.LoggerConfig(enabled=True, mode=client.LoggerMode.ALL),
                                    transformer=client.LoggerConfig(enabled=True, mode=client.LoggerMode.ALL))
    result = Logger.from_logger_response(logger_response)
    expected_result = Logger(model=LoggerConfig(enabled=True, mode=LoggerMode.ALL),
                             transformer=LoggerConfig(enabled=True, mode=LoggerMode.ALL))
    assert result.transformer is not None
    assert result.transformer.enabled == expected_result.transformer.enabled
    assert result.transformer.mode == expected_result.transformer.mode
    assert result.model is not None
    assert result.model.enabled == expected_result.model.enabled
    assert result.model.mode == expected_result.model.mode
    assert result.prediction is None

    result = Logger.from_logger_response(None)
    assert result.model is None
    assert result.transformer is None

    logger_response = client.Logger(model=client.LoggerConfig(enabled=True, mode=client.LoggerMode.ALL),
                                    transformer=client.LoggerConfig(enabled=True, mode=client.LoggerMode.ALL),
                                    prediction=client.PredictionLoggerConfig(enabled=True, raw_features_table="rawFeatures", entities_table="entities"))
    result = Logger.from_logger_response(logger_response)
    expected_result = Logger(model=LoggerConfig(enabled=True, mode=LoggerMode.ALL),
                             transformer=LoggerConfig(enabled=True, mode=LoggerMode.ALL),
                             prediction=PredictionLoggerConfig(enabled=True, raw_features_table="rawFeatures", entities_table="entities"))
    assert result.transformer is not None
    assert result.transformer.enabled == expected_result.transformer.enabled
    assert result.transformer.mode == expected_result.transformer.mode
    assert result.model is not None
    assert result.model.enabled == expected_result.model.enabled
    assert result.model.mode == expected_result.model.mode
    assert result.prediction is not None
    assert result.prediction.enabled == expected_result.prediction.enabled
    assert result.prediction.raw_features_table == expected_result.prediction.raw_features_table
    assert result.prediction.entities_table == expected_result.prediction.entities_table



@pytest.mark.unit
def test_to_logger_spec():
    logger = Logger(model=LoggerConfig(enabled=False, mode=LoggerMode.REQUEST))
    result = logger.to_logger_spec()
    expected_result = client.Logger(model=client.LoggerConfig(enabled=False, mode=client.LoggerMode.REQUEST))
    assert result.model is not None
    assert result.model.enabled == expected_result.model.enabled
    assert result.model.mode == expected_result.model.mode
    assert result.transformer is None
    assert result.prediction is None

    logger = Logger(transformer=LoggerConfig(enabled=True, mode=LoggerMode.RESPONSE))
    result = logger.to_logger_spec()
    expected_result = client.Logger(transformer=client.LoggerConfig(enabled=True, mode=client.LoggerMode.RESPONSE))
    assert result.transformer is not None
    assert result.transformer.enabled == expected_result.transformer.enabled
    assert result.transformer.mode == expected_result.transformer.mode
    assert result.model is None
    assert result.prediction is None


    logger = Logger(model=LoggerConfig(enabled=True, mode=LoggerMode.ALL),
                    transformer=LoggerConfig(enabled=True, mode=LoggerMode.ALL))
    result = logger.to_logger_spec()
    expected_result = client.Logger(model=client.LoggerConfig(enabled=True, mode=client.LoggerMode.ALL),
                                    transformer=client.LoggerConfig(enabled=True, mode=client.LoggerMode.ALL))
    assert result.transformer is not None
    assert result.transformer.enabled == expected_result.transformer.enabled
    assert result.transformer.mode == expected_result.transformer.mode
    assert result.model is not None
    assert result.model.enabled == expected_result.model.enabled
    assert result.model.mode == expected_result.model.mode
    assert result.prediction is None


    logger = Logger(model=LoggerConfig(enabled=True, mode=LoggerMode.ALL),
                    transformer=LoggerConfig(enabled=True, mode=LoggerMode.ALL),
                    prediction=PredictionLoggerConfig(enabled=True, raw_features_table="rawFeatures", entities_table="entities"))
    result = logger.to_logger_spec()
    expected_result = client.Logger(model=client.LoggerConfig(enabled=True, mode=client.LoggerMode.ALL),
                                    transformer=client.LoggerConfig(enabled=True, mode=client.LoggerMode.ALL),
                                    prediction=client.PredictionLoggerConfig(enabled=True, raw_features_table="rawFeatures", entities_table="entities"))
    assert result.transformer is not None
    assert result.transformer.enabled == expected_result.transformer.enabled
    assert result.transformer.mode == expected_result.transformer.mode
    assert result.model is not None
    assert result.model.enabled == expected_result.model.enabled
    assert result.model.mode == expected_result.model.mode
    assert result.prediction is not None
    assert result.prediction.enabled == expected_result.prediction.enabled
    assert result.prediction.raw_features_table == expected_result.prediction.raw_features_table
    assert result.prediction.entities_table == expected_result.prediction.entities_table

    logger = Logger()
    result = logger.to_logger_spec()
    assert result is None


