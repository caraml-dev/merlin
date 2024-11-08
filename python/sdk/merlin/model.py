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

import os
import pathlib
import shutil
import tempfile
import warnings
from datetime import datetime
from enum import Enum
from sys import version_info
from time import sleep
from typing import Any, Dict, List, Optional, Tuple, Union

import client
import docker
import pyprind
from client import (
    EndpointApi,
    EnvironmentApi,
    ModelEndpointsApi,
    ModelsApi,
    SecretApi,
    VersionApi,
    VersionImageApi,
)
from docker import APIClient
from docker.models.containers import Container
from merlin.autoscaling import (
    RAW_DEPLOYMENT_DEFAULT_AUTOSCALING_POLICY,
    SERVERLESS_DEFAULT_AUTOSCALING_POLICY,
    AutoscalingPolicy,
)
from merlin.batch.config import PredictionJobConfig
from merlin.batch.job import PredictionJob
from merlin.batch.sink import BigQuerySink
from merlin.batch.source import BigQuerySource
from merlin.deployment_mode import DeploymentMode
from merlin.docker.docker import copy_standard_dockerfile, wait_build_complete
from merlin.endpoint import ModelEndpoint, Status, VersionEndpoint
from merlin.logger import Logger
from merlin.model_schema import ModelSchema
from merlin.protocol import Protocol
from merlin.pyfunc import run_pyfunc_local_server
from merlin.requirements import process_conda_env
from merlin.resource_request import ResourceRequest
from merlin.transformer import Transformer
from merlin.util import (
    autostr,
    download_files_from_blob_storage,
    extract_optional_value_with_default,
    guess_mlp_ui_url,
    valid_name_check,
)
from merlin.validation import validate_model_dir
from merlin.version_image import VersionImage
from mlflow.entities import Run, RunData
from mlflow.exceptions import MlflowException
from mlflow.pyfunc import PythonModel
from mlflow.tracking.client import MlflowClient

import mlflow
from merlin import pyfunc

# Ensure backward compatibility after moving PyFuncModel and PyFuncV2Model to pyfunc.py
# This allows users to do following import statement
#
# from merlin.model import PyFuncModel, PyFuncV2Model
#
PyFuncModel = pyfunc.PyFuncModel
PyFuncV2Model = pyfunc.PyFuncV2Model
PyFuncV3Model = pyfunc.PyFuncV3Model
PYFUNC_EXTRA_ARGS_KEY = pyfunc.PYFUNC_EXTRA_ARGS_KEY
PYFUNC_MODEL_INPUT_KEY = pyfunc.PYFUNC_MODEL_INPUT_KEY

DEFAULT_MODEL_PATH = "model"
DEFAULT_MODEL_VERSION_LIMIT = 50
DEFAULT_API_CALL_RETRY = 5
DEFAULT_PREDICTION_JOB_DELAY = 5
DEFAULT_PREDICTION_JOB_RETRY_DELAY = 30
V1 = "v1"
PREDICTION_JOB = "PredictionJob"


class ModelEndpointDeploymentError(Exception):
    def __init__(self, model_name: str, version: int, details: str):
        self._model_name = model_name
        self._version = version
        self._details = details

    @property
    def model_name(self):
        return self._model_name

    @property
    def version(self):
        return self._version

    @property
    def details(self):
        return self._details


@autostr
class Project:
    def __init__(
        self, project: client.Project, mlp_url: str, api_client: client.ApiClient
    ):
        self._id = project.id
        self._name = project.name
        self._mlflow_tracking_url = project.mlflow_tracking_url
        self._created_at = project.created_at
        self._updated_at = project.updated_at
        self._url = mlp_url
        self._api_client = api_client
        self._readers = extract_optional_value_with_default(project.readers, [])
        self._administrators = extract_optional_value_with_default(
            project.administrators, []
        )

    @property
    def id(self) -> int:
        return int(self._id)

    @property
    def name(self) -> str:
        return self._name

    @property
    def mlflow_tracking_url(self) -> Optional[str]:
        return self._mlflow_tracking_url

    @property
    def readers(self) -> List[str]:
        return self._readers

    @property
    def administrators(self) -> List[str]:
        return self._administrators

    @property
    def created_at(self) -> Optional[datetime]:
        return self._created_at

    @property
    def updated_at(self) -> Optional[datetime]:
        return self._updated_at

    @property
    def url(self) -> str:
        return self._url

    def list_model(self) -> List["Model"]:
        """
        List all model available within the project
        :return: list of Model
        """
        model_api = ModelsApi(self._api_client)
        m_list = model_api.projects_project_id_models_get(project_id=int(self.id))

        result = []
        for model in m_list:
            result.append(Model(model, self, self._api_client))
        return result

    def get_or_create_model(
        self, model_name: str, model_type: "ModelType" = None
    ) -> "Model":
        """
        Get or create a model with given name

        :param model_name: model name
        :param model_type: type of model, mandatory when creation is needed
        :return: Model instance
        """
        if not valid_name_check(model_name):
            raise ValueError(
                """Your project/model name contains invalid characters.\
                    \nUse only the following characters\
                    \n- Characters: a-z (Lowercase ONLY)\
                    \n- Numbers: 0-9\
                    \n- Symbols: -
                """
            )

        model_api = ModelsApi(self._api_client)
        m_list = model_api.projects_project_id_models_get(
            project_id=int(self.id), name=model_name
        )

        if len(m_list) == 0:
            if model_type is None:
                raise ValueError(
                    f"model {model_name} is not found, specify "
                    f"{model_type} to create it"
                )
            model = model_api.projects_project_id_models_post(
                project_id=int(self.id),
                body=client.Model(name=model_name, type=model_type.value),
            )
        else:
            model = m_list[0]

        return Model(model, self, self._api_client)

    def create_secret(self, name: str, data: str):
        """
        Create a secret within the project

        :param name: secret name
        :param data: secret data
        :return:
        """
        secret_api = SecretApi(self._api_client)
        secret_api.projects_project_id_secrets_post(
            project_id=int(self.id), body=client.Secret(name=name, data=data)
        )

    def list_secret(self) -> List[str]:
        """
        List all secret name within the project

        :return:
        """
        secret_api = SecretApi(self._api_client)
        secrets = secret_api.projects_project_id_secrets_get(project_id=int(self.id))
        secret_names = []
        for s in secrets:
            if s.name is not None:
                secret_names.append(s.name)
        return secret_names

    def update_secret(self, name: str, data: str):
        """
        Update secret with given name

        :param name: secret name
        :param data: new secret data
        :return:
        """
        secret_api = SecretApi(self._api_client)
        match = self._find_secret(name)

        secret_api.projects_project_id_secrets_secret_id_patch(
            project_id=int(self.id),
            secret_id=int(match.id),
            body=client.Secret(name=name, data=data),
        )

    def delete_secret(self, name: str):
        """
        Delete secret with given name

        :param name: secret to be removed
        :return:
        """
        secret_api = SecretApi(self._api_client)
        match = self._find_secret(name)

        secret_api.projects_project_id_secrets_secret_id_delete(
            project_id=int(self.id), secret_id=int(match.id)
        )

    def _find_secret(self, name: str):
        secret_api = SecretApi(self._api_client)
        secrets = secret_api.projects_project_id_secrets_get(project_id=int(self.id))
        match = None
        for s in secrets:
            if s.name == name:
                match = s
        if match is None:
            raise ValueError(f"unable to find secret {name} in project {self.name}")
        return match


class ModelType(Enum):
    """
    Model type supported by merlin
    """

    XGBOOST = "xgboost"
    TENSORFLOW = "tensorflow"
    SKLEARN = "sklearn"
    PYTORCH = "pytorch"
    ONNX = "onnx"
    PYFUNC = "pyfunc"
    PYFUNC_V2 = "pyfunc_v2"
    CUSTOM = "custom"


@autostr
class Model:
    """
    Model representation
    """

    def __init__(
        self, model: client.Model, project: Project, api_client: client.ApiClient
    ):
        self._id = extract_optional_value_with_default(model.id, 0)
        self._name = model.name
        self._mlflow_experiment_id = model.mlflow_experiment_id
        self._type = ModelType(model.type)
        self._mlflow_url = model.mlflow_url
        self._created_at = model.created_at
        self._updated_at = model.updated_at
        self._project = project
        self._api_client = api_client

    @property
    def id(self) -> int:
        return int(self._id)

    @property
    def name(self) -> str:
        return self._name

    @property
    def type(self) -> ModelType:
        return self._type

    @property
    def mlflow_url(self) -> Optional[str]:
        return self._mlflow_url

    @property
    def mlflow_experiment_id(self) -> Optional[int]:
        if self._mlflow_experiment_id is not None:
            return int(self._mlflow_experiment_id)
        return None

    @property
    def created_at(self) -> Optional[datetime]:
        return self._created_at

    @property
    def updated_at(self) -> Optional[datetime]:
        return self._updated_at

    @property
    def project(self) -> Project:
        return self._project

    @property
    def endpoint(self) -> Optional[ModelEndpoint]:
        """
        Get endpoint of this model that is deployed in default environment

        :return: Endpoint if exist, otherwise None
        """
        mdl_endpoints_api = ModelEndpointsApi(self._api_client)
        mdl_endpoints_list = mdl_endpoints_api.models_model_id_endpoints_get(
            model_id=self.id
        )
        for endpoint in mdl_endpoints_list:
            if endpoint.environment is not None and endpoint.environment.is_default:
                return ModelEndpoint(endpoint)

        return None

    def list_endpoint(self) -> List[ModelEndpoint]:
        """
        List all model endpoint assosiated with this model

        :return: List[ModelEndpoint]
        """
        mdl_endpoints_api = ModelEndpointsApi(self._api_client)
        mdl_endpoints_list = mdl_endpoints_api.models_model_id_endpoints_get(
            model_id=self.id
        )
        mdl_endpoints = []
        for mdl_ep in mdl_endpoints_list:
            mdl_endpoints.append(ModelEndpoint(mdl_ep))
        return mdl_endpoints

    def get_version(self, id: int) -> Optional["ModelVersion"]:
        """
        Get version with specific ID

        :param id: version id to retrieve
        :return:
        """
        version_api = VersionApi(self._api_client)
        v_list = version_api.models_model_id_versions_get(int(self.id))
        for v in v_list:
            if v.id == id:
                return ModelVersion(v, self, self._api_client)
        return None

    def list_version(self, labels: Dict[str, List[str]] = None) -> List["ModelVersion"]:
        """
        List all version of the model

        :return: list of ModelVersion
        """
        result: List["ModelVersion"] = []
        search_dsl = self._build_search_labels_dsl(labels)
        versions, cursor = self._list_version_pagination(search=search_dsl)
        result = result + versions
        while cursor != "":
            versions, cursor = self._list_version_pagination(
                cursor=cursor, search=search_dsl
            )
            result = result + versions
        return result

    def _build_search_labels_dsl(self, labels: Dict[str, List[str]] = None):
        if labels is None:
            return ""

        all_search_kv_pair = []
        for key, list_value in labels.items():
            search_kv_pair = f"{key} in ({','.join(list_value)})"
            all_search_kv_pair.append(search_kv_pair)

        return f"labels:{','.join(all_search_kv_pair)}"

    def _get_next_cursor_from_headers(self, headers: Dict[str, str]) -> str:
        next_cursor_key = "next-cursor"
        for key in headers:
            if key.lower() == next_cursor_key:
                return headers[key]

        return ""

    def _list_version_pagination(
        self, limit=DEFAULT_MODEL_VERSION_LIMIT, cursor="", search=""
    ) -> Tuple[List["ModelVersion"], str]:
        """
        List version of the model with pagination
        :param limit: integer, max number of rows will be returned
        :param cursor: text, indicator where backend will start to look up the data
        :param search: text

        :return: list of ModelVersion
        :return: next cursor to fetch next page of version
        """
        version_api = VersionApi(self._api_client)
        version_api_response = version_api.models_model_id_versions_get_with_http_info(
            int(self.id), limit=limit, cursor=cursor, search=search
        )
        versions = version_api_response.data
        headers = extract_optional_value_with_default(version_api_response.headers, {})

        next_cursor = self._get_next_cursor_from_headers(headers)
        result = []
        for v in versions:
            result.append(ModelVersion(v, self, self._api_client))
        return result, next_cursor

    def new_model_version(
        self, labels: Dict[str, str] = None, model_schema: Optional[ModelSchema] = None
    ) -> "ModelVersion":
        """
        Create a new version of this model

        :param labels:
        :return:  new ModelVersion
        """
        version_api = VersionApi(self._api_client)
        python_version = f"{version_info.major}.{version_info.minor}.*"  # capture user's python version
        model_schema_payload = None
        if model_schema is not None:
            model_schema.model_id = self.id
            model_schema_payload = model_schema.to_client_model_schema()
        v = version_api.models_model_id_versions_post(
            int(self.id),
            body=client.Version(
                labels=labels,
                python_version=python_version,
                model_schema=model_schema_payload,
            ),
        )
        return ModelVersion(v, self, self._api_client)

    def serve_traffic(
        self, traffic_rule: Dict["VersionEndpoint", int], environment_name: str = None
    ) -> ModelEndpoint:
        """
        Set traffic rule for this model.

        :param traffic_rule: dict of version endpoint and the percentage of traffic.
        :param environment_name: target environment in which the model endpoint will be created. If left empty it will create in default environment.
        :return: ModelEndpoint
        """
        if not isinstance(traffic_rule, dict):
            raise ValueError(
                f"Traffic_rule should be dictionary, got: {type(traffic_rule)}"
            )

        if len(traffic_rule) > 1:
            raise ValueError("Traffic splitting is not yet supported")

        target_env = environment_name
        if target_env is None:
            env_api = EnvironmentApi(self._api_client)
            env_list = env_api.environments_get()
            target_env = _get_default_target_env_name(env_list)

        total_traffic = 0
        for version_endpoint, traffic_split in traffic_rule.items():
            if version_endpoint.environment_name != target_env:
                raise ValueError(
                    "Version Endpoint must have same " "environment as target"
                )

            if traffic_split <= 0:
                raise ValueError("Traffic percentage should be non negative")
            total_traffic += traffic_split

        if total_traffic != 100:
            raise ValueError("Total traffic percentage should be 100")

        version_endpoint = traffic_rule.popitem()[0]

        # get existing model endpoint
        mdl_epi_api = ModelEndpointsApi(self._api_client)
        endpoints = mdl_epi_api.models_model_id_endpoints_get(model_id=self.id)
        prev_endpoint = None
        for endpoint in endpoints:
            if endpoint.environment_name == target_env:
                prev_endpoint = endpoint

        if prev_endpoint is None:
            # create
            dst = client.ModelEndpointRuleDestination(
                version_endpoint_id=version_endpoint.id, weight=100
            )
            rule = client.ModelEndpointRule(destinations=[dst])
            ep = client.ModelEndpoint(
                model_id=self.id, environment_name=target_env, rule=rule
            )
            ep = mdl_epi_api.models_model_id_endpoints_post(model_id=self.id, body=ep)
        elif prev_endpoint.id is not None:
            # update: GET and PUT
            ep = mdl_epi_api.models_model_id_endpoints_model_endpoint_id_get(
                model_id=self.id, model_endpoint_id=prev_endpoint.id
            )
            if (
                ep.rule is not None
                and ep.rule.destinations is not None
                and len(ep.rule.destinations) > 0
            ):
                ep.rule.destinations[0].version_endpoint_id = version_endpoint.id
                ep.rule.destinations[0].weight = 100

            ep = mdl_epi_api.models_model_id_endpoints_model_endpoint_id_put(
                model_id=int(self.id),
                model_endpoint_id=prev_endpoint.id,
                body=ep,
            )

        return ModelEndpoint(ep)

    def stop_serving_traffic(self, environment_name: str = None):
        """
        Stop serving traffic for this model in given environment.

        :param environment_name: environment name in which the endpoint should be stopped from serving traffic. If environment_name is empty it will attempt to undeploy the model from default environment.
        """
        target_env = environment_name
        if target_env is None:
            env_api = EnvironmentApi(self._api_client)
            env_list = env_api.environments_get()
            for env in env_list:
                if env.is_default:
                    target_env = env.name

            if target_env is None:
                raise ValueError(
                    "Unable to find default environment, "
                    "pass environment_name to the method"
                )

        mdl_epi_api = ModelEndpointsApi(self._api_client)
        endpoints = mdl_epi_api.models_model_id_endpoints_get(model_id=self.id)

        target_endpoint = None
        for endpoint in endpoints:
            if endpoint.environment_name == target_env:
                target_endpoint = endpoint

        if target_endpoint is None:
            raise ValueError(
                f"there is no model endpoint for model "
                f"{self.name} in {target_env} environment"
            )

        print(
            f"Stopping serving traffic for model {self.name} "
            f"in {target_env} environment"
        )
        if target_endpoint.id is None:
            raise ValueError(f"model endpoint doesn't have id information")
        mdl_epi_api.models_model_id_endpoints_model_endpoint_id_delete(
            self.id, target_endpoint.id
        )

    def set_traffic(self, traffic_rule: Dict["ModelVersion", int]) -> ModelEndpoint:
        """
        Set traffic rule for this model.

        *This method is deprecated, use serve_traffic instead*

        :param traffic_rule: dict of model version and the percentage of traffic.
        :return: ModelEndpoint
        """
        print(
            "This method is going to be deprecated, please use " "serve_traffic instead"
        )
        if not isinstance(traffic_rule, dict):
            raise ValueError(
                f"Traffic_rule should be dictionary, got: {type(traffic_rule)}"
            )

        if len(traffic_rule) > 1:
            raise ValueError("Traffic splitting is not yet supported")

        total_traffic = 0
        for mdl_version, traffic_split in traffic_rule.items():
            if traffic_split <= 0:
                raise ValueError("Traffic percentage should be non negative")

            total_traffic += traffic_split
            if (
                mdl_version.endpoint is None
                or mdl_version.endpoint.status != Status.RUNNING
            ):
                raise ValueError(
                    f"Model version with id {mdl_version.id} is not running"
                )

        if total_traffic != 100:
            raise ValueError(f"Total traffic should be 100, got {total_traffic}")

        mdl_version = traffic_rule.popitem()[0]
        model_endpoint_api = ModelEndpointsApi(self._api_client)
        if mdl_version.endpoint is None:
            raise ValueError(
                f"there is no version endpoint for model version "
                f"{mdl_version.id} in default environment"
            )
        def_version_endpoint = mdl_version.endpoint

        if self.endpoint is None:
            # create model endpoint
            ep = model_endpoint_api.models_model_id_endpoints_post(
                body=client.ModelEndpoint(
                    model_id=self.id,
                    rule=client.ModelEndpointRule(
                        destinations=[
                            client.ModelEndpointRuleDestination(
                                version_endpoint_id=def_version_endpoint.id,
                                weight=100,
                            )
                        ]
                    ),
                ),
                model_id=int(self.id),
            )
            return ModelEndpoint(ep)
        else:
            def_model_endpoint = self.endpoint
            # GET and PUT
            ep = model_endpoint_api.models_model_id_endpoints_model_endpoint_id_get(
                model_id=int(self.id), model_endpoint_id=def_model_endpoint.id
            )
            if (
                ep.rule is not None
                and ep.rule.destinations is not None
                and len(ep.rule.destinations) > 0
            ):
                ep.rule.destinations[0].version_endpoint_id = def_version_endpoint.id
                ep.rule.destinations[0].weight = 100

            ep = model_endpoint_api.models_model_id_endpoints_model_endpoint_id_put(
                model_id=int(self.id),
                model_endpoint_id=def_model_endpoint.id,
                body=ep,
            )

        return ModelEndpoint(ep)

    def delete_model(self) -> int:
        """
        Delete this model. Please note that any inactive related entity (model versions, endpoints and prediction jobs) will get deleted by this process.
        Deletion will fail if there are any active related entity (model versions, endpoints and prediction jobs)

        :return: id of deleted model
        """
        model_api = ModelsApi(self._api_client)
        return model_api.projects_project_id_models_model_id_delete(
            int(self.project.id), int(self.id)
        )


@autostr
class ModelVersion:
    """
    Representation of version in a model
    """

    MODEL_TYPE_TO_IMAGE_MAP = {
        ModelType.SKLEARN: "gcr.io/kfserving/sklearnserver:0.2.2",
        ModelType.TENSORFLOW: "tensorflow/serving:1.14.0",
        ModelType.XGBOOST: "gcr.io/kfserving/xgbserver:0.2.2",
        ModelType.PYTORCH: "gcr.io/kfserving/pytorchserver:0.2.2",
    }

    def __init__(
        self, version: client.Version, model: Model, api_client: client.ApiClient
    ):
        self._api_client = api_client
        self._id = extract_optional_value_with_default(version.id, 0)
        self._mlflow_run_id = version.mlflow_run_id
        self._mlflow_url = version.mlflow_url
        self._created_at = version.created_at
        self._updated_at = version.updated_at
        self._properties = version.properties
        self._model = model
        self._artifact_uri = version.artifact_uri
        self._labels = version.labels
        self._custom_predictor = version.custom_predictor
        self._python_version = version.python_version
        self._model_schema = ModelSchema.from_model_schema_response(
            version.model_schema
        )
        mlflow.set_tracking_uri(model.project.mlflow_tracking_url)  # type: ignore  # noqa

        endpoints = None
        if version.endpoints is not None:
            endpoints = []
            for ep in version.endpoints:
                endpoints.append(VersionEndpoint(ep))
        self._version_endpoints = endpoints

    @property
    def id(self) -> int:
        return self._id

    @property
    def mlflow_run_id(self) -> Optional[str]:
        return self._mlflow_run_id

    @property
    def mlflow_url(self) -> Optional[str]:
        return self._mlflow_url

    @property
    def version_endpoints(self) -> List[VersionEndpoint]:
        # For backward compatibility, we call VersionEndpoint API if _version_endpoints empty.
        if self._version_endpoints is None:
            self._version_endpoints = self.list_endpoint()
        return self._version_endpoints

    @property
    def endpoint(self) -> Optional[VersionEndpoint]:
        """
        Return endpoint of this model version that is deployed in default
        environment

        :return: VersionEndpoint or None
        """
        for endpoint in self.version_endpoints:
            if endpoint.environment is not None and endpoint.environment.is_default:
                return endpoint
        return None

    @property
    def properties(self) -> object:
        return self._properties

    @property
    def created_at(self) -> Optional[datetime]:
        return self._created_at

    @property
    def updated_at(self) -> Optional[datetime]:
        return self._updated_at

    @property
    def model(self) -> Model:
        return self._model

    @property
    def artifact_uri(self):
        return self._artifact_uri

    @property
    def labels(self):
        return self._labels

    @property
    def url(self) -> str:
        project_id = self.model.project.id
        model_id = self.model.id
        base_url = guess_mlp_ui_url(self.model.project.url)
        return f"{base_url}/projects/{project_id}/models/{model_id}/versions"

    @property
    def model_schema(self) -> Optional[ModelSchema]:
        return self._model_schema

    def start(self):
        """
        Start a model version

        :return:
        """
        # set the experiment in case there was other experiment set prior
        # this assume the mlflow run and experiment were created during init
        client = MlflowClient()
        r = client.get_run(self._mlflow_run_id)
        mlflow.set_experiment(experiment_id=r.info.experiment_id)
        mlflow.start_run(self._mlflow_run_id)

    def finish(self):
        """
        Finalize a model version

        :return:
        """
        mlflow.end_run()

    def log_param(self, key, value):
        """
        Log param to a model version

        :param key:
        :param value:
        :return:
        """
        mlflow.log_param(key, value)

    def log_metric(self, key, value):
        """
        Log a metric to model version

        :param key:
        :param value:
        :return:
        """
        mlflow.log_metric(key, value)

    def set_tag(self, key, value):
        """
        Set tag in a model version

        :param key:
        :param value:
        :return:
        """
        mlflow.set_tag(key, value)

    def delete_tag(self, key):
        """
        Delete tag in a model version

        :param key:
        :return:
        """
        mlflow.delete_tag(key)

    def get_run(self) -> Optional[Run]:
        """
        Get MLFlow Run in a model version
        """
        try:
            return mlflow.get_run(self._mlflow_run_id)  # type: ignore  # noqa
        except MlflowException:
            return None

    def get_run_data(self) -> Optional[RunData]:
        mlflow_run = self.get_run()
        if mlflow_run is None:
            return None

        run_data = mlflow_run.data
        return run_data

    def get_param(self, key) -> Optional[str]:
        """
        Get param value for specific param name(key)

        :param key:
        :return value:
        """
        run_data = self.get_run_data()
        if run_data is None:
            return None

        return run_data.params.get(key)

    def get_metric(self, key) -> Optional[float]:
        """
        Get metric value from metric name(key)

        :param key:
        :return value:
        """
        run_data = self.get_run_data()
        if run_data is None:
            return None

        return run_data.metrics.get(key)

    def get_tag(self, key) -> Optional[str]:
        """
        Get tag value from name(key)

        :param key:
        :return value:
        """
        run_data = self.get_run_data()
        if run_data is None:
            return None

        return run_data.tags.get(key)

    def list_tag(self) -> Dict[str, str]:
        """
        Get all tags

        :return list of tags:
        """
        run_data = self.get_run_data()
        if run_data is None:
            return {}

        return run_data.tags

    def download_artifact(self, destination_path):
        """
        Download artifact

        :param destination_path:
        :return:
        """
        run = self.get_run()
        if run is None:
            raise Exception("There is no mlflow run for this model version")

        run_info = run.info
        artifact_uri = run_info.artifact_uri
        if artifact_uri is None or artifact_uri == "":
            raise Exception("There is no artifact uri for this model version")

        download_files_from_blob_storage(artifact_uri, destination_path)

    def log_artifacts(self, local_dir, artifact_path=None):
        """
        Log artifacts

        :param local_dir:
        :param artifact_path:
        :return:
        """
        mlflow.log_artifacts(local_dir, artifact_path)

    def log_artifact(self, local_path, artifact_path=None):
        """
        Log artifact

        :param local_path:
        :param artifact_path:
        :return:
        """
        mlflow.log_artifact(local_path, artifact_path)

    def log_pyfunc_model(
        self,
        model_instance: PythonModel,
        conda_env: Union[str, Dict[str, Any]],
        code_dir: Optional[List[str]] = None,
        artifacts: Dict[str, str] = None,
    ):
        """
        Upload PyFunc based model into artifact storage.
        User has to specify model_instance and
        conda_env. `model_instance` shall implement all method specified in
        PyFuncModel. conda_env shall contain all dependency required by the
        model

        :param model_instance: instance of python function model
        :param conda_env: path to conda env.yaml file or dictionary representation of it
        :param code_dir: additional code directory that will be loaded with ModelType.PYFUNC model
        :param artifacts: dictionary of artifact that will be stored together with the model. This will be passed to PythonModel.initialize. Example: {"config" : "config/staging.yaml"}
        """
        if (
            self._model.type != ModelType.PYFUNC
            and self._model.type != ModelType.PYFUNC_V2
        ):
            raise ValueError(
                "log_pyfunc_model is only for PyFunc, PyFuncV2 and PyFuncV3 model"
            )

        merlin_requirements = ["merlin-pyfunc-server<0.42.0"]
        if self._model.type == ModelType.PYFUNC_V2:
            merlin_requirements = ["merlin-batch-predictor<0.42.0"]

        # add/replace python version in conda to match that used to create model version
        conda_env = process_conda_env(
            conda_env, self._python_version, merlin_requirements
        )

        mlflow.pyfunc.log_model(
            DEFAULT_MODEL_PATH,
            python_model=model_instance,
            code_path=code_dir,
            conda_env=conda_env,
            artifacts=artifacts,
        )

    def log_pytorch_model(self, model_dir, model_class_name=None):
        """
        Upload PyTorch model to artifact storage.

        :param model_dir: directory containing serialized PyTorch model
        :param model_class_name: class name of PyTorch model. By default the model class name is 'PyTorchModel'
        """
        if self._model.type != ModelType.PYTORCH:
            raise ValueError("log_pytorch_model is only for PyTorch model")

        warnings.warn(
            "'log_pytorch_model' is deprecated, use 'log_model' instead",
            DeprecationWarning,
        )
        self.log_model(model_dir)

    def log_model(self, model_dir=None):
        """
        Upload model to artifact storage.
        This method is used to upload model for xgboost, tensorflow,
        and sklearn model.

        :param model_dir: directory which contain serialized model
        """
        if (
            self._model.type == ModelType.PYFUNC
            or self._model.type == ModelType.PYFUNC_V2
        ):
            raise ValueError("use log_pyfunc_model to log pyfunc model")

        validate_model_dir(self._model.type, model_dir)
        mlflow.log_artifacts(model_dir, DEFAULT_MODEL_PATH)

    def log_custom_model(
        self, image: str, model_dir: str = None, command: str = "", args: str = ""
    ):
        """
        Upload model to artifact storage.
        This method is used to upload model for custom model type.

        :param image: Docker image that will be used as predictor
        :param model_dir: directory which contain serialized model
        :param command: Command to run docker image
        :param args: Arguments that needs to be specified when running docker
        """
        if self._model.type != ModelType.CUSTOM:
            raise ValueError("use log_custom_model to log custom model")

        is_using_temp_dir = False
        model_properties_file = "model.properties"

        if model_dir is None:
            """
            Create temp directory, which later on will be uploaded
            The reason is iff no data that will be uploaded to mlflow artifact (gcs), given artifact URI will not exist
            Hence will raise error when creating inferenceservice
            """
            is_using_temp_dir = True
            model_dir = tempfile.mkdtemp(suffix="merlin-custom-model")

        with open(os.path.join(model_dir, model_properties_file), "w") as writer:
            writer.write(f"image = {image}\n")
            writer.write(f"command = {command}\n")
            writer.write(f"args = {args}\n")

        validate_model_dir(self._model.type, model_dir)
        mlflow.log_artifacts(model_dir, DEFAULT_MODEL_PATH)  # type: ignore  # noqa

        if is_using_temp_dir:
            """
            If user didn't specify model_dir, sdk will create new temp directory.
            This directory needs to be deleted after it is been uploaded to mlflow
            """
            shutil.rmtree(model_dir)

        version_api = VersionApi(self._api_client)
        custom_predictor_body = client.CustomPredictor(
            image=image, command=command, args=args
        )
        version_api.models_model_id_versions_version_id_patch(
            int(self.model.id),
            int(self.id),
            body=client.Version(custom_predictor=custom_predictor_body),
        )

    def list_endpoint(self) -> List[VersionEndpoint]:
        """
        Return all endpoint deployment for this particular model version

        :return: List of VersionEndpoint
        """
        endpoint_api = EndpointApi(self._api_client)
        ep_list = endpoint_api.models_model_id_versions_version_id_endpoint_get(
            model_id=self.model.id, version_id=self.id
        )

        endpoints = []
        for ep in ep_list:
            endpoints.append(VersionEndpoint(ep))
        return endpoints

    def build_image(
        self,
        backoff_limit: int = 0,
        resource_request: Optional[ResourceRequest] = None,
    ):
        """
        Build the Docker image for this model version.

        :param backoff_limit: The maximum number of retries before considering a job as failed.
        :param resource_request: The resource requirement (CPU & memory) for image building job.

        :return: VersionImage object
        """

        target_resource_request = None
        if resource_request is not None:
            target_resource_request = client.ResourceRequest(
                cpu_request=resource_request.cpu_request,
                memory_request=resource_request.memory_request,
            )

        options = client.BuildImageOptions(
            backoff_limit=backoff_limit,
            resource_request=target_resource_request,
        )

        version_image_api = VersionImageApi(self._api_client)
        version_image_api.models_model_id_versions_version_id_image_put(
            model_id=self.model.id,
            version_id=self.id,
            body=options,
        )

        bar = pyprind.ProgBar(
            100,
            track_time=True,
            title=f"Building Docker image for model {self.model.name} version {self.id}",
        )
        bar.update()
        sleep(10)

        while bar.active:
            image = version_image_api.models_model_id_versions_version_id_image_get(
                model_id=self.model.id, version_id=self.id
            )

            if image.exists:
                break

            if (
                image.image_building_job_status is not None
                and image.image_building_job_status.state != "active"
            ):
                break

            bar.update()
            sleep(10)
        bar.stop()

        if image.exists:
            print(
                f"Succefully built Docker image for model {self.model.name} version {self.id}."
                f"\nDocker image ref: {image.image_ref}"
            )
        else:
            print(
                f"Failed to build Docker image for model {self.model.name} version {self.id}"
            )
            if (
                image.image_building_job_status is not None
                and image.image_building_job_status.message != ""
            ):
                print(f"{image.image_building_job_status.message}")

        return VersionImage(image)

    def deploy(
        self,
        environment_name: str = None,
        resource_request: Optional[ResourceRequest] = None,
        image_builder_resource_request: Optional[ResourceRequest] = None,
        env_vars: Dict[str, str] = None,
        transformer: Transformer = None,
        logger: Logger = None,
        deployment_mode: DeploymentMode = None,
        autoscaling_policy: AutoscalingPolicy = None,
        protocol: Protocol = None,
        enable_model_observability: bool = False,
    ) -> VersionEndpoint:
        """
        Deploy current model to MLP One of log_model, log_pytorch_model,
        and log_pyfunc_model has to be called beforehand

        :param environment_name: target environment to which the model version will be deployed to. If left empty it will deploy to default environment.
        :param resource_request: The resource requirement and replicas requests for model version endpoint.
        :param env_vars: List of environment variables to be passed to the model container.
        :param transformer: The service to be deployed alongside the model for pre/post-processing steps.
        :param logger: Response/Request logging configuration for model or transformer.
        :param deployment_mode: mode of deployment for the endpoint (default: None)
        :param autoscaling_policy: autoscaling policy to be used for the deployment (default: None)
        :param protocol: protocol to be used for deploying the model (default: None)
        :return: VersionEndpoint object
        """
        env_list = self._get_env_list()
        target_env_name = (
            _get_default_target_env_name(env_list)
            if environment_name is None
            else environment_name
        )

        current_endpoint = self._get_endpoint_in_environment(target_env_name)
        default_resource_request = ModelVersion._get_default_resource_request(
            target_env_name, env_list
        )

        target_deployment_mode = None
        target_protocol = None
        target_resource_request = None
        target_image_builder_resource_request = None
        target_autoscaling_policy = None
        target_env_vars: List[client.EnvVar] = []
        target_transformer = None
        target_logger = None

        # Get the currently deployed endpoint and if there's no deployed endpoint yet, use the default values for
        # non-nullable fields
        if current_endpoint is None:
            target_deployment_mode = DeploymentMode.SERVERLESS.value
            target_protocol = Protocol.HTTP_JSON.value
            target_resource_request = default_resource_request
            target_autoscaling_policy = ModelVersion._get_default_autoscaling_policy(
                deployment_mode.value
                if deployment_mode is not None
                else target_deployment_mode
            )
        else:
            target_deployment_mode = current_endpoint.deployment_mode.value
            target_protocol = current_endpoint.protocol.value

        if deployment_mode is not None:
            target_deployment_mode = deployment_mode.value

        if protocol is not None:
            target_protocol = protocol.value

        if resource_request is not None:
            resource_request.validate()
            target_resource_request = client.ResourceRequest(
                min_replica=resource_request.min_replica,
                max_replica=resource_request.max_replica,
                cpu_request=resource_request.cpu_request,
                cpu_limit=resource_request.cpu_limit,
                memory_request=resource_request.memory_request,
            )

            if target_resource_request.min_replica is None:
                target_resource_request.min_replica = (
                    default_resource_request.min_replica
                )

            if (
                target_resource_request.max_replica is None
                or target_resource_request.max_replica < 1
            ):
                target_resource_request.max_replica = (
                    default_resource_request.max_replica
                )

            if (
                resource_request.gpu_request is not None
                and resource_request.gpu_name is not None
            ):
                for env in env_list:
                    if env.gpus is None:
                        continue

                    for gpu in env.gpus:
                        if resource_request.gpu_name == gpu.name:
                            if (
                                gpu.values is not None
                                and resource_request.gpu_request not in gpu.values
                            ):
                                raise ValueError(
                                    f"Invalid GPU request count. Supported GPUs count for  {resource_request.gpu_name} is {gpu.values}"
                                )

                            if target_resource_request is not None:
                                target_resource_request.gpu_name = (
                                    resource_request.gpu_name
                                )
                                target_resource_request.gpu_request = (
                                    resource_request.gpu_request
                                )
                            break

        if image_builder_resource_request is not None:
            target_image_builder_resource_request = client.ResourceRequest(
                cpu_request=image_builder_resource_request.cpu_request,
                memory_request=image_builder_resource_request.memory_request,
            )

        if autoscaling_policy is not None:
            target_autoscaling_policy = client.AutoscalingPolicy(
                metrics_type=client.MetricsType(autoscaling_policy.metrics_type.value),
                target_value=autoscaling_policy.target_value,
            )

        if env_vars is not None:
            target_env_vars = ModelVersion._add_env_vars(target_env_vars, env_vars)

        if transformer is not None:
            target_transformer = ModelVersion._create_transformer_spec(
                transformer, target_env_name, env_list
            )
            if (
                current_endpoint is not None
                and current_endpoint.transformer is not None
            ):
                target_transformer.id = current_endpoint.transformer.id

        if logger is not None:
            target_logger = logger.to_logger_spec()

        model = self._model
        endpoint_api = EndpointApi(self._api_client)

        endpoint = client.VersionEndpoint(
            environment_name=target_env_name,
            resource_request=target_resource_request,
            image_builder_resource_request=target_image_builder_resource_request,
            env_vars=target_env_vars,
            transformer=target_transformer,
            logger=target_logger,
            deployment_mode=client.DeploymentMode(target_deployment_mode),
            autoscaling_policy=target_autoscaling_policy,
            protocol=client.Protocol(target_protocol),
            enable_model_observability=enable_model_observability,
        )
        if current_endpoint is not None:
            # This allows a serving deployment to be updated while it is serving
            if current_endpoint.status == Status.SERVING:
                endpoint.status = client.EndpointStatus.SERVING
            else:
                endpoint.status = client.EndpointStatus.RUNNING

            if current_endpoint.id is None:
                raise ValueError("current endpoint must have id")

            endpoint = endpoint_api.models_model_id_versions_version_id_endpoint_endpoint_id_put(
                int(model.id),
                int(self.id),
                current_endpoint.id,
                body=endpoint,
            )
        else:
            endpoint = endpoint_api.models_model_id_versions_version_id_endpoint_post(
                int(model.id), int(self.id), body=endpoint
            )
        bar = pyprind.ProgBar(
            100,
            track_time=True,
            title=f"Deploying model {model.name} version " f"{self.id}",
        )

        while True:
            # Emulate a do-while loop. Re-get the endpoint so that the API server would have
            # started acting after the deployment job has been submitted.
            if endpoint.id is None:
                raise ValueError("endpoint id must be set")

            endpoint = endpoint_api.models_model_id_versions_version_id_endpoint_endpoint_id_get(
                model_id=model.id, version_id=self.id, endpoint_id=endpoint.id
            )
            if endpoint.status != "pending":
                break
            sleep(5)
            bar.update()
        bar.stop()

        if (
            endpoint.status != "running"
            and endpoint.status != "serving"
            and endpoint.message is not None
        ):
            raise ModelEndpointDeploymentError(model.name, self.id, endpoint.message)

        log_url = f"{self.url}/{self.id}/endpoints/{endpoint.id}/logs"
        print(
            f"Model {model.name} version {self.id} is deployed."
            f"\nView model version logs: {log_url}"
        )

        self._version_endpoints = self.list_endpoint()

        return VersionEndpoint(endpoint, log_url)

    def undeploy(self, environment_name: str = None):
        """
        Delete deployment of the model version

        :param environment_name: environment name in which the endpoint should be undeployed from. If environment_name is empty it will attempt to undeploy the model from default environment
        """
        target_env = environment_name
        if target_env is None:
            env_list = self._get_env_list()
            for env in env_list:
                if env.is_default:
                    target_env = env.name

            if target_env is None:
                raise ValueError(
                    "Unable to find default environment, "
                    "pass environment_name to the method"
                )

        endpoint_api = EndpointApi(self._api_client)
        endpoints = endpoint_api.models_model_id_versions_version_id_endpoint_get(
            model_id=self.model.id, version_id=self.id
        )
        target_endpoint = None
        for endpoint in endpoints:
            if endpoint.environment_name == target_env:
                target_endpoint = endpoint

        if target_endpoint is None or target_endpoint.id is None:
            print(f"No endpoint found for environment: {target_endpoint}")
            return

        print(
            f"Deleting deployment of model {self._model.name} "
            f"version {self.id} from enviroment {target_env}"
        )
        endpoint_api = EndpointApi(self._api_client)
        endpoint_api.models_model_id_versions_version_id_endpoint_endpoint_id_delete(
            self.model.id, self.id, target_endpoint.id
        )

    def create_prediction_job(
        self, job_config: PredictionJobConfig, sync: bool = True
    ) -> PredictionJob:
        """
        Create and run prediction job with given config using this model version

        :param sync: boolean to set synchronicity of job. The default is set to True.
        :param job_config: prediction job config
        :return: prediction job
        """
        if self.model.type != ModelType.PYFUNC_V2:
            raise ValueError(
                f"model type is not supported for prediction job: {self.model.type}"
            )

        job_cfg = client.PredictionJobConfig(
            version=V1,
            kind=PREDICTION_JOB,
            model=client.PredictionJobConfigModel(
                type=self.model.type.value.upper(),
                uri=os.path.join(self.artifact_uri, DEFAULT_MODEL_PATH),
                result=client.PredictionJobConfigModelResult(
                    type=client.ResultType(job_config.result_type.value),
                    item_type=client.ResultType(job_config.item_type.value),
                ),
            ),
        )

        if isinstance(job_config.source, BigQuerySource):
            job_cfg.bigquery_source = job_config.source.to_client_bq_source()
        else:
            raise ValueError(f"source type is not supported {type(job_config.source)}")

        if isinstance(job_config.sink, BigQuerySink):
            job_cfg.bigquery_sink = job_config.sink.to_client_config()
        else:
            raise ValueError(f"sink type is not supported {type(job_config.sink)}")

        cfg = client.Config(
            job_config=job_cfg, service_account_name=job_config.service_account_name
        )

        if job_config.resource_request is not None:
            cfg.resource_request = job_config.resource_request.to_dict()

        if job_config.image_builder_resource_request is not None:
            cfg.image_builder_resource_request = client.ResourceRequest(
                min_replica=0,
                max_replica=0,
                cpu_request=job_config.image_builder_resource_request.cpu_request,
                memory_request=job_config.image_builder_resource_request.memory_request,
            )

        target_env_vars = []
        if job_config.env_vars is not None:
            if not isinstance(job_config.env_vars, dict):
                raise ValueError(
                    f"env_vars should be dictionary, got: {type(job_config.env_vars)}"
                )

            if len(job_config.env_vars) > 0:
                for name, value in job_config.env_vars.items():
                    target_env_vars.append(client.EnvVar(name=name, value=value))
                cfg.env_vars = target_env_vars

        req = client.PredictionJob(
            version_id=self.id, model_id=self.model.id, config=cfg
        )
        job_client = client.PredictionJobsApi(self._api_client)
        j = job_client.models_model_id_versions_version_id_jobs_post(
            model_id=self.model.id, version_id=self.id, body=req
        )

        bar = pyprind.ProgBar(
            100,
            track_time=True,
            title=f"Running prediction job {j.id} from model {self.model.name} version {self.id} "
            f"under project {self.model.project.name}",
        )
        retry = DEFAULT_API_CALL_RETRY
        while (
            j.status == "pending" or j.status == "running" or j.status == "terminating"
        ):
            job_id = j.id
            if job_id is None:
                raise ValueError("job id must be exist")

            if not sync:
                j = job_client.models_model_id_versions_version_id_jobs_job_id_get(
                    model_id=self.model.id, version_id=self.id, job_id=job_id
                )
                return PredictionJob(j, self._api_client)
            else:
                try:
                    j = job_client.models_model_id_versions_version_id_jobs_job_id_get(
                        model_id=self.model.id, version_id=self.id, job_id=job_id
                    )
                    retry = DEFAULT_API_CALL_RETRY
                except Exception:
                    retry -= 1
                    if retry == 0:
                        j.status = "failed"
                        break
                    sleep(DEFAULT_PREDICTION_JOB_RETRY_DELAY)
            bar.update()
            sleep(DEFAULT_PREDICTION_JOB_DELAY)
        bar.stop()

        if j.status == "failed" or j.status == "failed_submission":
            raise ValueError(f"Prediction job {j.id} failed: {j.error}")

        return PredictionJob(j, self._api_client)

    def list_prediction_job(self) -> List[PredictionJob]:
        """
        List all prediction job created from the model version

        :return: list of prediction jobs
        """
        job_client = client.PredictionJobsApi(self._api_client)

        res = job_client.models_model_id_versions_version_id_jobs_by_page_get(
            model_id=self.model.id, version_id=self.id, page=1
        )
        jobs = []
        for j in res.results:
            jobs.append(PredictionJob(j, self._api_client))

        # Paginated response. Parse the rest of the pages.
        total_pages = res.paging.pages
        page = 2
        while page <= total_pages:
            res = job_client.models_model_id_versions_version_id_jobs_by_page_get(
                model_id=self.model.id, version_id=self.id, page=page
            )
            for j in res.results:
                jobs.append(PredictionJob(j, self._api_client))
            page += 1

        return jobs

    def start_server(
        self,
        env_vars: Dict[str, str] = None,
        port: int = 8080,
        protocol: Protocol = Protocol.HTTP_JSON,
        pyfunc_base_image: str = None,
        kill_existing_server: bool = False,
        tmp_dir: Optional[str] = os.environ.get("MERLIN_TMP_DIR"),
        build_image: bool = False,
        debug: bool = False,
    ):
        """
        Start a local server running the model version

        :param env_vars: dictionary of environment variables to be passed to the server
        :param port: host port that will be used to expose model server
        :param protocol: protocol to be used by the deployed model (default: HTTP_JSON)
        :param pyfunc_base_image: (optional, default=None) docker image to be used as base image for building pyfunc model
        :param kill_existing_server: (optional, default=False) kill existing server if has been started previously
        :param tmp_dir: (optional, default=None) specify base path for storing model artifact
        :param build_image: (optional, default=False) build image for standard model instead of directly mounting the model artifact to model container
        :param debug: (optional, default=False) enable debug mode that will print docker build log
        :return:
        """
        if tmp_dir is None:
            tmp_dir = "/tmp"
        artifact_path = f"{tmp_dir}/merlin/{self.model.project.name}/{self.model.name}/{self.id}/{DEFAULT_MODEL_PATH}"
        pathlib.Path(artifact_path).mkdir(parents=True, exist_ok=True)
        if len(os.listdir(artifact_path)) < 1:
            print(
                f"Downloading model artifact for model {self.model.name} version {self.id}"
            )
            self.download_artifact(artifact_path)

        # stop all previous containers to avoid port conflict
        client = docker.from_env()
        if kill_existing_server:
            started_containers = client.containers.list(
                filters={"name": self._container_name()}
            )
            for started_container in started_containers:
                print(f"Stopping model server {started_container.name}")
                started_container.remove(force=True)

        model_type = self.model.type
        if model_type == ModelType.PYFUNC:
            context_path = (
                f"{tmp_dir}/merlin/{self.model.project.name}/{self.model.name}"
            )

            conda_env = f"{tmp_dir}/merlin/{self.model.project.name}/{self.model.name}/{self.id}/{DEFAULT_MODEL_PATH}/conda.yaml"
            dependencies_path = f"{context_path}/env.yaml"
            shutil.copy(conda_env, dependencies_path)

            artifact_path = f"{self.id}/{DEFAULT_MODEL_PATH}"

            run_pyfunc_local_server(
                context_path=context_path,
                dependencies_path=dependencies_path,
                artifact_path=artifact_path,
                model_name=self.model.name,
                model_version=f"{self.id}",
                pyfunc_base_image=pyfunc_base_image,
                port=port,
                env_vars=env_vars,
                protocol=protocol,
                debug=debug,
            )
            return

        if (
            model_type == ModelType.TENSORFLOW
            or model_type == ModelType.XGBOOST
            or model_type == ModelType.SKLEARN
            or model_type == ModelType.PYTORCH
        ):
            self._run_standard_model_local_server(
                artifact_path, env_vars, port, build_image
            )
            return

        raise ValueError(
            f"running local model server is not supported for model type: {model_type}"
        )

    def _create_launch_command(self):
        model_type = self.model.type
        print(f"model type: {model_type}")
        if (
            model_type == ModelType.SKLEARN
            or model_type == ModelType.XGBOOST
            or model_type == ModelType.PYTORCH
        ):
            return f"--port=9000 --rest_api_port=8080 --model_name={self.model.name}-{self.id} --model_dir=/mnt/models"

        if model_type == ModelType.TENSORFLOW:
            return f"--port=9000 --rest_api_port=8080 --model_name={self.model.name}-{self.id} --model_base_path=/mnt/models"

        raise ValueError(f"unknown model type: {model_type}")

    def _run_standard_model_local_server(
        self, artifact_path, env_vars, port, build_image
    ):
        container: Optional[Container] = None  # type: ignore
        try:
            container_name = self._container_name()
            image_name = self.MODEL_TYPE_TO_IMAGE_MAP[self.model.type]
            cmd = self._create_launch_command()

            if build_image:
                apiClient = APIClient()
                image_tag = f"{self.model.project.name}-{self.model.name}:{self.id}"
                dockerfile_path = copy_standard_dockerfile(artifact_path)
                print(f"Building {self.model.type} image: {image_tag}")
                logs = apiClient.build(
                    path=artifact_path,
                    tag=image_tag,
                    buildargs={"BASE_IMAGE": image_name, "MODEL_PATH": artifact_path},
                    dockerfile=os.path.basename(dockerfile_path),
                    decode=True,
                )
                wait_build_complete(logs)
                image_name = image_tag

            print(f"Starting model server {container_name} at port: {port}")
            print(f"This process might take several minutes to complete")
            volumes = {artifact_path: {"bind": "/mnt/models", "mode": "rw"}}
            if build_image:
                volumes = None

            client = docker.from_env()
            container = client.containers.run(
                image_name,
                name=container_name,
                labels={"managed-by": "merlin"},
                command=cmd,
                ports={"8080/tcp": port},
                volumes=volumes,
                environment=env_vars,
                detach=True,
                remove=True,
            )

            # continously print docker log until the process is interrupted
            for log in container.logs(stream=True):
                print(log)
        finally:
            if container is not None:
                container.remove(force=True)

    def _container_name(self):
        return f"{self.model.project.name}-{self.model.name}-{self.id}"

    def _get_env_list(self) -> List[client.models.Environment]:
        return EnvironmentApi(self._api_client).environments_get()

    def _get_endpoint_in_environment(
        self, environment_name: Optional[str]
    ) -> Optional[VersionEndpoint]:
        """
        Return the FIRST endpoint of this model version that is deployed in the environment specified

        :param environment_name: environment name where the endpoint is deployed in
        :return: VersionEndpoint or None
        """
        for endpoint in self.version_endpoints:
            if endpoint.environment_name == environment_name:
                return endpoint
        return None

    @staticmethod
    def _get_default_resource_request(
        env_name: str, env_list: List[client.models.Environment]
    ) -> client.ResourceRequest:
        resource_request = None
        for env in env_list:
            if env.name == env_name and env.default_resource_request is not None:
                resource_request = ResourceRequest(
                    env.default_resource_request.min_replica,
                    env.default_resource_request.max_replica,
                    env.default_resource_request.cpu_request,
                    env.default_resource_request.cpu_limit,
                    env.default_resource_request.memory_request,
                )

        # This case is when the default resource request is not specified in the environment config
        if resource_request is None:
            raise ValueError(
                "default resource request not found in the environment config"
            )

        resource_request.validate()
        return client.ResourceRequest(
            min_replica=resource_request.min_replica,
            max_replica=resource_request.max_replica,
            cpu_request=resource_request.cpu_request,
            cpu_limit=resource_request.cpu_limit,
            memory_request=resource_request.memory_request,
        )

    @staticmethod
    def _get_default_autoscaling_policy(
        deployment_mode: str,
    ) -> client.AutoscalingPolicy:
        if deployment_mode == DeploymentMode.RAW_DEPLOYMENT.value:
            autoscaling_policy = RAW_DEPLOYMENT_DEFAULT_AUTOSCALING_POLICY
        else:
            autoscaling_policy = SERVERLESS_DEFAULT_AUTOSCALING_POLICY
        return client.AutoscalingPolicy(
            metrics_type=client.MetricsType(autoscaling_policy.metrics_type.value),
            target_value=autoscaling_policy.target_value,
        )

    @staticmethod
    def _add_env_vars(target_env_vars, new_env_vars):
        if not isinstance(new_env_vars, dict):
            raise ValueError(
                f"env_vars should be dictionary, got: {type(new_env_vars)}"
            )

        if len(new_env_vars) > 0:
            for name, value in new_env_vars.items():
                target_env_vars.append(client.EnvVar(name=str(name), value=str(value)))
        return target_env_vars

    @staticmethod
    def _create_transformer_spec(
        transformer: Transformer,
        target_env_name: str,
        env_list: List[client.models.Environment],
    ) -> client.Transformer:
        resource_request = transformer.resource_request
        if resource_request is None:
            target_resource_request = ModelVersion._get_default_resource_request(
                target_env_name, env_list
            )
        else:
            resource_request.validate()
            target_resource_request = client.ResourceRequest(
                min_replica=resource_request.min_replica,
                max_replica=resource_request.max_replica,
                cpu_request=resource_request.cpu_request,
                cpu_limit=resource_request.cpu_limit,
                memory_request=resource_request.memory_request,
            )

        target_env_vars: List[client.EnvVar] = []
        if transformer.env_vars is not None:
            target_env_vars = ModelVersion._add_env_vars(
                target_env_vars, transformer.env_vars
            )

        return client.Transformer(
            id=transformer.id,
            enabled=transformer.enabled,
            transformer_type=transformer.transformer_type.value,
            image=transformer.image,
            command=transformer.command,
            args=transformer.args,
            resource_request=target_resource_request,
            env_vars=target_env_vars,
        )

    def delete_model_version(self) -> int:
        """
        Delete this model version. Please note that any inactive related entity (endpoints and prediction jobs) will get deleted by this process.
        Deletion will fail if there are any active related entity (endpoints and prediction jobs)

        :return: id of deleted model
        """
        versionApi = VersionApi(self._api_client)
        return versionApi.models_model_id_versions_version_id_delete(
            int(self.model.id), int(self.id)
        )


def _get_default_target_env_name(env_list: List[client.models.Environment]) -> str:
    target_env_name = None
    for env in env_list:
        if env.is_default:
            target_env_name = env.name
    if target_env_name is None:
        raise ValueError(
            "Unable to find default environment, " "pass environment_name to the method"
        )
    return target_env_name
