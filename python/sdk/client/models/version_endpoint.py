# coding: utf-8

"""
    Merlin

    API Guide for accessing Merlin's model management, deployment, and serving functionalities

    The version of the OpenAPI document: 0.14.0
    Generated by OpenAPI Generator (https://openapi-generator.tech)

    Do not edit the class manually.
"""  # noqa: E501


from __future__ import annotations
import pprint
import re  # noqa: F401
import json

from datetime import datetime
from typing import Any, ClassVar, Dict, List, Optional
from pydantic import BaseModel, StrictBool, StrictInt, StrictStr
from client.models.autoscaling_policy import AutoscalingPolicy
from client.models.deployment_mode import DeploymentMode
from client.models.endpoint_status import EndpointStatus
from client.models.env_var import EnvVar
from client.models.environment import Environment
from client.models.logger import Logger
from client.models.model_observability import ModelObservability
from client.models.mounted_mlp_secret import MountedMLPSecret
from client.models.protocol import Protocol
from client.models.resource_request import ResourceRequest
from client.models.transformer import Transformer
try:
    from typing import Self
except ImportError:
    from typing_extensions import Self

class VersionEndpoint(BaseModel):
    """
    VersionEndpoint
    """ # noqa: E501
    id: Optional[StrictStr] = None
    version_id: Optional[StrictInt] = None
    status: Optional[EndpointStatus] = None
    url: Optional[StrictStr] = None
    service_name: Optional[StrictStr] = None
    environment_name: Optional[StrictStr] = None
    environment: Optional[Environment] = None
    monitoring_url: Optional[StrictStr] = None
    message: Optional[StrictStr] = None
    resource_request: Optional[ResourceRequest] = None
    image_builder_resource_request: Optional[ResourceRequest] = None
    env_vars: Optional[List[EnvVar]] = None
    secrets: Optional[List[MountedMLPSecret]] = None
    transformer: Optional[Transformer] = None
    logger: Optional[Logger] = None
    deployment_mode: Optional[DeploymentMode] = None
    autoscaling_policy: Optional[AutoscalingPolicy] = None
    protocol: Optional[Protocol] = None
    enable_model_observability: Optional[StrictBool] = None
    model_observability: Optional[ModelObservability] = None
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None
    __properties: ClassVar[List[str]] = ["id", "version_id", "status", "url", "service_name", "environment_name", "environment", "monitoring_url", "message", "resource_request", "image_builder_resource_request", "env_vars", "secrets", "transformer", "logger", "deployment_mode", "autoscaling_policy", "protocol", "enable_model_observability", "model_observability", "created_at", "updated_at"]

    model_config = {
        "populate_by_name": True,
        "validate_assignment": True
    }


    def to_str(self) -> str:
        """Returns the string representation of the model using alias"""
        return pprint.pformat(self.model_dump(by_alias=True))

    def to_json(self) -> str:
        """Returns the JSON representation of the model using alias"""
        # TODO: pydantic v2: use .model_dump_json(by_alias=True, exclude_unset=True) instead
        return json.dumps(self.to_dict())

    @classmethod
    def from_json(cls, json_str: str) -> Self:
        """Create an instance of VersionEndpoint from a JSON string"""
        return cls.from_dict(json.loads(json_str))

    def to_dict(self) -> Dict[str, Any]:
        """Return the dictionary representation of the model using alias.

        This has the following differences from calling pydantic's
        `self.model_dump(by_alias=True)`:

        * `None` is only added to the output dict for nullable fields that
          were set at model initialization. Other fields with value `None`
          are ignored.
        """
        _dict = self.model_dump(
            by_alias=True,
            exclude={
            },
            exclude_none=True,
        )
        # override the default output from pydantic by calling `to_dict()` of environment
        if self.environment:
            _dict['environment'] = self.environment.to_dict()
        # override the default output from pydantic by calling `to_dict()` of resource_request
        if self.resource_request:
            _dict['resource_request'] = self.resource_request.to_dict()
        # override the default output from pydantic by calling `to_dict()` of image_builder_resource_request
        if self.image_builder_resource_request:
            _dict['image_builder_resource_request'] = self.image_builder_resource_request.to_dict()
        # override the default output from pydantic by calling `to_dict()` of each item in env_vars (list)
        _items = []
        if self.env_vars:
            for _item in self.env_vars:
                if _item:
                    _items.append(_item.to_dict())
            _dict['env_vars'] = _items
        # override the default output from pydantic by calling `to_dict()` of each item in secrets (list)
        _items = []
        if self.secrets:
            for _item in self.secrets:
                if _item:
                    _items.append(_item.to_dict())
            _dict['secrets'] = _items
        # override the default output from pydantic by calling `to_dict()` of transformer
        if self.transformer:
            _dict['transformer'] = self.transformer.to_dict()
        # override the default output from pydantic by calling `to_dict()` of logger
        if self.logger:
            _dict['logger'] = self.logger.to_dict()
        # override the default output from pydantic by calling `to_dict()` of autoscaling_policy
        if self.autoscaling_policy:
            _dict['autoscaling_policy'] = self.autoscaling_policy.to_dict()
        # override the default output from pydantic by calling `to_dict()` of model_observability
        if self.model_observability:
            _dict['model_observability'] = self.model_observability.to_dict()
        return _dict

    @classmethod
    def from_dict(cls, obj: Dict) -> Self:
        """Create an instance of VersionEndpoint from a dict"""
        if obj is None:
            return None

        if not isinstance(obj, dict):
            return cls.model_validate(obj)

        _obj = cls.model_validate({
            "id": obj.get("id"),
            "version_id": obj.get("version_id"),
            "status": obj.get("status"),
            "url": obj.get("url"),
            "service_name": obj.get("service_name"),
            "environment_name": obj.get("environment_name"),
            "environment": Environment.from_dict(obj.get("environment")) if obj.get("environment") is not None else None,
            "monitoring_url": obj.get("monitoring_url"),
            "message": obj.get("message"),
            "resource_request": ResourceRequest.from_dict(obj.get("resource_request")) if obj.get("resource_request") is not None else None,
            "image_builder_resource_request": ResourceRequest.from_dict(obj.get("image_builder_resource_request")) if obj.get("image_builder_resource_request") is not None else None,
            "env_vars": [EnvVar.from_dict(_item) for _item in obj.get("env_vars")] if obj.get("env_vars") is not None else None,
            "secrets": [MountedMLPSecret.from_dict(_item) for _item in obj.get("secrets")] if obj.get("secrets") is not None else None,
            "transformer": Transformer.from_dict(obj.get("transformer")) if obj.get("transformer") is not None else None,
            "logger": Logger.from_dict(obj.get("logger")) if obj.get("logger") is not None else None,
            "deployment_mode": obj.get("deployment_mode"),
            "autoscaling_policy": AutoscalingPolicy.from_dict(obj.get("autoscaling_policy")) if obj.get("autoscaling_policy") is not None else None,
            "protocol": obj.get("protocol"),
            "enable_model_observability": obj.get("enable_model_observability"),
            "model_observability": ModelObservability.from_dict(obj.get("model_observability")) if obj.get("model_observability") is not None else None,
            "created_at": obj.get("created_at"),
            "updated_at": obj.get("updated_at")
        })
        return _obj


