import os
import uuid
from time import sleep
from typing import List

import grpc
import numpy as np
import pytest
import xgboost as xgb
from caraml.upi.v1 import upi_pb2, value_pb2, upi_pb2_grpc
from sklearn.datasets import load_iris

import merlin
from merlin.deployment_mode import DeploymentMode
from merlin.endpoint import Status
from merlin.model import ModelType, PyFuncModel
from merlin.protocol import Protocol
from test.utils import undeploy_all_version


class IrisClassifier(PyFuncModel):
    feature_names = [
        "sepal-length",
        "sepal-width",
        "petal-length",
        "petal-width"
    ]

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
        return {"predictions": result.tolist()}

    def upiv1_infer(self, request: upi_pb2.PredictValuesRequest,
                    context: grpc.ServicerContext) -> upi_pb2.PredictValuesResponse:
        features = self._get_features_from_request(request)
        predictions = self._predict(features)
        return self._create_response(predictions, request)

    def _create_response(self, predictions: np.ndarray, request: upi_pb2.PredictValuesRequest) -> upi_pb2.PredictValuesResponse:
        result_rows = self._predictions_to_result_rows(predictions, request)
        response_metadata = upi_pb2.ResponseMetadata(prediction_id=request.metadata.prediction_id)
        return upi_pb2.PredictValuesResponse(prediction_result_rows=result_rows, target_name=self.target_name, metadata=response_metadata)

    def _get_features_from_request(self, request: upi_pb2.PredictValuesRequest) -> List[List[float]]:
        features = []
        for row in request.prediction_rows:
            if len(row.model_inputs) != len(self.feature_names):
                raise ValueError(f"invalid features length, got {len(row.model_inputs)} expected: {len(self.feature_names)}")

            feature = []
            for idx, model_input in enumerate(row.model_inputs):
                if model_input.name != self.feature_names[idx]:
                    raise ValueError(f"invalid feature names at index {idx}, got {model_input.name} expected: {self.feature_names[idx]}")
                feature.append(model_input.double_value)
            features.append(feature)

        return features

    def _predictions_to_result_rows(self, predictions: np.ndarray, request: upi_pb2.PredictValuesRequest):
        result_rows = []
        for row_idx, row in enumerate(predictions):
            result_row = []
            for idx, col in enumerate(row):
                val = value_pb2.NamedValue(name=self.target_names[idx], double_value=col)
                result_row.append(val)
            result_rows.append(upi_pb2.PredictionResultRow(row_id=request.prediction_rows[row_idx].row_id, values=result_row))
        return result_rows

    def _predict(self, features: List[List[float]]) -> List[List[float]]:
        features_matrix = xgb.DMatrix(features)
        return self._model.predict(features_matrix)


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


def validate_iris_upi(model, stub):
    request = create_upi_request_from_iris_dataset()
    response = stub.PredictValues(request=request)

    assert response.metadata.prediction_id == request.metadata.prediction_id
    assert response.target_name == request.target_name
    # verify row_id
    for idx, row in enumerate(request.prediction_rows):
        assert response.prediction_result_rows[idx].row_id == row.row_id
    # verify prediction results
    X = load_iris()['data']
    y = model.predict(xgb.DMatrix(X))
    for row_id, row in enumerate(y.tolist()):
        for col_id, val in enumerate(row):
            assert response.prediction_result_rows[row_id].values[col_id].name == IrisClassifier.target_names[col_id]
            assert response.prediction_result_rows[row_id].values[col_id].double_value == val


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


def create_upi_request_from_iris_dataset() -> upi_pb2.PredictValuesRequest:
    target_name = IrisClassifier.target_name
    iris_dataset = load_iris()
    X = iris_dataset['data']
    prediction_rows = []
    for row_id, row in enumerate(X.tolist()):
        model_inputs = []
        for idx, val in enumerate(row):
            named_val = value_pb2.NamedValue(name=IrisClassifier.feature_names[idx], type=value_pb2.NamedValue.TYPE_DOUBLE, double_value=val)
            model_inputs.append(named_val)
        prediction_rows.append(upi_pb2.PredictionRow(model_inputs=model_inputs, row_id=str(row_id)))

    return upi_pb2.PredictValuesRequest(
        target_name=target_name,
        prediction_rows=prediction_rows,
        metadata=upi_pb2.RequestMetadata(
            prediction_id=str(uuid.uuid1())
        )
    )