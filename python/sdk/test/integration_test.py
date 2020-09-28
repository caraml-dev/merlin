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
from time import sleep

import pytest
import requests
import xgboost as xgb
from joblib import dump
from sklearn import svm
from sklearn.datasets import load_iris

import merlin
from merlin.endpoint import Status
from merlin.model import ModelType
from merlin.resource_request import ResourceRequest
from test.utils import undeploy_all_version

request_json = {
    "instances": [
        [2.8, 1.0, 6.8, 0.4],
        [3.1, 1.4, 4.5, 1.6]
    ]
}


@pytest.mark.integration
def test_sklearn(integration_test_url, project_name):
    merlin.set_url(integration_test_url)
    merlin.set_project(project_name)
    merlin.set_model("sklearn-sample", ModelType.SKLEARN)

    model_dir = "test/sklearn-model"
    MODEL_FILE = "model.joblib"

    undeploy_all_version()

    with merlin.new_model_version() as v:
        clf = svm.SVC(gamma='scale')
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
    assert len(resp.json()['predictions']) == len(request_json['instances'])

    merlin.undeploy(v)
    sleep(5)
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 404


@pytest.mark.integration
def test_xgboost(integration_test_url, project_name):
    merlin.set_url(integration_test_url)
    merlin.set_project(project_name)
    merlin.set_model("xgboost-sample", ModelType.XGBOOST)

    model_dir = "test/xgboost-model"
    BST_FILE = "model.bst"

    undeploy_all_version()

    with merlin.new_model_version() as v:
        iris = load_iris()
        y = iris['target']
        X = iris['data']
        dtrain = xgb.DMatrix(X, label=y)
        param = {'max_depth': 6,
                 'eta': 0.1,
                 'silent': 1,
                 'nthread': 4,
                 'num_class': 10,
                 'objective': 'multi:softmax'
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
    assert len(resp.json()['predictions']) == len(request_json['instances'])

    merlin.undeploy(v)
    sleep(5)
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 404

@pytest.mark.integration
def test_mlflow_tracking(integration_test_url, project_name):
    merlin.set_url(integration_test_url)
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
        assert merlin.get_param("iteration") == '5'  # Stringify value which is integer originally
        assert merlin.get_param("random_key") is None

        assert merlin.get_tag("version") == "v1.0"
        assert merlin.get_tag("xxx") is None
        assert merlin.get_tag("team_id") == "1"  # Stringify value which is integer originally

        assert merlin.get_metric("model_loaded") == 10.23
        assert merlin.get_metric("response_time") is None

        assert merlin.list_tag() == {"version": "v1.0", "build": "latest", "team_id": "1"}

        merlin.download_artifact("test/downloaded_artifact")
        artifact_dir = os.listdir('test/downloaded_artifact')
        assert len(artifact_dir) > 0  # not empty directory


@pytest.mark.integration
def test_tensorflow(integration_test_url, project_name):
    merlin.set_url(integration_test_url)
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
            {"sepal_length": 2.8, "sepal_width": 1.0, "petal_length": 6.8,
             "petal_width": 0.4},
            {"sepal_length": 0.1, "sepal_width": 0.5, "petal_length": 1.8,
             "petal_width": 2.4}
        ]
    }
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()['predictions']) == len(request_json['instances'])

    merlin.undeploy(v)
    sleep(5)
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 404


@pytest.mark.integration
def test_pytorch(integration_test_url, project_name):
    merlin.set_url(integration_test_url)
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
    assert len(resp.json()['predictions']) == len(request_json['instances'])

    merlin.undeploy(v)
    sleep(5)
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 404


@pytest.mark.integration
def test_set_traffic(integration_test_url, project_name):
    merlin.set_url(integration_test_url)
    merlin.set_project(project_name)
    merlin.set_model("set-traffic-sample", ModelType.SKLEARN)

    model_dir = "test/sklearn-model"
    MODEL_FILE = "model.joblib"

    undeploy_all_version()

    with merlin.new_model_version() as v:
        clf = svm.SVC(gamma='scale')
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
    assert len(resp.json()['predictions']) == len(request_json['instances'])

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
    assert len(resp.json()['predictions']) == len(request_json['instances'])

    # Try to undeploy serving model version. It must be fail
    with pytest.raises(Exception):
        assert merlin.undeploy(v)

    # Undeploy other running model version endpoints
    undeploy_all_version()


@pytest.mark.integration
def test_serve_traffic(integration_test_url, project_name):
    merlin.set_url(integration_test_url)
    merlin.set_project(project_name)
    merlin.set_model("serve-traffic-sample", ModelType.SKLEARN)

    model_dir = "test/sklearn-model"
    MODEL_FILE = "model.joblib"

    undeploy_all_version()

    with merlin.new_model_version() as v:
        clf = svm.SVC(gamma='scale')
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
    assert len(resp.json()['predictions']) == len(request_json['instances'])

    model_endpoint = merlin.serve_traffic({endpoint: 100})
    sleep(5)
    resp = requests.post(f"{model_endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()['predictions']) == len(request_json['instances'])

    # Try to undeploy serving model version. It must be fail
    with pytest.raises(Exception):
        assert merlin.undeploy(v)

    # Undeploy other running model version endpoints
    undeploy_all_version()


@pytest.mark.integration
def test_stop_serving_traffic(integration_test_url, project_name):
    merlin.set_url(integration_test_url)
    merlin.set_project(project_name)
    merlin.set_model("stop-serving-traffic", ModelType.SKLEARN)

    model_dir = "test/sklearn-model"
    MODEL_FILE = "model.joblib"

    undeploy_all_version()

    with merlin.new_model_version() as v:
        clf = svm.SVC(gamma='scale')
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
    assert len(resp.json()['predictions']) == len(request_json['instances'])

    model_endpoint = merlin.serve_traffic({endpoint: 100})
    sleep(5)
    resp = requests.post(f"{model_endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()['predictions']) == len(request_json['instances'])

    merlin.stop_serving_traffic(model_endpoint.environment_name)

    endpoints = merlin.list_model_endpoints()
    for endpoint in endpoints:
        if endpoint.environment_name == model_endpoint.environment_name:
            assert endpoint.status == Status.TERMINATED

    # Undeploy other running model version endpoints
    undeploy_all_version()


@pytest.mark.integration
def test_multi_env(integration_test_url, project_name):
    merlin.set_url(integration_test_url)
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
        y = iris['target']
        X = iris['data']
        dtrain = xgb.DMatrix(X, label=y)
        param = {'max_depth': 6,
                 'eta': 0.1,
                 'silent': 1,
                 'nthread': 4,
                 'num_class': 10,
                 'objective': 'multi:softmax'
                 }
        xgb_model = xgb.train(params=param, dtrain=dtrain)
        model_file = os.path.join(model_dir, BST_FILE)
        xgb_model.save_model(model_file)

        # Upload the serialized model to MLP
        merlin.log_model(model_dir=model_dir)
        resource_request = ResourceRequest(1, 1, "100m", "200Mi")
        endpoint = merlin.deploy(v, environment_name=default_env.name, resource_request=resource_request)

    sleep(5)
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()['predictions']) == len(request_json['instances'])

    merlin.undeploy(v)
    sleep(5)
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 404

@pytest.mark.integration
def test_resource_request(integration_test_url, project_name):
    merlin.set_url(integration_test_url)
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
        y = iris['target']
        X = iris['data']
        dtrain = xgb.DMatrix(X, label=y)
        param = {'max_depth': 6,
                 'eta': 0.1,
                 'silent': 1,
                 'nthread': 4,
                 'num_class': 10,
                 'objective': 'multi:softmax'
                 }
        xgb_model = xgb.train(params=param, dtrain=dtrain)
        model_file = os.path.join(model_dir, BST_FILE)
        xgb_model.save_model(model_file)

        # Upload the serialized model to MLP
        merlin.log_model(model_dir=model_dir)

        resource_request = ResourceRequest(1, 1, "100m", "200Mi")
        endpoint = merlin.deploy(v, environment_name=default_env.name, resource_request=resource_request)

    sleep(5)
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()['predictions']) == len(request_json['instances'])

    merlin.undeploy(v)
    sleep(5)
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 404
