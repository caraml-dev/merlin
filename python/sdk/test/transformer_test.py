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
import os
from merlin.transformer import StandardTransformer, TransformerType
from merlin.resource_request import ResourceRequest


@pytest.mark.unit
def test_feast_enricher():
    transformer_config_path = os.path.join("test/transformer", "feast_enricher.yaml")
    transformer = StandardTransformer(config_file=transformer_config_path, enabled=False)
    assert transformer.env_vars == {'STANDARD_TRANSFORMER_CONFIG': '{"transformerConfig": {"feast": [{"project": "merlin", "entities": [{"name": "merlin_test_driver_id", "valueType": "STRING", "jsonPath": "$.driver_id"}], "features": [{"name": "merlin_test_driver_features:test_int32", "valueType": "INT32", "defaultValue": "0"}, {"name": "merlin_test_driver_features:test_float", "valueType": "FLOAT", "defaultValue": "0.0"}, {"name": "merlin_test_driver_features:test_double", "valueType": "DOUBLE", "defaultValue": "0.0"}, {"name": "merlin_test_driver_features:test_string", "valueType": "STRING", "defaultValue": ""}]}]}}'}
    assert not transformer.enabled
    assert transformer.command is None
    assert transformer.args is None
    assert transformer.transformer_type == TransformerType.STANDARD_TRANSFORMER


def test_feast_enricher_with_env_vars():
    transformer_config_path = os.path.join("test/transformer", "feast_enricher.yaml")
    resource = ResourceRequest(min_replica=1, max_replica=2, cpu_request="100m", memory_request="128Mi")
    transformer = StandardTransformer(config_file=transformer_config_path,
                                      enabled=True,
                                      resource_request=resource,
                                      env_vars={"MODEL_URL": "http://model.default"})
    assert transformer.env_vars == {'MODEL_URL': "http://model.default", 'STANDARD_TRANSFORMER_CONFIG': '{"transformerConfig": {"feast": [{"project": "merlin", "entities": [{"name": "merlin_test_driver_id", "valueType": "STRING", "jsonPath": "$.driver_id"}], "features": [{"name": "merlin_test_driver_features:test_int32", "valueType": "INT32", "defaultValue": "0"}, {"name": "merlin_test_driver_features:test_float", "valueType": "FLOAT", "defaultValue": "0.0"}, {"name": "merlin_test_driver_features:test_double", "valueType": "DOUBLE", "defaultValue": "0.0"}, {"name": "merlin_test_driver_features:test_string", "valueType": "STRING", "defaultValue": ""}]}]}}'}
    assert transformer.enabled
    assert transformer.command is None
    assert transformer.args is None
    assert transformer.resource_request.min_replica == 1
    assert transformer.resource_request.max_replica == 2
    assert transformer.resource_request.cpu_request == "100m"
    assert transformer.resource_request.memory_request == "128Mi"
    assert transformer.transformer_type == TransformerType.STANDARD_TRANSFORMER
