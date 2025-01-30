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


from contextlib import contextmanager
from typing import Any, Dict, List, Optional

from client import PredictionJob
from merlin.autoscaling import AutoscalingPolicy
from merlin.batch.config import PredictionJobConfig
from merlin.client import MerlinClient
from merlin.deployment_mode import DeploymentMode
from merlin.endpoint import ModelEndpoint, VersionEndpoint
from merlin.environment import Environment
from merlin.logger import Logger
from merlin.model import Model, ModelType, ModelVersion, Project
from merlin.model_schema import ModelSchema
from merlin.protocol import Protocol
from merlin.resource_request import ResourceRequest
from merlin.mounted_mlp_secret import MountedMLPSecret
from merlin.transformer import Transformer
from merlin.version_image import VersionImage
from merlin.model_observability import ModelObservability

_merlin_client: Optional[MerlinClient] = None
_active_project: Optional[Project]
_active_model: Optional[Model]
_active_model_version: Optional[ModelVersion]


def set_url(url: str, use_google_oauth: bool = True):
    """
    Set Merlin URL

    :param url: Merlin URL
    """
    global _merlin_client
    _merlin_client = MerlinClient(url, use_google_oauth)


def get_url() -> Optional[str]:
    """
    Get currently active Merlin URL

    :return: merlin url if set, otherwise None
    """
    if _merlin_client is not None:
        return _merlin_client.url
    return None


def list_environment() -> List[Environment]:
    """
    List all available environment for deployment

    :return: List[Environment]
    """
    _check_active_client()
    return _merlin_client.list_environment()  # type: ignore


def get_environment(env_name: str) -> Environment:
    """
    Get environment for given env name

    :return: Environment or None
    """
    _check_active_client()
    envs = _merlin_client.list_environment()  # type: ignore
    for env in envs:
        if env.name == env_name:
            return env
    return None  # type: ignore


def get_default_environment() -> Optional[Environment]:
    """
    Get default environment

    :return: Environment or None
    """
    _check_active_client()
    return _merlin_client.get_default_environment()  # type: ignore


def list_project() -> List[Project]:
    """
    List all project in MLP

    :return: list of project
    """
    _check_active_client()
    return _merlin_client.list_project()  # type: ignore


def set_project(project_name: str):
    """
    Set active project

    :param project_name: project name. If project_name is not found, it will create the project.
    """
    _check_active_client()

    p = _merlin_client.get_project(project_name)  # type: ignore
    global _active_project
    _active_project = p


def active_project() -> Optional[Project]:
    """
    Get current active project

    :return: active project
    """
    _check_active_client()
    _check_active_project()
    return _active_project


def set_model(model_name, model_type: ModelType = None):
    """
    Set active model

    :param model_name: model name to be set as active model. If model name is not found, it will create the model.
    :param model_type: type of the model
    :return:
    """
    _check_active_project()
    active_project_name = _active_project.name  # type: ignore
    mdl = _merlin_client.get_or_create_model(  # type: ignore
        model_name, active_project_name, model_type
    )
    global _active_model
    _active_model = mdl


def active_model() -> Optional[Model]:
    """
    Get active model

    :return: active model
    """
    _check_active_client()
    _check_active_project()
    _check_active_model()

    return _active_model


@contextmanager
def new_model_version(
    labels: Dict[str, str] = None, model_schema: Optional[ModelSchema] = None
):
    """
    Create new model version under currently active model

    :param labels: dictionary containing the label that will be stored in the new model version
    :param model_schema: Detail schema specification of a model
    :return: ModelVersion
    """
    v = None
    try:
        _check_active_client()
        _check_active_project()
        _check_active_model()
        v = _merlin_client.new_model_version(_active_model.name, _active_project.name, labels, model_schema)  # type: ignore
        v.start()
        global _active_model_version
        _active_model_version = v
        yield v
    finally:
        _active_model_version = None
        if v is not None:
            v.finish()


def log_param(key: str, value: str):
    """
    Log parameter to the active model version

    :param key: parameter key
    :param value: parameter value
    """
    _check_active_model_version()
    _active_model_version.log_param(key, value)  # type: ignore


def log_metric(key: str, value: float):
    """
    Log a metric to the active model version

    :param key: metric key
    :param value: metric value
    """
    _check_active_model_version()
    _active_model_version.log_metric(key, value)  # type: ignore


def set_tag(key: str, value: str):
    """
    Set tag in the active model version

    :param key: tag name
    :param value: tag value
    """
    _check_active_model_version()
    _active_model_version.set_tag(key, value)  # type: ignore


def delete_tag(key: str):
    """
    Delete tag from the active model version

    :param key: tag name
    """
    _check_active_model_version()
    _active_model_version.delete_tag(key)  # type: ignore


def get_param(key: str) -> Optional[str]:
    """
    Get param value from the active model version

    :param key: param name
    """
    _check_active_model_version()
    return _active_model_version.get_param(key)  # type: ignore


def get_metric(key: str) -> Optional[float]:
    """
    Get metric value from the active model version

    :param key: metric name
    """
    _check_active_model_version()
    return _active_model_version.get_metric(key)  # type: ignore


def get_tag(key: str) -> Optional[str]:
    """
    Get tag value from the active model version

    :param key: tag name
    """
    _check_active_model_version()
    return _active_model_version.get_tag(key)  # type: ignore


def list_tag() -> Dict[str, str]:
    """
    Get list of tag from the active model version
    """
    _check_active_model_version()
    return _active_model_version.list_tag()  # type: ignore


def download_artifact(destination_path: str):
    """
    Download artifact from the active model version

    :param destination_path: destination of file when downloaded
    """
    _check_active_model_version()
    _active_model_version.download_artifact(destination_path)  # type: ignore


def log_artifact(local_path: str, artifact_path: str = None):
    """
    Log artifacts for the active model version

    :param local_path: directory to be uploaded into artifact store
    :param artifact_path: destination directory in artifact store
    """
    _check_active_model_version()
    _active_model_version.log_artifact(local_path, artifact_path)  # type: ignore


def log_pyfunc_model(
    model_instance: Any,
    conda_env: str,
    code_dir: List[str] = None,
    artifacts: Dict[str, str] = None,
):
    """
    Upload PyFunc based model into artifact storage.

    User has to specify model_instance and
    conda_env. `model_instance` shall implement all method specified in
    PyFuncModel. conda_env shall contain all dependency required by the
    model


    :param model_instance: instance of python function model
    :param conda_env: path to conda env.yaml file
    :param code_dir: additional code directory that will be loaded with ModelType.PYFUNC model
    :param artifacts: dictionary of artifact that will be stored together with the model. This will be passed to PythonModel.initialize. Example: {"config": "config/staging.yaml"}
    """
    _check_active_model_version()
    _active_model_version.log_pyfunc_model(  # type: ignore
        model_instance, conda_env, code_dir, artifacts
    )


def log_pytorch_model(model_dir: str, model_class_name: str = None):
    """
    Upload PyTorch model to artifact storage.

    :param model_dir: directory containing serialized PyTorch model
    :param model_class_name: class name of PyTorch model. By default the model class name is 'PyTorchModel'
    """
    _check_active_model_version()
    _active_model_version.log_pytorch_model(model_dir, model_class_name)  # type: ignore


def log_model(model_dir):
    """
    Upload model to artifact storage.
    This method is used to upload model for xgboost, tensorflow, and sklearn model.

    :param model_dir: directory which contain serialized model
    """
    _check_active_model_version()
    _active_model_version.log_model(model_dir)  # type: ignore


def log_custom_model(
    image: str, model_dir: str = None, command: str = "", args: str = ""
):
    """
    Upload model to artifact storage.
    This method is used to upload model for custom model type.

    :param image: Docker image that will be used as predictor
    :param model_dir: directory which contain serialized model
    :param command: Command to run docker image
    :param args: Arguments that needs to be specified when running docker
    """
    _check_active_model_version()
    _active_model_version.log_custom_model(  # type: ignore
        image=image, model_dir=model_dir, command=command, args=args
    )


def build_image(
    model_version: ModelVersion = None,
    backoff_limit: int = 0,
    resource_request: ResourceRequest = None,
) -> VersionImage:
    """
    Build image for a model version

    :param model_version: If model_version is not given it will build image for active model version
    :param backoff_limit: The maximum number of retries before considering a job as failed.
    :param resource_request: The resource requirement (CPU & memory) for image builder job.
    :return: VersionImage
    """
    _check_active_client()
    if model_version is None:
        _check_active_model_version()
        return _active_model_version.build_image(backoff_limit=backoff_limit, resource_request=resource_request)  # type: ignore

    return _merlin_client.build_image(model_version=model_version, backoff_limit=backoff_limit, resource_request=resource_request)  # type: ignore


def deploy(
    model_version: ModelVersion = None,
    environment_name: str = None,
    resource_request: ResourceRequest = None,
    image_builder_resource_request: ResourceRequest = None,
    env_vars: Dict[str, str] = None,
    secrets: List[MountedMLPSecret] = None,
    transformer: Transformer = None,
    logger: Logger = None,
    deployment_mode: DeploymentMode = None,
    autoscaling_policy: AutoscalingPolicy = None,
    protocol: Protocol = None,
    enable_model_observability: bool = False,
    model_observability: Optional[ModelObservability] = None,
) -> VersionEndpoint:
    """
    Deploy a model version.

    :param model_version: If model_version is not given it will deploy active model version
    :param environment_name: target environment to which the model version will be deployed to. If left empty it will deploy to default environment.
    :param resource_request: The resource requirement and replicas requests for model version endpoint.
    :param image_builder_resource_request: The resource requirement and replicas requests for image builder job.
    :param env_vars: List of environment variables to be passed to the model container.
    :param secrets: list of MLP secrets to mount into the prediction job environment as environment variables
    :param transformer: The service to be deployed alongside the model for pre/post-processing steps.
    :param logger: Response/Request logging configuration for model or transformer.
    :param deployment_mode: mode of deployment for the endpoint (default: DeploymentMode.SERVERLESS)
    :param autoscaling_policy: autoscaling policy to be used for the deployment (default: None)
    :param protocol: protocol to be used by the deployed model (default: HTTP_JSON)
    :param enable_model_observability: flag to determine whether model observability enabled for the endpoint
    :param model_observability: detail of model observability configuration
    :return: VersionEndpoint object
    """
    _check_active_client()
    if model_version is None:
        _check_active_model_version()
        return _active_model_version.deploy(  # type: ignore
            environment_name,
            resource_request,
            image_builder_resource_request,
            env_vars,
            secrets,
            transformer,
            logger,
            deployment_mode,
            autoscaling_policy,
            protocol,
            enable_model_observability,
            model_observability
        )

    return _merlin_client.deploy(  # type: ignore
        model_version,
        environment_name,
        resource_request,
        image_builder_resource_request,
        env_vars,
        secrets,
        transformer,
        logger,
        deployment_mode,
        autoscaling_policy,
        protocol,
        enable_model_observability,
        model_observability
    )


def undeploy(model_version=None, environment_name: str = None):
    """
    Delete deployment of a model version.

    :param model_version: model version to be undeployed. If model_version is not given it will undeploy active model version
    """
    _check_active_client()
    if model_version is None:
        _check_active_model_version()
        _active_model_version.undeploy(environment_name)  # type: ignore
        return

    _merlin_client.undeploy(model_version, environment_name)  # type: ignore


def serve_traffic(
    traffic_rule: Dict["VersionEndpoint", int], environment_name: str = None
) -> ModelEndpoint:
    """
    Update traffic rule of the active model.

    :param traffic_rule: dict of version endpoint and the percentage of traffic.
    :param environment_name: environment in which the traffic rule shall be applied
    :return: ModelEndpoint
    """
    _check_active_model()
    return _active_model.serve_traffic(traffic_rule, environment_name)  # type: ignore


def stop_serving_traffic(environment_name: str = None):
    """
    Stop serving traffic for a given model endpoint in given environment.

    :param environment_name: environment in which the model endpoint will be stopped.
    """
    _check_active_model()
    return _active_model.stop_serving_traffic(environment_name)  # type: ignore


def set_traffic(traffic_rule: Dict[ModelVersion, int]) -> ModelEndpoint:
    """
     Update traffic rule of the active model.

    :param traffic_rule: dict of model version and the percentage of traffic.
    :return: ModelEndpoint
    """
    _check_active_model()
    return _active_model.set_traffic(traffic_rule)  # type: ignore


def list_model_endpoints() -> List[ModelEndpoint]:
    """
    Get list of all serving model endpoints.

    :return: List of model endpoints.
    """
    _check_active_model()
    return _active_model.list_endpoint()  # type: ignore


def create_prediction_job(
    job_config: PredictionJobConfig, sync: bool = True
) -> PredictionJob:
    """

    :param sync:
    :param job_config:
    :return:
    """
    _check_active_client()
    _check_active_project()
    _check_active_model()
    _check_active_model_version()

    return _active_model_version.create_prediction_job(job_config=job_config, sync=sync)  # type: ignore


def _check_active_project():
    if _active_project is None:
        raise Exception("Active project isn't set, use set_project(...) to set it")


def _check_active_client():
    if _merlin_client is None:
        raise Exception("URL is not set, use set_url(...) to set it")


def _check_active_model():
    if _active_model is None:
        raise Exception("Active model isn't set, use set_model(...) to set it")


def _check_active_model_version():
    if _active_model_version is None:
        raise Exception(
            "Active model version isn't set, use new_model_version(...) to " "create it"
        )
