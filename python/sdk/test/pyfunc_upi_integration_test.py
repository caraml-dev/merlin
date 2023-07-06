import os
import uuid
from test.utils import undeploy_all_version
from time import sleep
from typing import List

import grpc
import merlin
import pandas as pd
import pytest
import xgboost as xgb
from caraml.upi.utils import df_to_table, table_to_df
from caraml.upi.v1 import type_pb2, upi_pb2, upi_pb2_grpc, variable_pb2
from merlin.deployment_mode import DeploymentMode
from merlin.endpoint import Status
from merlin.model import ModelType, PyFuncModel
from merlin.protocol import Protocol
from merlin.transformer import StandardTransformer
from sklearn.datasets import load_iris


class IrisClassifier(PyFuncModel):
    target_names = [
        "setosa",
        "versicolor",
        "virginica"
    ]

    target_name = "iris-species"

    def initialize(self, artifacts):
        self._model = xgb.Booster(model_file=artifacts["xgb_model"])

    def infer(self, request: dict, **kwargs):
        result = self._predict(request['instances'])
        return {
            "predictions": result.tolist()
        }

    def upiv1_infer(self, request: upi_pb2.PredictValuesRequest,
                    context: grpc.ServicerContext) -> upi_pb2.PredictValuesResponse:
        features_df, _ = table_to_df(request.prediction_table)
        prediction_result_df = self._predict(features_df)
        return self._create_response(prediction_result_df, request)

    def _create_response(self, predictions: pd.DataFrame, request: upi_pb2.PredictValuesRequest) -> upi_pb2.PredictValuesResponse:
        prediction_result_table = df_to_table(predictions, "prediction_result")
        response_metadata = upi_pb2.ResponseMetadata(prediction_id=request.metadata.prediction_id)
        return upi_pb2.PredictValuesResponse(prediction_result_table=prediction_result_table, target_name=self.target_name, metadata=response_metadata)

    def _predict(self, features: pd.DataFrame) -> pd.DataFrame:
        features_matrix = xgb.DMatrix(features)
        return pd.DataFrame(self._model.predict(features_matrix), columns = self.target_names)


class SimpleForwarder(PyFuncModel):
    target_name = "probability"
    def infer(self, request: dict, **kwargs):
        return request

    def upiv1_infer(self, request: upi_pb2.PredictValuesRequest,
                    context: grpc.ServicerContext) -> upi_pb2.PredictValuesResponse:
        return upi_pb2.PredictValuesResponse(prediction_result_table=request.prediction_table, target_name=self.target_name)



@pytest.mark.pyfunc
@pytest.mark.integration
def test_deploy(integration_test_url, project_name, use_google_oauth, requests):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("pyfunc-upi", ModelType.PYFUNC)

    undeploy_all_version()
    with merlin.new_model_version() as v:
        xgb_path, xgb_model = train_xgboost_model()

        v.log_pyfunc_model(model_instance=IrisClassifier(),
                           conda_env="test/pyfunc/env.yaml",
                           code_dir=["test"],
                           artifacts={"xgb_model": xgb_path})

    endpoint = merlin.deploy(v, protocol=Protocol.UPI_V1)

    assert endpoint.protocol == Protocol.UPI_V1
    assert endpoint.status == Status.RUNNING
    assert endpoint.deployment_mode == DeploymentMode.SERVERLESS

    channel = grpc.insecure_channel(f"{endpoint.url}:80")
    stub = upi_pb2_grpc.UniversalPredictionServiceStub(channel)

    validate_iris_upi(xgb_model, stub)
    merlin.undeploy(v)

@pytest.mark.pyfunc
@pytest.mark.integration
def test_pyfunc_with_standard_transformer(integration_test_url, project_name, use_google_oauth, requests):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("pyfunc-upi-std", ModelType.PYFUNC)

    undeploy_all_version()
    with merlin.new_model_version() as v:
        v.log_pyfunc_model(model_instance=SimpleForwarder(),
                           conda_env="test/pyfunc/env.yaml",
                           code_dir=["test"])

    transformer_config_path = os.path.join(
        "test/transformer", "upi_standard_transformer_no_feast.yaml"
    )
    transformer = StandardTransformer(config_file=transformer_config_path, enabled=True)
    endpoint = merlin.deploy(v, transformer=transformer, protocol=Protocol.UPI_V1)

    assert endpoint.protocol == Protocol.UPI_V1
    assert endpoint.status == Status.RUNNING

    channel = grpc.insecure_channel(f"{endpoint.url}:80")
    stub = upi_pb2_grpc.UniversalPredictionServiceStub(channel)

    print(endpoint.url)
    sleep(5)

    request = create_simple_forwarder_request()
    response = stub.PredictValues(request=request)

    assert response.metadata.prediction_id == request.metadata.prediction_id
    assert response.target_name == request.target_name
    exp_response = simple_forwarder_response()
    assert response == exp_response

    merlin.undeploy(v)


@pytest.mark.pyfunc
@pytest.mark.integration
def test_serve_traffic(integration_test_url, project_name, use_google_oauth, requests):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("pyfunc-upi-serve", ModelType.PYFUNC)

    undeploy_all_version()
    with merlin.new_model_version() as v:
        xgb_path, xgb_model = train_xgboost_model()

        v.log_pyfunc_model(model_instance=IrisClassifier(),
                           conda_env="test/pyfunc/env.yaml",
                           code_dir=["test"],
                           artifacts={"xgb_model": xgb_path})

    endpoint = merlin.deploy(v, protocol=Protocol.UPI_V1)
    model_endpoint = merlin.serve_traffic({endpoint:100})

    assert model_endpoint.protocol == Protocol.UPI_V1
    assert model_endpoint.status == Status.SERVING

    channel = grpc.insecure_channel(f"{model_endpoint.url}:80")
    stub = upi_pb2_grpc.UniversalPredictionServiceStub(channel)

    print(model_endpoint.url)
    sleep(5)
    validate_iris_upi(xgb_model, stub)

    merlin.stop_serving_traffic(model_endpoint.environment_name)
    merlin.undeploy(v)


def test_model():
    model_path, xgb_model = train_xgboost_model()
    pyfunc_model = IrisClassifier()
    pyfunc_model.initialize({"xgb_model": model_path})
    request = create_upi_request_from_iris_dataset()
    response = pyfunc_model.upiv1_infer(request, {})

    X = load_iris()['data']
    y = xgb_model.predict(xgb.DMatrix(X))
    exp_df = pd.DataFrame(y, columns=IrisClassifier.target_names)
    exp_table = df_to_table(exp_df, "prediction_result")

    assert exp_table == response.prediction_result_table


def validate_iris_upi(model, stub):
    request = create_upi_request_from_iris_dataset()
    response = stub.PredictValues(request=request)

    assert response.metadata.prediction_id == request.metadata.prediction_id
    assert response.target_name == request.target_name

    X = load_iris()['data']
    y = model.predict(xgb.DMatrix(X))
    exp_df = pd.DataFrame(y, columns=IrisClassifier.target_names)
    exp_table = df_to_table(exp_df, "prediction_result")

    assert exp_table == response.prediction_result_table


def train_xgboost_model():
    model_dir = "test/pyfunc/"
    BST_FILE = "model_1.bst"

    iris = load_iris()
    y = iris['target']
    X = iris['data']

    dtrain = xgb.DMatrix(X, label=y)
    param = {'max_depth': 6,
             'eta': 0.1,
             'silent': 1,
             'nthread': 4,
             'num_class': 3,
             'objective': 'multi:softprob'
             }
    xgb_model = xgb.train(params=param, dtrain=dtrain)
    model_path = os.path.join(model_dir, BST_FILE)
    xgb_model.save_model(model_path)
    return model_path, xgb_model

def create_simple_forwarder_request() -> upi_pb2.PredictValuesRequest:
    target_name = SimpleForwarder.target_name
    cols = ["id", "name", "vehicle", "previous_vehicle", "rating", "test_time", "row_number"]
    indices = ["row1", "row2"]
    rows = [
        [1, "driver-1", "motorcycle", "suv", 4.0, 90, 0],
        [2, "driver-2", "sedan", "mpv", 3.0, 90, 1]
    ]
    driver_df = pd.DataFrame(columns=cols, data=rows, index=indices)
    driver_table = df_to_table(driver_df, "driver_table")
    variables: List[variable_pb2.Variable] = [
        variable_pb2.Variable(name="customer_id", type=type_pb2.TYPE_INTEGER, integer_value=1111)
    ]
    return upi_pb2.PredictValuesRequest(
        target_name=target_name,
        transformer_input=upi_pb2.TransformerInput(
            tables=[driver_table],
            variables=variables
        )
    )

def simple_forwarder_response() -> upi_pb2.PredictValuesResponse:
    target_name = SimpleForwarder.target_name
    cols = ["customer_id", "name", "rank", "rating", "vehicle", "previous_vehicle"]
    indices = ["row2", "row1"]
    rows = [
        [1111, "driver-2", 2.5, 0.5, 2, 3],
        [1111, "driver-1", -2.5, 0.75, 0, 1]
    ]
    response_df = pd.DataFrame(columns=cols, data=rows, index=indices)
    response_table = df_to_table(response_df, "transformed_driver_table")
    return upi_pb2.PredictValuesResponse(
        target_name=target_name,
        prediction_result_table=response_table
    )

def create_upi_request_from_iris_dataset() -> upi_pb2.PredictValuesRequest:
    target_name = IrisClassifier.target_name
    iris_dataset = load_iris()
    X = iris_dataset['data']
    df = pd.DataFrame(X, columns=iris_dataset.feature_names)

    prediction_table = df_to_table(df, "features")
    return upi_pb2.PredictValuesRequest(
        target_name=target_name,
        prediction_table=prediction_table,
        metadata=upi_pb2.RequestMetadata(
            prediction_id=str(uuid.uuid1())
        )
    )
