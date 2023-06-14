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

import warnings
from sys import version_info
from typing import Dict, List, Optional

import urllib3
from caraml_auth.id_token_credentials import get_default_id_token_credentials
from client import (ApiClient, Configuration, EndpointApi, EnvironmentApi,
                    ModelsApi, ProjectApi, VersionApi)
from google.auth.transport.requests import Request
from google.auth.transport.urllib3 import AuthorizedHttp
from merlin.autoscaling import AutoscalingPolicy
from merlin.deployment_mode import DeploymentMode
from merlin.endpoint import VersionEndpoint
from merlin.environment import Environment
from merlin.logger import Logger
from merlin.model import Model, ModelType, ModelVersion, Project
from merlin.protocol import Protocol
from merlin.resource_request import ResourceRequest
from merlin.transformer import Transformer
from merlin.util import valid_name_check
from merlin.version import VERSION


class MerlinClient:
    def __init__(self, merlin_url: str, use_google_oauth: bool=True):
        self._merlin_url = merlin_url
        config = Configuration()
        config.host = self._merlin_url + "/v1"

        self._api_client = ApiClient(config)
        if use_google_oauth:
            credentials = get_default_id_token_credentials(target_audience="sdk.caraml")
            # Refresh credentials, in case it's coming from Compute Engine.
            # See: https://github.com/googleapis/google-auth-library-python/issues/1211
            credentials.refresh(Request())
            authorized_http = AuthorizedHttp(credentials, urllib3.PoolManager())
            self._api_client.rest_client.pool_manager = authorized_http

        python_version = f'{version_info.major}.{version_info.minor}.{version_info.micro}'  # capture user's python version
        self._api_client.user_agent = f"merlin-sdk/{VERSION} python/{python_version}"
        self._project_api = ProjectApi(self._api_client)
        self._model_api = ModelsApi(self._api_client)
        self._version_api = VersionApi(self._api_client)
        self._endpoint_api = EndpointApi(self._api_client)
        self._env_api = EnvironmentApi(self._api_client)

    @property
    def url(self):
        return self._merlin_url

    def list_environment(self) -> List[Environment]:
        """
        List all available environment for deployment

        :return: list of Environment
        """
        env_list = self._env_api.environments_get()
        envs = []
        for env in env_list:
            envs.append(Environment(env))
        return envs


    def get_environment(self, env_name: str) -> Optional[Environment]:
        """
        Get environment for given env name

        :return: Environment or None
        """
        env_list = self._env_api.environments_get()
        for env in env_list:
            if env.name == env_name:
                return env
        return None

    def get_default_environment(self) -> Optional[Environment]:
        """
        Return default environment

        :return: Environment or None
        """
        env_list = self._env_api.environments_get()
        for env in env_list:
            if env.is_default:
                return Environment(env)
        return None

    def list_project(self) -> List[Project]:
        """
        List project in the connected MLP server

        :return: list of Project
        """
        p_list = self._project_api.projects_get()
        result = []
        for p in p_list:
            result.append(Project(p, self.url, self._api_client))
        return result

    def get_or_create_project(self, project_name: str) -> Project:
        warnings.warn(
            "get_or_create_project is deprecated please use get_project",
            category=DeprecationWarning,
        )
        return self.get_project(project_name)

    def get_project(self, project_name: str) -> Project:
        """
        Get a project in Merlin and optionally assign list of readers and administrators.
        The identity used for creating the project will be automatically included as project's administrators.

        :param project_name: project name
        :return: project
        """
        if not valid_name_check(project_name):
            raise ValueError(
                '''Your project/model name contains invalid characters.\
                    \nUse only the following characters\
                    \n- Characters: a-z (Lowercase ONLY)\
                    \n- Numbers: 0-9\
                    \n- Symbols: -
                '''
            )

        p_list = self._project_api.projects_get(name=project_name)
        p = None
        for prj in p_list:
            if prj.name == project_name:
                p = prj

        if p is None:
            raise Exception(f"{project_name} does not exist or you don't have access to the project. Please create new "
                            f"project using MLP console or ask the project's administrator to be able to access "
                            f"existing project.")

        return Project(p, self.url, self._api_client)

    def get_model(self, model_name: str, project_name: str) \
            -> Optional[Model]:
        """
        Get model with given name

        :param model_name: model name to be retrieved
        :param project_name: project name
        :return: Model or None
        """
        prj = self.get_project(project_name)
        m_list = self._model_api.projects_project_id_models_get(
            project_id=int(prj.id), name=model_name)
        model = m_list[0]
        if len(m_list) == 0:
            return None

        for mdl in m_list:
            if mdl.name == model_name:
                model = mdl

        return Model(model, prj, self._api_client)

    def get_or_create_model(self, model_name: str,
                            project_name: str,
                            model_type: ModelType = None) -> Model:
        """
        Get or create a model under a project

        If project_name is not given it will use currently active project
        otherwise will raise Exception

        :param model_type:
        :param model_name: model name
        :param project_name: project name (optional)
        :param model_type: model type
        :return: Model
        """
        if not valid_name_check(model_name):
            raise ValueError(
                '''Your project/model name contains invalid characters.\
                    \nUse only the following characters\
                    \n- Characters: a-z (Lowercase ONLY)\
                    \n- Numbers: 0-9\
                    \n- Symbols: -
                '''
            )

        prj = self.get_project(project_name)
        m_list = self._model_api.projects_project_id_models_get(
            project_id=int(prj.id), name=model_name)

        model = None
        for mdl in m_list:
            if mdl.name == model_name:
                model = mdl

        if model is None:
            if model_type is None:
                raise ValueError(f"model {model_name} is not found, specify "
                                 f"{model_type} to create it")
            model = self._model_api.projects_project_id_models_post(
                project_id=int(prj.id), body={
                    "name": model_name,
                    "type": model_type.value
                })

        return Model(model, prj, self._api_client)

    def new_model_version(self, model_name: str, project_name: str, labels: Dict[str, str] = None) \
            -> ModelVersion:
        """
        Create new model version for the given model and project

        :param model_name:
        :param project_name:
        :param labels:
        :return:
        """
        mdl = self.get_model(model_name, project_name)
        if mdl is None:
            raise ValueError(f"Model with name: {model_name} is not found")
        return mdl.new_model_version(labels=labels)

    def deploy(self, model_version: ModelVersion,
               environment_name: str = None,
               resource_request: ResourceRequest = None,
               env_vars: Dict[str, str] = None,
               transformer: Transformer = None,
               logger: Logger = None,
               deployment_mode: DeploymentMode = DeploymentMode.SERVERLESS,
               autoscaling_policy: AutoscalingPolicy = None,
               protocol: Protocol = Protocol.HTTP_JSON) -> VersionEndpoint:
        return model_version.deploy(environment_name, resource_request, env_vars, transformer, logger, deployment_mode,
                                    autoscaling_policy, protocol)

    def undeploy(self, model_version: ModelVersion,
                 environment_name: str = None):
        model_version.undeploy(environment_name)
