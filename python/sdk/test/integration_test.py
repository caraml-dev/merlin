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
import json
import os
from time import sleep

import pytest
import requests
import pandas as pd
import xgboost as xgb
from joblib import dump
from sklearn import svm
from sklearn.datasets import load_iris

import merlin
from merlin.endpoint import Status
from merlin.model import ModelType
from merlin.resource_request import ResourceRequest
from merlin.transformer import Transformer, StandardTransformer
from merlin.logger import Logger, LoggerConfig, LoggerMode
from test.utils import undeploy_all_version
from test.feast_model import EchoModel

request_json = {"instances": [[2.8, 1.0, 6.8, 0.4], [3.1, 1.4, 4.5, 1.6]]}


@pytest.mark.integration
@pytest.mark.dependency()
def test_model_version_with_labels(
    integration_test_url, project_name, use_google_oauth
):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("sklearn-labels", ModelType.SKLEARN)

    model_dir = "test/sklearn-model"
    MODEL_FILE = "model.joblib"

    undeploy_all_version()

    with merlin.new_model_version(labels={"model": "T-800"}) as v:
        clf = svm.SVC(gamma="scale")
        iris = load_iris()
        X, y = iris.data, iris.target
        clf.fit(X, y)
        dump(clf, os.path.join(model_dir, MODEL_FILE))

        # Upload the serialized model to MLP
        merlin.log_model(model_dir=model_dir)
        assert len(v.labels) == 1
        assert v.labels["model"] == "T-800"

    merlin_active_model = merlin.active_model()
    all_versions = merlin_active_model.list_version(labels={"model": ["T-800"]})
    for version in all_versions:
        assert version.labels["model"] == "T-800"

    should_not_exist_versions = merlin_active_model.list_version(
        labels={"model": ["T-1000"]}
    )
    assert len(should_not_exist_versions) == 0


@pytest.mark.integration
@pytest.mark.dependency()
def test_sklearn(integration_test_url, project_name, use_google_oauth):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("sklearn-sample", ModelType.SKLEARN)

    model_dir = "test/sklearn-model"
    MODEL_FILE = "model.joblib"

    undeploy_all_version()

    with merlin.new_model_version() as v:
        clf = svm.SVC(gamma="scale")
        iris = load_iris()
        X, y = iris.data, iris.target
        clf.fit(X, y)
        dump(clf, os.path.join(model_dir, MODEL_FILE))

        # Upload the serialized model to MLP
        merlin.log_model(model_dir=model_dir)

    endpoint = merlin.deploy(v)
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()["predictions"]) == len(request_json["instances"])

    merlin.undeploy(v)


@pytest.mark.integration
@pytest.mark.dependency()
def test_xgboost(integration_test_url, project_name, use_google_oauth):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("xgboost-sample", ModelType.XGBOOST)

    model_dir = "test/xgboost-model"
    BST_FILE = "model.bst"

    undeploy_all_version()

    with merlin.new_model_version() as v:
        iris = load_iris()
        y = iris["target"]
        X = iris["data"]
        dtrain = xgb.DMatrix(X, label=y)
        param = {
            "max_depth": 6,
            "eta": 0.1,
            "silent": 1,
            "nthread": 4,
            "num_class": 10,
            "objective": "multi:softmax",
        }
        xgb_model = xgb.train(params=param, dtrain=dtrain)
        model_file = os.path.join(model_dir, BST_FILE)
        xgb_model.save_model(model_file)

        # Upload the serialized model to MLP
        merlin.log_model(model_dir=model_dir)

    resource_request = ResourceRequest(1, 1, "100m", "200Mi")
    endpoint = merlin.deploy(v, resource_request=resource_request)
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()["predictions"]) == len(request_json["instances"])

    merlin.undeploy(v)


@pytest.mark.integration
def test_mlflow_tracking(integration_test_url, project_name, use_google_oauth):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("mlflow-tracking", ModelType.PYTORCH)

    model_dir = "test/pytorch-model"

    undeploy_all_version()

    with merlin.new_model_version() as v:
        merlin.log_pytorch_model(model_dir=model_dir)
        merlin.log_param("model_type", "pytorch")
        merlin.log_param("iteration", 5)

        merlin.set_tag("version", "v1.0")
        merlin.set_tag("build", "latest")
        merlin.set_tag("team_id", 1)

        merlin.log_metric("model_loaded", 10.23)

        assert merlin.get_param("model_type") == "pytorch"
        # Stringify value which is integer originally
        assert merlin.get_param("iteration") == "5"
        assert merlin.get_param("random_key") is None

        assert merlin.get_tag("version") == "v1.0"
        assert merlin.get_tag("xxx") is None
        # Stringify value which is integer originally
        assert merlin.get_tag("team_id") == "1"

        assert merlin.get_metric("model_loaded") == 10.23
        assert merlin.get_metric("response_time") is None

        assert merlin.list_tag() == {
            "version": "v1.0",
            "build": "latest",
            "team_id": "1",
        }

        # TODO: Support downloading artifacts from S3 or S3-compatible alternative (such as MinIO)
        # merlin.download_artifact("test/downloaded_artifact")
        # artifact_dir = os.listdir('test/downloaded_artifact')
        # assert len(artifact_dir) > 0  # not empty directory


@pytest.mark.integration
@pytest.mark.dependency()
def test_tensorflow(integration_test_url, project_name, use_google_oauth):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("tensorflow-sample", ModelType.TENSORFLOW)

    model_dir = "test/tensorflow-model"

    undeploy_all_version()

    with merlin.new_model_version() as v:
        merlin.log_model(model_dir=model_dir)

    endpoint = merlin.deploy(v)
    request_json = {
        "signature_name": "predict",
        "instances": [
            {
                "sepal_length": 2.8,
                "sepal_width": 1.0,
                "petal_length": 6.8,
                "petal_width": 0.4,
            },
            {
                "sepal_length": 0.1,
                "sepal_width": 0.5,
                "petal_length": 1.8,
                "petal_width": 2.4,
            },
        ],
    }
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()["predictions"]) == len(request_json["instances"])

    merlin.undeploy(v)


@pytest.mark.integration
@pytest.mark.dependency()
def test_pytorch(integration_test_url, project_name, use_google_oauth):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("pytorch-sample", ModelType.PYTORCH)

    model_dir = "test/pytorch-model"

    undeploy_all_version()

    with merlin.new_model_version() as v:
        merlin.log_pytorch_model(model_dir=model_dir)
        endpoint = merlin.deploy()

    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()["predictions"]) == len(request_json["instances"])

    merlin.undeploy(v)


@pytest.mark.integration
def test_set_traffic(integration_test_url, project_name, use_google_oauth):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("set-traffic-sample", ModelType.SKLEARN)

    model_dir = "test/sklearn-model"
    MODEL_FILE = "model.joblib"

    undeploy_all_version()

    with merlin.new_model_version() as v:
        clf = svm.SVC(gamma="scale")
        iris = load_iris()
        X, y = iris.data, iris.target
        clf.fit(X, y)
        dump(clf, os.path.join(model_dir, MODEL_FILE))

        # Upload the serialized model to MLP
        merlin.log_model(model_dir=model_dir)
        endpoint = merlin.deploy(v)

    sleep(5)
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()["predictions"]) == len(request_json["instances"])

    # Undeploy deployed model version
    merlin.undeploy(v)
    sleep(5)

    # Redeploy and set traffic
    merlin.deploy(v)

    endpoint = merlin.set_traffic({v: 100})
    sleep(5)
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()["predictions"]) == len(request_json["instances"])

    # Try to undeploy serving model version. It must be fail
    with pytest.raises(Exception):
        assert merlin.undeploy(v)

    # Undeploy other running model version endpoints
    undeploy_all_version()


@pytest.mark.integration
def test_serve_traffic(integration_test_url, project_name, use_google_oauth):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("serve-traffic-sample", ModelType.SKLEARN)

    model_dir = "test/sklearn-model"
    MODEL_FILE = "model.joblib"

    undeploy_all_version()

    with merlin.new_model_version() as v:
        clf = svm.SVC(gamma="scale")
        iris = load_iris()
        X, y = iris.data, iris.target
        clf.fit(X, y)
        dump(clf, os.path.join(model_dir, MODEL_FILE))

        # Upload the serialized model to MLP
        merlin.log_model(model_dir=model_dir)
        endpoint = merlin.deploy(v)

    sleep(5)
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()["predictions"]) == len(request_json["instances"])

    model_endpoint = merlin.serve_traffic({endpoint: 100})
    sleep(5)
    resp = requests.post(f"{model_endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()["predictions"]) == len(request_json["instances"])

    # Try to undeploy serving model version. It must be fail
    with pytest.raises(Exception):
        assert merlin.undeploy(v)

    # Undeploy other running model version endpoints
    undeploy_all_version()


@pytest.mark.integration
def test_stop_serving_traffic(integration_test_url, project_name, use_google_oauth):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("stop-serving-traffic", ModelType.SKLEARN)

    model_dir = "test/sklearn-model"
    MODEL_FILE = "model.joblib"

    undeploy_all_version()

    with merlin.new_model_version() as v:
        clf = svm.SVC(gamma="scale")
        iris = load_iris()
        X, y = iris.data, iris.target
        clf.fit(X, y)
        dump(clf, os.path.join(model_dir, MODEL_FILE))

        # Upload the serialized model to MLP
        merlin.log_model(model_dir=model_dir)
        endpoint = merlin.deploy(v)

    sleep(5)
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()["predictions"]) == len(request_json["instances"])

    model_endpoint = merlin.serve_traffic({endpoint: 100})
    sleep(5)
    resp = requests.post(f"{model_endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()["predictions"]) == len(request_json["instances"])

    merlin.stop_serving_traffic(model_endpoint.environment_name)

    endpoints = merlin.list_model_endpoints()
    for endpoint in endpoints:
        if endpoint.environment_name == model_endpoint.environment_name:
            assert endpoint.status == Status.TERMINATED

    # Undeploy other running model version endpoints
    undeploy_all_version()


@pytest.mark.integration
def test_multi_env(integration_test_url, project_name, use_google_oauth):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("multi-env", ModelType.XGBOOST)

    model_dir = "test/xgboost-model"
    BST_FILE = "model.bst"

    envs = merlin.list_environment()
    assert len(envs) >= 1

    default_env = merlin.get_default_environment()
    assert default_env is not None

    undeploy_all_version()

    with merlin.new_model_version() as v:
        iris = load_iris()
        y = iris["target"]
        X = iris["data"]
        dtrain = xgb.DMatrix(X, label=y)
        param = {
            "max_depth": 6,
            "eta": 0.1,
            "silent": 1,
            "nthread": 4,
            "num_class": 10,
            "objective": "multi:softmax",
        }
        xgb_model = xgb.train(params=param, dtrain=dtrain)
        model_file = os.path.join(model_dir, BST_FILE)
        xgb_model.save_model(model_file)

        # Upload the serialized model to MLP
        merlin.log_model(model_dir=model_dir)
        resource_request = ResourceRequest(1, 1, "100m", "200Mi")
        endpoint = merlin.deploy(
            v, environment_name=default_env.name, resource_request=resource_request
        )

    sleep(5)
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()["predictions"]) == len(request_json["instances"])

    merlin.undeploy(v)


@pytest.mark.integration
def test_resource_request(integration_test_url, project_name, use_google_oauth):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("resource-request", ModelType.XGBOOST)

    model_dir = "test/xgboost-model"
    BST_FILE = "model.bst"

    envs = merlin.list_environment()
    assert len(envs) >= 1

    default_env = merlin.get_default_environment()
    assert default_env is not None

    undeploy_all_version()
    with merlin.new_model_version() as v:
        iris = load_iris()
        y = iris["target"]
        X = iris["data"]
        dtrain = xgb.DMatrix(X, label=y)
        param = {
            "max_depth": 6,
            "eta": 0.1,
            "silent": 1,
            "nthread": 4,
            "num_class": 10,
            "objective": "multi:softmax",
        }
        xgb_model = xgb.train(params=param, dtrain=dtrain)
        model_file = os.path.join(model_dir, BST_FILE)
        xgb_model.save_model(model_file)

        # Upload the serialized model to MLP
        merlin.log_model(model_dir=model_dir)

        resource_request = ResourceRequest(1, 1, "100m", "200Mi")
        endpoint = merlin.deploy(
            v, environment_name=default_env.name, resource_request=resource_request
        )

    sleep(5)
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()["predictions"]) == len(request_json["instances"])

    merlin.undeploy(v)


@pytest.mark.integration
def test_pytorch_logger(integration_test_url, project_name, use_google_oauth):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("pytorch-logger", ModelType.PYTORCH)

    model_dir = "test/pytorch-model"

    undeploy_all_version()

    logger = Logger(model=LoggerConfig(enabled=True, mode=LoggerMode.REQUEST))
    with merlin.new_model_version() as v:
        merlin.log_pytorch_model(model_dir=model_dir)
        endpoint = merlin.deploy(logger=logger)

    model_config = endpoint.logger.model
    assert model_config is not None
    assert model_config.enabled
    assert model_config.mode == LoggerMode.REQUEST

    transformer_config = endpoint.logger.transformer
    assert transformer_config is None

    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()["predictions"]) == len(request_json["instances"])

    undeploy_all_version()


@pytest.mark.integration
def test_trasformer_pytorch_logger(
    integration_test_url, project_name, use_google_oauth
):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("transformer-logger", ModelType.PYTORCH)

    model_dir = "test/transformer"

    undeploy_all_version()

    resource_request = ResourceRequest(1, 1, "100m", "200Mi")
    transformer = Transformer(
        "gcr.io/kubeflow-ci/kfserving/image-transformer:latest",
        resource_request=resource_request,
    )

    logger = Logger(
        model=LoggerConfig(enabled=True, mode=LoggerMode.ALL),
        transformer=LoggerConfig(enabled=True, mode=LoggerMode.ALL),
    )
    with merlin.new_model_version() as v:
        merlin.log_pytorch_model(model_dir=model_dir)
        endpoint = merlin.deploy(transformer=transformer, logger=logger)

    assert endpoint.logger is not None

    model_config = endpoint.logger.model
    assert model_config is not None
    assert model_config.enabled
    assert model_config.mode == LoggerMode.ALL

    transformer_config = endpoint.logger.transformer
    assert transformer_config is not None
    assert transformer_config.enabled
    assert transformer_config.mode == LoggerMode.ALL

    with open(os.path.join("test/transformer", "input.json"), "r") as f:
        req = json.load(f)

    sleep(5)
    resp = requests.post(f"{endpoint.url}", json=req)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()["predictions"]) == len(req["instances"])

    model_endpoint = merlin.serve_traffic({endpoint: 100})
    sleep(5)
    resp = requests.post(f"{model_endpoint.url}", json=req)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()["predictions"]) == len(req["instances"])

    # Try to undeploy serving model version. It must be fail
    with pytest.raises(Exception):
        assert merlin.undeploy(v)

    # Undeploy other running model version endpoints
    undeploy_all_version()


@pytest.mark.integration
def test_transformer_pytorch(integration_test_url, project_name, use_google_oauth):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("transformer-pytorch", ModelType.PYTORCH)

    model_dir = "test/transformer"

    undeploy_all_version()

    resource_request = ResourceRequest(1, 1, "100m", "200Mi")
    transformer = Transformer(
        "gcr.io/kubeflow-ci/kfserving/image-transformer:latest",
        resource_request=resource_request,
    )
    print("transformer test", transformer)

    with merlin.new_model_version() as v:
        merlin.log_pytorch_model(model_dir=model_dir)
        endpoint = merlin.deploy(transformer=transformer)

    with open(os.path.join("test/transformer", "input.json"), "r") as f:
        req = json.load(f)

    sleep(5)
    resp = requests.post(f"{endpoint.url}", json=req)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()["predictions"]) == len(req["instances"])

    model_endpoint = merlin.serve_traffic({endpoint: 100})
    sleep(5)
    resp = requests.post(f"{model_endpoint.url}", json=req)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()["predictions"]) == len(req["instances"])

    # Try to undeploy serving model version. It must be fail
    with pytest.raises(Exception):
        assert merlin.undeploy(v)

    # Undeploy other running model version endpoints
    undeploy_all_version()


@pytest.mark.feast
@pytest.mark.integration
def test_feast_enricher(integration_test_url, project_name, use_google_oauth):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("feast-enricher", ModelType.PYFUNC)

    undeploy_all_version()
    with merlin.new_model_version() as v:
        v.log_pyfunc_model(
            model_instance=EchoModel(),
            conda_env="test/pyfunc/env.yaml",
            code_dir=["test"],
            artifacts={},
        )

    transformer_config_path = os.path.join("test/transformer", "feast_enricher.yaml")
    transformer = StandardTransformer(config_file=transformer_config_path, enabled=True)

    request_json = {"driver_id": "1000"}
    endpoint = merlin.deploy(v, transformer=transformer)
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    feast_features = resp.json()["feast_features"]
    assert feast_features is not None
    assert pd.DataFrame(feast_features) is not None

    merlin.undeploy(v)


@pytest.mark.integration
def test_standard_transformer_without_feast(
    integration_test_url, project_name, use_google_oauth
):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("std-transformer", ModelType.PYFUNC)

    undeploy_all_version()
    with merlin.new_model_version() as v:
        v.log_pyfunc_model(
            model_instance=EchoModel(),
            conda_env="test/pyfunc/env.yaml",
            code_dir=["test"],
            artifacts={},
        )

    transformer_config_path = os.path.join(
        "test/transformer", "standard_transformer_no_feast.yaml"
    )
    transformer = StandardTransformer(config_file=transformer_config_path, enabled=True)

    endpoint = merlin.deploy(v, transformer=transformer)
    request_json = {
        "drivers": [
            {"id": 1, "name": "driver-1", "vehicle": "motorcycle", "previous_vehicle": "suv","rating": 4}, 
            {"id": 2, "name": "driver-2", "vehicle": "sedan", "previous_vehicle": "mpv", "rating": 3}],
        "customer": {"id": 1111},
    }
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    exp_resp = {
        "instances": {
            "columns": ["customer_id", "name", "rank", "rating", "vehicle", "previous_vehicle"],
            "data": [
                [1111, "driver-2", 2.5, 0.5, 2, 3], 
                [1111, "driver-1", -2.5, 0.75, 0, 1]],
        }
    }

    assert resp.json()["instances"] == exp_resp["instances"]
    merlin.undeploy(v)


@pytest.mark.feast
@pytest.mark.integration
def test_standard_transformer_with_feast(
    integration_test_url, project_name, use_google_oauth
):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("std-transformer-feast", ModelType.PYFUNC)

    undeploy_all_version()
    with merlin.new_model_version() as v:
        v.log_pyfunc_model(
            model_instance=EchoModel(),
            conda_env="test/pyfunc/env.yaml",
            code_dir=["test"],
            artifacts={},
        )

    transformer_config_path = os.path.join(
        "test/transformer", "standard_transformer_with_feast.yaml"
    )
    transformer = StandardTransformer(config_file=transformer_config_path, enabled=True)

    endpoint = merlin.deploy(v, transformer=transformer)
    request_json = {
        "drivers": [
            {"id": "1234", "name": "driver-1"},
            {"id": "5678", "name": "driver-2"},
        ],
        "customer": {"id": 1111},
    }
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    exp_resp = {
        "instances": {
            "columns": [
                "rank",
                "driver_id",
                "customer_id",
                "merlin_test_driver_features:test_int32",
                "merlin_test_driver_features:test_float",
            ],
            "data": [[0, "1234", 1111, -1, 0], [1, "5678", 1111, -1, 0]],
        }
    }

    assert resp.json()["instances"] == exp_resp["instances"]
    merlin.undeploy(v)


@pytest.mark.feast
@pytest.mark.integration
def test_standard_transformer_with_multiple_feast(
    integration_test_url,
    project_name,
    use_google_oauth,
    feast_serving_redis_url,
    feast_serving_bigtable_url,
):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("std-transformer-feasts", ModelType.PYFUNC)

    undeploy_all_version()
    with merlin.new_model_version() as v:
        v.log_pyfunc_model(
            model_instance=EchoModel(),
            conda_env="test/pyfunc/env.yaml",
            code_dir=["test"],
            artifacts={},
        )

    config_template_file_path = os.path.join(
        "test/transformer", "standard_transformer_multiple_feast.yaml.tmpl"
    )
    config_file_path = os.path.join(
        "test/transformer", "standard_transformer_multiple_feast.yaml"
    )

    from string import Template

    config_template_file = open(config_template_file_path, "rt")
    t = Template(config_template_file.read())
    rendered_config = t.substitute(
        {
            "feast_serving_redis_url": feast_serving_redis_url,
            "feast_serving_bigtable_url": feast_serving_bigtable_url,
        }
    )
    config_file = open(config_file_path, "wt")
    config_file.write(rendered_config)
    config_file.close()

    transformer = StandardTransformer(config_file=config_file_path, enabled=True)

    endpoint = merlin.deploy(v, transformer=transformer)
    request_json = {
        "drivers": [
            {"id": "driver_1", "name": "driver-1"},
            {"id": "driver_2", "name": "driver-2"},
        ],
        "customer": {"id": 1111},
    }
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    exp_resp = {
        "instances": {
            "columns": [
            "rank",
            "driver_id",
            "customer_id",
            "merlin_test_redis_driver_features:completion_rate",
            "merlin_test_redis_driver_features:cancellation_rate",
            "merlin_test_bt_driver_features:rating"
            ],
            "data": [
            [
                0,
                "driver_1",
                1111,
                0.85,
                0.15,
                4.2
            ],
            [
                1,
                "driver_2",
                1111,
                0.6,
                0.4,
                4.2
            ]
            ]
        }
    }


    assert resp.json()["instances"] == exp_resp["instances"]
    merlin.undeploy(v)

@pytest.mark.feast
@pytest.mark.integration
def test_standard_transformer_with_multiple_feast_with_source(
    integration_test_url,
    project_name,
    use_google_oauth,
    feast_serving_bigtable_url,
):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("std-trf-feasts-source", ModelType.PYFUNC)

    undeploy_all_version()
    with merlin.new_model_version() as v:
        v.log_pyfunc_model(
            model_instance=EchoModel(),
            conda_env="test/pyfunc/env.yaml",
            code_dir=["test"],
            artifacts={},
        )

    config_template_file_path = os.path.join(
        "test/transformer", "standard_transformer_feast_with_source.yaml.tmpl"
    )
    config_file_path = os.path.join(
        "test/transformer", "standard_transformer_multiple_feast.yaml"
    )

    from string import Template

    config_template_file = open(config_template_file_path, "rt")
    t = Template(config_template_file.read())
    rendered_config = t.substitute(
        {
            "feast_serving_bigtable_url": feast_serving_bigtable_url,
        }
    )
    config_file = open(config_file_path, "wt")
    config_file.write(rendered_config)
    config_file.close()

    transformer = StandardTransformer(config_file=config_file_path, enabled=True)

    endpoint = merlin.deploy(v, transformer=transformer)
    request_json = {
        "drivers": [
            {"id": "driver_1", "name": "driver-1"},
            {"id": "driver_2", "name": "driver-2"},
        ],
        "customer": {"id": 1111},
    }
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    exp_resp = {
        "instances": {
            "columns": [
            "rank",
            "driver_id",
            "customer_id",
            "merlin_test_redis_driver_features:completion_rate",
            "merlin_test_redis_driver_features:cancellation_rate",
            "merlin_test_bt_driver_features:rating"
            ],
            "data": [
            [
                0,
                "driver_1",
                1111,
                0.85,
                0.15,
                4.2
            ],
            [
                1,
                "driver_2",
                1111,
                0.6,
                0.4,
                4.2
            ]
            ]
        }
    }


    assert resp.json()["instances"] == exp_resp["instances"]
    merlin.undeploy(v)


@pytest.mark.integration
def test_custom_model_without_artifact(
    integration_test_url, project_name, use_google_oauth
):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("custom-wo-artifact", ModelType.CUSTOM)

    undeploy_all_version()

    resource_request = ResourceRequest(1, 1, "1", "1Gi")

    with merlin.new_model_version() as v:
        v.log_custom_model(image="ghcr.io/tiopramayudi/custom-predictor:v0.2")

    endpoint = merlin.deploy(v, resource_request=resource_request)
    with open(os.path.join("test/custom-model", "input.json"), "r") as f:
        req = json.load(f)

    sleep(5)
    resp = requests.post(f"{endpoint.url}", json=req)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert resp.json()["predictions"] is not None

    model_endpoint = merlin.serve_traffic({endpoint: 100})
    sleep(5)
    resp = requests.post(f"{model_endpoint.url}", json=req)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert resp.json()["predictions"] is not None

    # Try to undeploy serving model version. It must be fail
    with pytest.raises(Exception):
        assert merlin.undeploy(v)

    # Undeploy other running model version endpoints
    undeploy_all_version()


@pytest.mark.integration
def test_custom_model_with_artifact(
    integration_test_url, project_name, use_google_oauth
):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("custom-w-artifact", ModelType.CUSTOM)
    undeploy_all_version()

    resource_request = ResourceRequest(1, 1, "1", "1Gi")
    model_dir = "test/custom-model"
    BST_FILE = "model.bst"

    iris = load_iris()
    y = iris["target"]
    X = iris["data"]
    dtrain = xgb.DMatrix(X, label=y)
    param = {
        "max_depth": 6,
        "eta": 0.1,
        "silent": 1,
        "nthread": 4,
        "num_class": 10,
        "objective": "multi:softmax",
    }
    xgb_model = xgb.train(params=param, dtrain=dtrain)
    model_file = os.path.join((model_dir), BST_FILE)
    xgb_model.save_model(model_file)

    with merlin.new_model_version() as v:
        v.log_custom_model(
            image="ghcr.io/tiopramayudi/custom-predictor-go:v0.2", model_dir=model_dir
        )

    endpoint = merlin.deploy(
        v, resource_request=resource_request, env_vars={"MODEL_FILE_NAME": BST_FILE}
    )

    sleep(5)
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert resp.json()["predictions"] is not None

    model_endpoint = merlin.serve_traffic({endpoint: 100})
    sleep(5)
    resp = requests.post(f"{model_endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert resp.json()["predictions"] is not None

    # Try to undeploy serving model version. It must be fail
    with pytest.raises(Exception):
        assert merlin.undeploy(v)

    # Undeploy other running model version endpoints
    undeploy_all_version()
