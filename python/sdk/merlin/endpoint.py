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
from typing import Dict, Optional, List

import client
from merlin.autoscaling import (RAW_DEPLOYMENT_DEFAULT_AUTOSCALING_POLICY,
                                SERVERLESS_DEFAULT_AUTOSCALING_POLICY,
                                AutoscalingPolicy, MetricsType)
from merlin.deployment_mode import DeploymentMode
from merlin.environment import Environment
from merlin.logger import Logger
from merlin.protocol import Protocol
from merlin.util import autostr, get_url
from merlin.resource_request import ResourceRequest
from merlin.mounted_mlp_secret import MountedMLPSecret
from merlin.transformer import Transformer, TransformerType
from merlin.model_observability import ModelObservability
from merlin.util import extract_optional_value_with_default

class Status(Enum):
    PENDING = 'pending'
    RUNNING = 'running'
    SERVING = 'serving'
    FAILED = 'failed'
    TERMINATED = 'terminated'


@autostr
class VersionEndpoint:
    def __init__(self, endpoint: client.VersionEndpoint, log_url: str = None):
        self._protocol = Protocol.HTTP_JSON
        if endpoint.protocol:
            self._protocol = Protocol(endpoint.protocol)

        self._url = endpoint.url
        if self._protocol == Protocol.HTTP_JSON and endpoint.url is not None and ":predict" not in endpoint.url:
            self._url = f"{endpoint.url}:predict"

        self._status = Status(endpoint.status)
        self._id = endpoint.id
        self._environment_name = endpoint.environment_name
        self._environment = Environment(endpoint.environment) if endpoint.environment is not None else None
        self._env_vars = endpoint.env_vars
        self._secrets = [MountedMLPSecret.from_response(secret) for secret in endpoint.secrets] if endpoint.secrets is not None else None
        self._logger = Logger.from_logger_response(endpoint.logger)
        self._resource_request = ResourceRequest.from_response(endpoint.resource_request) if endpoint.resource_request is not None else None
        self._deployment_mode = DeploymentMode.SERVERLESS if not endpoint.deployment_mode \
            else DeploymentMode(endpoint.deployment_mode)
        self._transformer: Optional[Transformer] = None

        if endpoint.autoscaling_policy is None:
            if self._deployment_mode == DeploymentMode.SERVERLESS:
                self._autoscaling_policy = SERVERLESS_DEFAULT_AUTOSCALING_POLICY
            else:
                self._autoscaling_policy = RAW_DEPLOYMENT_DEFAULT_AUTOSCALING_POLICY
        else:
            self._autoscaling_policy = AutoscalingPolicy(metrics_type=MetricsType(endpoint.autoscaling_policy.metrics_type),
                                                         target_value=endpoint.autoscaling_policy.target_value)

        transformer = endpoint.transformer
        if transformer is not None:
            image = transformer.image if transformer.image is not None else ""
            transformer_type_value = extract_optional_value_with_default(transformer.transformer_type, TransformerType.STANDARD_TRANSFORMER.value)
            transformer_type = TransformerType(transformer_type_value)
            
            transformer_request = None
            if transformer.resource_request is not None:
                transformer_request = ResourceRequest(
                    min_replica=transformer.resource_request.min_replica,
                    max_replica=transformer.resource_request.max_replica,
                    cpu_request=transformer.resource_request.cpu_request,
                    cpu_limit=transformer.resource_request.cpu_limit,
                    memory_request=transformer.resource_request.memory_request
                )

            env_vars: Dict[str, str] = {}
            if transformer.env_vars is not None:
                for env_var in transformer.env_vars:
                    if env_var.name is not None and env_var.value is not None:
                        env_vars[env_var.name] = env_var.value

            secrets: List[MountedMLPSecret] = []
            if transformer.secrets is not None:
                for secret in transformer.secrets:
                    secrets.append(
                        MountedMLPSecret(
                            mlp_secret_name=secret.mlp_secret_name,
                            env_var_name=secret.env_var_name,
                        )
                    )

            self._transformer = Transformer(
                image, 
                id=extract_optional_value_with_default(transformer.id, ""),
                enabled=extract_optional_value_with_default(transformer.enabled, False), 
                command=transformer.command,
                args=transformer.args, 
                transformer_type=transformer_type,
                resource_request=transformer_request, 
                env_vars=env_vars,
                secrets=secrets,
            )

        if log_url is not None:
            self._log_url = log_url

        self._enable_model_observability = extract_optional_value_with_default(endpoint.enable_model_observability, False)
        if endpoint.model_observability is not None:
            model_observability = ModelObservability.from_model_observability_response(endpoint.model_observability)
            self._model_observability = extract_optional_value_with_default(model_observability, None)

    @property
    def url(self):
        return self._url

    @property
    def status(self) -> Status:
        return self._status

    @property
    def id(self) -> Optional[str]:
        return self._id

    @property
    def environment_name(self) -> Optional[str]:
        return self._environment_name

    @property
    def environment(self) -> Optional[Environment]:
        return self._environment

    @property
    def env_vars(self) -> Dict[str, str]:
        env_vars = {}
        if self._env_vars:
            for ev in self._env_vars:
                if ev.name is not None and ev.value is not None:
                    env_vars[ev.name] = ev.value
        return env_vars

    @property
    def secrets(self) -> Optional[List[MountedMLPSecret]]:
        return self._secrets

    @property
    def logger(self) -> Logger:
        return self._logger

    @property
    def log_url(self) -> str:
        return self._log_url

    @property
    def deployment_mode(self) -> DeploymentMode:
        return self._deployment_mode

    @property
    def autoscaling_policy(self) -> AutoscalingPolicy:
        return self._autoscaling_policy

    @property
    def protocol(self) -> Protocol:
        return self._protocol
    
    @property
    def resource_request(self) -> Optional[ResourceRequest]:
        return self._resource_request
    
    @property
    def transformer(self) -> Optional[Transformer]:
        return self._transformer
    
    @property
    def enable_model_observability(self) -> bool:
        return self._enable_model_observability
    
    @property
    def model_observability(self) -> Optional[client.ModelObservability]:
        return self._model_observability

    def _repr_html_(self):
        return f"""<a href="{self._url}">{self._url}</a>"""


@autostr
class ModelEndpoint:
    def __init__(self, endpoint: client.ModelEndpoint):
        self._protocol = Protocol.HTTP_JSON
        if endpoint.protocol is not None:
            self._protocol = Protocol(endpoint.protocol)

        if self._protocol == Protocol.HTTP_JSON:
            self._url = get_url(f"{endpoint.url}/v1/predict")
        else:
            self._url = endpoint.url
        self._status = Status(endpoint.status)
        self._id = extract_optional_value_with_default(endpoint.id, 0)
        self._environment_name = endpoint.environment_name
        self._environment = Environment(endpoint.environment) if endpoint.environment is not None else None


    @property
    def url(self):
        return self._url

    @property
    def status(self) -> Status:
        return self._status

    @property
    def id(self) -> int:
        return self._id

    @property
    def environment_name(self) -> Optional[str]:
        return self._environment_name

    @property
    def environment(self) -> Optional[Environment]:
        return self._environment

    @property
    def protocol(self) -> Protocol:
        return self._protocol

    def _repr_html_(self):
        return f"""<a href="{self._url}">{self._url}</a>"""
