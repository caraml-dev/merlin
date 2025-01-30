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

from typing import Dict, Optional, Any, List

from merlin.resource_request import ResourceRequest
from merlin.mounted_mlp_secret import MountedMLPSecret
from merlin.util import autostr
from merlin import fluent

from enum import Enum

import client
import yaml
import json


class TransformerType(Enum):
    CUSTOM_TRANSFORMER = "custom"
    STANDARD_TRANSFORMER = "standard"


@autostr
class Transformer:
    StandardTransformerConfigKey = "STANDARD_TRANSFORMER_CONFIG"

    def __init__(
        self,
        image: str,
        id: str = "",
        enabled: bool = True,
        command: Optional[str] = None,
        args: Optional[str] = None,
        resource_request: Optional[ResourceRequest] = None,
        env_vars: Optional[Dict[str, str]] = None,
        secrets: List[MountedMLPSecret] = None,
        transformer_type: TransformerType = TransformerType.CUSTOM_TRANSFORMER,
    ):
        self._id = id
        self._image = image
        self._enabled = enabled
        self._command = command
        self._args = args
        self._resource_request = resource_request
        self._env_vars = env_vars
        self._secrets = secrets
        self._transformer_type = transformer_type

    @property
    def image(self) -> str:
        return self._image

    @property
    def id(self) -> str:
        return self._id

    @property
    def enabled(self) -> bool:
        return self._enabled

    @property
    def command(self) -> Optional[str]:
        return self._command

    @property
    def args(self) -> Optional[str]:
        return self._args

    @property
    def resource_request(self) -> Optional[ResourceRequest]:
        return self._resource_request

    @property
    def env_vars(self) -> Optional[Dict[str, str]]:
        return self._env_vars

    @property
    def secrets(self) -> Optional[List[MountedMLPSecret]]:
        return self._secrets

    @property
    def transformer_type(self) -> TransformerType:
        return self._transformer_type


class StandardTransformer(Transformer):
    def __init__(
        self,
        config_file: str,
        enabled: bool = True,
        resource_request: ResourceRequest = None,
        env_vars: Dict[str, str] = None,
    ):
        self._load_transformer_config(config_file)
        transformer_env_var = self._transformer_config_env_var()
        merged_env_vars = env_vars or {}
        merged_env_vars = {**merged_env_vars, **transformer_env_var}
        super().__init__(
            image="",
            enabled=enabled,
            resource_request=resource_request,
            env_vars=merged_env_vars,
            transformer_type=TransformerType.STANDARD_TRANSFORMER,
        )

    def _load_transformer_config(self, config_file: str):
        with open(config_file, "r") as stream:
            self.transformer_config = yaml.safe_load(stream)

    def _transformer_config_env_var(self):
        config_json_string = json.dumps(self.transformer_config)
        return {self.StandardTransformerConfigKey: config_json_string}

    def simulate(
        self,
        payload: Dict,
        headers: Optional[Dict[Any, Any]] = None,
        model_prediction_config: Optional[Dict[Any, Any]] = None,
        protocol: str = "HTTP_JSON",
        exclude_tracing: bool = False,
    ) -> Dict:
        fluent._check_active_client()
        if not fluent._merlin_client:
            raise Exception("Merlin client is not initialized")

        response = fluent._merlin_client.standard_transformer_simulate(
            payload=payload,
            headers=headers,
            config=self.transformer_config,
            model_prediction_config=model_prediction_config,
            protocol=protocol,
        )

        # if exclude tracing delete key operation_tracing
        response = response.to_dict()
        if exclude_tracing:
            del response["operation_tracing"]

        return response
