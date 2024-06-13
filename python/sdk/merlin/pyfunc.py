import os
import shutil
from abc import abstractmethod
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional, Union

import docker
import grpc
import numpy
import pandas
from caraml.upi.v1 import upi_pb2
from docker import APIClient
from merlin.docker.docker import copy_pyfunc_dockerfile, wait_build_complete
from merlin.protocol import Protocol
from merlin.version import VERSION
from mlflow.pyfunc import PythonModel

import mlflow

PYFUNC_EXTRA_ARGS_KEY = "__EXTRA_ARGS__"
PYFUNC_MODEL_INPUT_KEY = "__INPUT__"
PYFUNC_PROTOCOL_KEY = "__PROTOCOL__"
PYFUNC_GRPC_CONTEXT = "__GRPC_CONTEXT__"


class PyFuncModel(PythonModel):
    def load_context(self, context):
        """
        Override method of PythonModel `load_context`. This method load artifacts from context
        that can be used in predict function. This method is called by internal MLflow when an MLflow
        is loaded.

        :param context: A :class:`~PythonModelContext` instance containing artifacts that the model
                        can use to perform inference
        """
        self.initialize(context.artifacts)
        self._use_kwargs_infer = True

    def predict(self, context, model_input):
        """
        Implementation of PythonModel `predict` method. This method evaluates model input and produces model output.

        :param context: A :class:`~PythonModelContext` instance containing artifacts that the model
                        can use to perform inference
        :param model_input: A pyfunc-compatible input for the model to evaluate.
        """
        extra_args = model_input.get(PYFUNC_EXTRA_ARGS_KEY, {})
        input = model_input.get(PYFUNC_MODEL_INPUT_KEY, {})
        protocol = model_input.get(PYFUNC_PROTOCOL_KEY, Protocol.HTTP_JSON)

        if protocol == Protocol.HTTP_JSON:
            if extra_args is not None:
                return self._do_http_predict(input, **extra_args)
            return self._do_http_predict(input)
        elif protocol == Protocol.UPI_V1:
            grpc_context = model_input.get(PYFUNC_GRPC_CONTEXT, None)
            grpc_response = self.upiv1_infer(input, grpc_context)
            return grpc_response
        else:
            raise NotImplementedError(f"protocol {protocol} is not supported")

    def _do_http_predict(self, model_input, **kwargs):
        if self._use_kwargs_infer:
            try:
                http_response = self.infer(model_input, **kwargs)
                return http_response
            except TypeError as e:
                if "infer() got an unexpected keyword argument" in str(e):
                    print(
                        "Fallback to the old infer() method, got TypeError exception: {}".format(
                            e
                        )
                    )
                    self._use_kwargs_infer = False
                else:
                    raise e

        http_response = self.infer(model_input)
        return http_response

    @abstractmethod
    def initialize(self, artifacts: dict):
        """
        Implementation of PyFuncModel can specify initialization step which
        will be called one time during model initialization.

        :param artifacts: dictionary of artifacts passed to log_model method
        """
        pass

    @abstractmethod
    def infer(self, request: dict, **kwargs) -> dict:
        """
        Do inference.

        This method MUST be implemented by concrete implementation of
        PyFuncModel to support HTTP_JSON protocol.
        This method accept 'request' which is the body content of incoming
        request.
        Implementation should return inference a json object of response.

        :param request: Dictionary containing incoming request body content
        :param **kwargs: See below.

        :return: Dictionary containing response body

        :keyword arguments:
        * headers (dict): Dictionary containing incoming HTTP request headers
        """
        raise NotImplementedError("infer is not implemented")

    @abstractmethod
    def upiv1_infer(
        self, request: upi_pb2.PredictValuesRequest, context: grpc.ServicerContext
    ) -> upi_pb2.PredictValuesResponse:
        """
        Do inference.

        This method MUST be implemented by concrete implementation of PyFunModel to support UPI_V1 protocol.
        The method accept request in for of PredictValuesRequest proto and should return PredictValuesResponse response proto.

        :param request: Inference request as PredictValuesRequest
        :param context: grpc context
        :return: Prediction result as PredictValuesResponse proto
        """
        raise NotImplementedError("upiv1_infer is not implemented")


@dataclass
class Values:
    columns: List[str]
    data: Union[List[List], numpy.ndarray]

    def to_dict(self):
        data = self.data
        if isinstance(self.data, numpy.ndarray):
            data = self.data.tolist()

        return {"columns": self.columns, "data": data}


@dataclass
class ModelInput:
    # features data for model prediction, the length of data of features must be the same as `prediction_ids` length
    features: Union[Values, pandas.DataFrame]
    # unique identifier of each prediction
    # will be deprecate soon
    prediction_ids: Optional[List[str]] = None
    # entities data is additional data that are not used for prediction, but this data is used to retrieved another features.
    # The length of data of entities must be the same as `prediction_ids` length
    entities: Optional[Union[Values, pandas.DataFrame]] = None
    # session id is identifier for the request
    session_id: str = ""
    # unique identifier for each row in a prediction
    row_ids: List[str] = field(default_factory=lambda: [])

    def _get_row_ids(self) -> List[str]:
        row_ids = self.row_ids
        if len(row_ids) == 0 and self.prediction_ids is not None:
            row_ids = self.prediction_ids
        return row_ids

    def features_dict(self) -> Optional[dict]:
        if self.features is None:
            return None

        val = self.features
        if isinstance(self.features, pandas.DataFrame):
            val = Values(
                self.features.columns.values.tolist(), self.features.values.tolist()
            )

        result = val.to_dict()
        result["row_ids"] = self._get_row_ids()
        return result

    def entities_dict(self) -> Optional[dict]:
        if self.entities is None:
            return None
        val = self.entities
        if isinstance(self.entities, pandas.DataFrame):
            val = Values(
                self.entities.columns.values.tolist(), self.entities.values.tolist()
            )
        result = val.to_dict()
        result["row_ids"] = self._get_row_ids()
        return result


@dataclass
class ModelOutput:
    # predictions contains prediction output from ml_predict
    # it may contains multiple columns e.g for multiclass classification or for binary classification that contains prediction score and label
    # length of the data must be the same as predicion_ids
    predictions: Values
    # unique identifier of each prediction
    # will be deprecate and use `row_ids` instead
    prediction_ids: Optional[List[str]] = None
    # unique identifier for each row in a prediction
    row_ids: List[str] = field(default_factory=lambda: [])

    def predictions_dict(self) -> dict:
        if self.predictions is None:
            return None

        predictions = self.predictions
        if isinstance(self.predictions, pandas.DataFrame):
            predictions = Values(
                self.predictions.columns.values.tolist(),
                self.predictions.values.tolist(),
            )

        result = predictions.to_dict()
        row_ids = self.row_ids
        if len(row_ids) == 0 and self.prediction_ids is not None:
            row_ids = self.prediction_ids

        result["row_ids"] = row_ids
        return result


@dataclass
class PyFuncOutput:
    # final pyfunc response payload when using `HTTP_JSON` protocol
    http_response: Optional[dict] = None
    # final pyfunc response payload when using `UPI_V!` protocol
    upi_response: Optional[upi_pb2.PredictValuesResponse] = None
    # model input contains information about features, entities
    model_input: Optional[ModelInput] = None
    # model output contains information about ml prediction output
    model_output: Optional[ModelOutput] = None

    def get_session_id(self) -> Optional[str]:
        if self.model_input is None:
            return ""
        return self.model_input.session_id

    def contains_prediction_log(self):
        return self.model_input is not None and self.model_output is not None


class PyFuncV3Model(PythonModel):
    def load_context(self, context):
        """
        Override method of PythonModel `load_context`. This method load artifacts from context
        that can be used in predict function. This method is called by internal MLflow when an MLflow
        is loaded.

        :param context: A :class:`~PythonModelContext` instance containing artifacts that the model
                        can use to perform inference
        """
        self.initialize(context.artifacts)

    def predict(self, context, model_input):
        """
        Implementation of PythonModel `predict` method. This method evaluates model input and produces model output.

        :param context: A :class:`~PythonModelContext` instance containing artifacts that the model
                        can use to perform inference
        :param model_input: A pyfunc-compatible input for the model to evaluate.
        """
        extra_args = model_input.get(PYFUNC_EXTRA_ARGS_KEY, {})
        input = model_input.get(PYFUNC_MODEL_INPUT_KEY, {})
        protocol = model_input.get(PYFUNC_PROTOCOL_KEY, Protocol.HTTP_JSON)

        if protocol == Protocol.HTTP_JSON:
            if extra_args is not None:
                return self._do_http_predict(input, **extra_args)
            return self._do_http_predict(input)
        elif protocol == Protocol.UPI_V1:
            grpc_context = model_input.get(PYFUNC_GRPC_CONTEXT, None)
            ml_model_input = self.upiv1_preprocess(input, grpc_context)
            ml_model_output = self.infer(ml_model_input)
            final_output = self.upiv1_postprocess(ml_model_output, input)
            return PyFuncOutput(
                upi_response=final_output,
                model_input=ml_model_input,
                model_output=ml_model_output,
            )
        else:
            raise NotImplementedError(f"protocol {protocol} is not supported")

    def _do_http_predict(self, model_input, **kwargs):
        ml_model_input = self.preprocess(model_input, **kwargs)
        ml_model_output = self.infer(ml_model_input)
        final_output = self.postprocess(ml_model_output, model_input)

        return PyFuncOutput(
            http_response=final_output,
            model_input=ml_model_input,
            model_output=ml_model_output,
        )

    @abstractmethod
    def initialize(self, artifacts: dict):
        """
        Implementation of PyFuncModel can specify initialization step which
        will be called one time during model initialization.

        :param artifacts: dictionary of artifacts passed to log_model method
        """
        pass

    @abstractmethod
    def preprocess(self, request: dict, **kwargs) -> ModelInput:
        """
        preprocess is the method that doing preprocessing before calling the ml framework for prediction.
        This method is called when user use HTTP_JSON protocol
        :param request: raw input to the model (dict)
        :return: model input, this model input must already have all the features that required for model prediction
        """
        raise NotImplementedError("preprocess is not implemented")

    @abstractmethod
    def infer(self, model_input: ModelInput) -> ModelOutput:
        """
        infer is the method that will call the respective ml framework to do prediction.
        Since there are various types that supported by each ml framework, user must do conversion from the given model input into acceptable input for the model
        :param model_input: ModelInput that is produced by the `preprocess` method
        :return: model output
        """
        raise NotImplementedError("infer is not implemented")

    @abstractmethod
    def postprocess(self, model_output: ModelOutput, request: dict) -> dict:
        """
        postprocess is the method that is caled after `infer`, the main function of this method is to do postprocessing
        that may including build overall pyfunc response payload, additional computation based on the model output
        :param model_output: the output of the `infer` function
        :param request: raw request payload
        :return: final payload in dictionary type
        """
        raise NotImplementedError("postprocess is not implemented")

    @abstractmethod
    def upiv1_preprocess(
        self, request: upi_pb2.PredictValuesRequest, context: grpc.ServicerContext
    ) -> ModelInput:
        """
        upiv1_preprocess is the preprocessing method that only applicable for UPI_V1 protocol.
        basically the method is the same with `preprocess` the difference is on the type of the incoming request
        for `upiv1_preprocess` the type of request is `PredictValuesRequest`
        :param request: raw request payload
        :param context: grpc request context
        :return: model input
        """
        raise NotImplementedError("upiv1_preprocess is not implemented")

    @abstractmethod
    def upiv1_postprocess(
        self, model_output: ModelOutput, request: upi_pb2.PredictValuesRequest
    ) -> upi_pb2.PredictValuesResponse:
        """
        upiv1_postprocess is the postprocessing method that only applicable for UPI_V1 protocol.
        :param model_output: the output of the `infer` function
        :param request: raw request payload
        :return: final payload in `PredictValuesResponse` type
        """
        raise NotImplementedError("upiv1_postprocess is not implemented")


class PyFuncV2Model(PythonModel):
    def load_context(self, context):
        """
        Override method of PythonModel `load_context`. This method load artifacts from context
        that can be used in predict function. This method is called by internal MLflow when an MLflow
        is loaded.

        :param context: A :class:`~PythonModelContext` instance containing artifacts that the model
                        can use to perform inference
        """
        self.initialize(context.artifacts)

    def predict(self, context, model_input):
        """
        Implementation of PythonModel `predict` method. This method evaluates model input and produces model output.

        :param context: A :class:`~PythonModelContext` instance containing artifacts that the model
                        can use to perform inference
        :param model_input: A pyfunc-compatible input for the model to evaluate.
        """
        return self.infer(model_input)

    def initialize(self, artifacts: dict):
        """
        Implementation of PyFuncModel can specify initialization step which
        will be called one time during model initialization.

        :param artifacts: dictionary of artifacts passed to log_model method
        """
        pass

    def infer(
        self, model_input: pandas.DataFrame
    ) -> Union[numpy.ndarray, pandas.Series, pandas.DataFrame]:
        """
        Infer method is the main method that will be called when calculating
        the inference result for both online prediction and batch
        prediction. The method accepts pandas Dataframe and returns either
        another panda Dataframe / pandas Series / ndarray of the same length
        as the input. In the batch prediction case the model_input will
        contain an arbitrary partition of the whole dataset that the user
        defines as the data source. As such, it is advisable not to do
        aggregation within the infer method, as it will be incorrect since
        it will only apply to the partition in contrary to the whole dataset.

        :param model_input: input to the model (pandas.DataFrame)
        :return: inference result as numpy.ndarray or pandas.Series or pandas.DataFrame

        """
        raise NotImplementedError("infer is not implemented")


def run_pyfunc_model(
    model_instance: Any,
    conda_env: str,
    code_dir: List[str] = None,
    artifacts: Dict[str, str] = None,
    pyfunc_base_image: str = None,
    port: int = 8080,
    env_vars: Dict[str, str] = None,
    protocol: Protocol = Protocol.HTTP_JSON,
    debug: bool = False,
):
    """
    Run pyfunc model locally using Docker. The function will log the model artifacts onto local mlflow, build a docker image, and run the container locally.

    :param model_instance: instance of python function model
    :param conda_env: path to conda env.yaml file
    :param code_dir: additional code directory that will be loaded with ModelType.PYFUNC model
    :param artifacts: dictionary of artifact that will be stored together with the model. This will be passed to PythonModel.initialize. Example: {"config": "config/staging.yaml"}
    :param pyfunc_base_image: base image for building pyfunc model
    :param port: port to expose the model
    :param env_vars: dictionary of environment variables to be passed to the server
    :param protocol: protocol to be used by the deployed model (default: HTTP_JSON)
    :param debug: flag to enable debug mode that will print docker build log
    """

    model_name = str.lower(model_instance.__class__.__name__)

    # Log model to local mlflow
    print("Logging model to local MLflow")
    tracking_uri = f"file:///tmp/merlin/pyfunc-models"
    mlflow.set_tracking_uri(tracking_uri)
    experiment = mlflow.set_experiment(model_name)

    model_info = mlflow.pyfunc.log_model(
        artifact_path="model",
        python_model=model_instance,
        conda_env=conda_env,
        code_path=code_dir,
        artifacts=artifacts,
    )

    context_path = experiment.artifact_location
    if "file://" in context_path:
        context_path = context_path.replace("file://", "")

    dependencies_path = f"{context_path}/env.yaml"
    shutil.copy(conda_env, dependencies_path)

    artifact_path = f"{model_info.run_id}/artifacts/model"

    run_pyfunc_local_server(
        context_path=context_path,
        dependencies_path=dependencies_path,
        artifact_path=artifact_path,
        model_name=model_name,
        model_version="dev",
        pyfunc_base_image=pyfunc_base_image,
        port=port,
        env_vars=env_vars,
        protocol=protocol,
        debug=debug,
    )


def run_pyfunc_local_server(
    context_path: str,
    dependencies_path: str,
    artifact_path: str,
    model_name: str,
    model_version: str,
    pyfunc_base_image: str = None,
    port: int = 8080,
    env_vars: Dict[str, str] = None,
    protocol: Protocol = Protocol.HTTP_JSON,
    debug: bool = False,
):
    if pyfunc_base_image is None:
        if "dev" in VERSION or "0.0.0" in VERSION:
            pyfunc_base_image = "ghcr.io/caraml-dev/merlin/merlin-pyfunc-base:0.41.0"
        else:
            pyfunc_base_image = (
                f"ghcr.io/caraml-dev/merlin/merlin-pyfunc-base:{VERSION}"
            )

    dockerfile_path = copy_pyfunc_dockerfile(context_path)

    _clone_merlin_repo(context_path)

    image_tag = f"{model_name}-{model_version}"

    print(f"Building Docker image {image_tag}")
    if debug:
        print("Context path:", context_path)
        print("Dockerfile path:", dockerfile_path)
        print("Dependencies path:", dependencies_path)
        print("Model artifacts path:", artifact_path)

    _build_image(
        image_tag=image_tag,
        context_path=context_path,
        dependencies_path=dependencies_path,
        artifact_path=artifact_path,
        pyfunc_base_image=pyfunc_base_image,
        dockerfile_path=dockerfile_path,
        debug=debug,
    )

    print("Running PyFunc local server")
    _run_container(
        image_tag=image_tag,
        model_name=model_name,
        model_version=model_version,
        model_full_name=f"{model_name}-{model_version}",
        port=port,
        env_vars=env_vars,
        protocol=protocol,
    )


def _clone_merlin_repo(context_path: str):
    # Can't be imported at module level because the initialization will fail when the git executable
    # does not exists.
    from git import Repo

    repo_url = "https://github.com/caraml-dev/merlin.git"
    repo_path = f"{context_path}/merlin"
    if not os.path.isdir(repo_path):
        repo = Repo.clone_from(repo_url, repo_path)
        if "dev" in VERSION or "0.0.0" in VERSION:
            repo.git.checkout("main")
        else:
            repo.git.checkout(f"v{VERSION}")


def _build_image(
    image_tag,
    context_path,
    dependencies_path,
    artifact_path,
    pyfunc_base_image,
    dockerfile_path,
    debug,
):
    docker_api_client = APIClient()
    logs = docker_api_client.build(
        path=context_path,
        tag=image_tag,
        buildargs={
            "BASE_IMAGE": pyfunc_base_image,
            "MODEL_DEPENDENCIES_URL": os.path.basename(dependencies_path),
            "MODEL_ARTIFACTS_URL": artifact_path,
        },
        dockerfile=os.path.basename(dockerfile_path),
        decode=True,
    )

    wait_build_complete(logs, debug)


def _run_container(
    image_tag,
    model_name,
    model_version,
    model_full_name,
    port,
    env_vars: Dict[str, str] = None,
    protocol: Protocol = Protocol.HTTP_JSON,
):
    docker_client = docker.from_env()

    # Stop all previous containers to avoid port conflict
    started_containers = docker_client.containers.list(
        filters={"name": model_full_name}
    )
    for started_container in started_containers:
        started_container.remove(force=True)

    if env_vars is None:
        env_vars = {}

    env_vars["CARAML_HTTP_PORT"] = "8080"
    env_vars["CARAML_GRPC_PORT"] = "9000"
    env_vars["CARAML_MODEL_NAME"] = model_name
    env_vars["CARAML_MODEL_VERSION"] = model_version
    env_vars["CARAML_MODEL_FULL_NAME"] = model_full_name
    env_vars["WORKERS"] = "1"

    ports = {"8080/tcp": port}
    if protocol == Protocol.UPI_V1:
        env_vars["CARAML_PROTOCOL"] = protocol.value
        ports = {"9000/tcp": port}

    try:
        container = docker_client.containers.run(
            image=image_tag,
            name=model_name,
            labels={"managed-by": "merlin"},
            ports=ports,
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
