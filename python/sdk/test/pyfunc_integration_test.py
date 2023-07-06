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
import warnings
from test.utils import undeploy_all_version
from time import sleep

import joblib
import merlin
import numpy as np
import pytest
import xgboost as xgb
from merlin.model import ModelType, PyFuncModel
from sklearn import svm
from sklearn.datasets import load_iris

warnings.filterwarnings('ignore')

request_json = {
    "instances": [
        [2.8, 1.0, 6.8, 0.4],
        [3.1, 1.4, 4.5, 1.6]
    ]
}


class EnsembleModel(PyFuncModel):
    def initialize(self, artifacts):
        self._model_1 = xgb.Booster(model_file=artifacts["xgb_model"])
        self._model_2 = joblib.load(artifacts["sklearn_model"])

    def infer(self, model_input):
        inputs = np.array(model_input['instances'])
        dmatrix = xgb.DMatrix(model_input['instances'])
        result_1 = self._model_1.predict(dmatrix)
        result_2 = self._model_2.predict_proba(inputs)
        return {"predictions": ((result_1 + result_2) / 2).tolist()}


class EnvVarModel(PyFuncModel):
    def initialize(self, artifacts):
        self.env_var = {}
        self.env_var["workers"] = os.environ.get("WORKERS")
        self.env_var["env_var_1"] = os.environ.get("ENV_VAR_1")
        self.env_var["env_var_2"] = os.environ.get("ENV_VAR_2")

    def infer(self, model_input):
        return self.env_var

@pytest.mark.pyfunc
@pytest.mark.integration
@pytest.mark.dependency()
def test_pyfunc(integration_test_url, project_name, use_google_oauth, requests):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("pyfunc-sample", ModelType.PYFUNC)

    undeploy_all_version()
    with merlin.new_model_version() as v:
        iris = load_iris()
        y = iris['target']
        X = iris['data']
        xgb_path = train_xgboost_model(X, y)
        sklearn_path = train_sklearn_model(X, y)

        v.log_pyfunc_model(model_instance=EnsembleModel(),
                           conda_env="test/pyfunc/env.yaml",
                           code_dir=["test"],
                           artifacts={"xgb_model": xgb_path,
                                      "sklearn_model": sklearn_path})

    endpoint = merlin.deploy(v)

    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert len(resp.json()['predictions']) == len(request_json['instances'])

    merlin.undeploy(v)


@pytest.mark.pyfunc
@pytest.mark.integration
def test_pyfunc_env_vars(integration_test_url, project_name, use_google_oauth, requests):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("pyfunc-env-vars-sample", ModelType.PYFUNC)

    undeploy_all_version()
    with merlin.new_model_version() as v:
        v.log_pyfunc_model(model_instance=EnvVarModel(),
                           conda_env="test/pyfunc/env.yaml",
                           code_dir=["test"],
                           artifacts={})

    env_vars = {"WORKERS": "8", "ENV_VAR_1": "1", "ENV_VAR_2": "2"}
    endpoint = merlin.deploy(v, env_vars=env_vars)
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json() is not None
    assert resp.json()['workers'] == "8"
    assert resp.json()['env_var_1'] == "1"
    assert resp.json()['env_var_2'] == "2"
    assert env_vars.items() <= endpoint.env_vars.items()

    merlin.undeploy(v)


# This implementation of PyFuncModel uses the old infer method (no keyword arguments).
# The keywords arguments for infer() method introduced in Merlin 0.5.2.
class OldInferModel(PyFuncModel):
    def initialize(self, artifacts):
        pass

    def infer(self, model_input):
        return model_input


@pytest.mark.pyfunc
@pytest.mark.integration
def test_pyfunc_old_infer(integration_test_url, project_name, use_google_oauth, requests):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("pyfunc-old-infer-sample", ModelType.PYFUNC)

    undeploy_all_version()
    with merlin.new_model_version() as v:
        v.log_pyfunc_model(model_instance=OldInferModel(),
                           conda_env="test/pyfunc/env.yaml",
                           code_dir=["test"],
                           artifacts={})

    endpoint = merlin.deploy(v)
    resp = requests.post(f"{endpoint.url}", json=request_json)

    assert resp.status_code == 200
    assert resp.json()['instances'] == request_json['instances']

    merlin.undeploy(v)


def train_xgboost_model(X, y):
    model_1_dir = "test/pyfunc/"
    BST_FILE = "model_1.bst"
    dtrain = xgb.DMatrix(X, label=y)
    param = {'max_depth': 6,
             'eta': 0.1,
             'silent': 1,
             'nthread': 4,
             'num_class': 3,
             'objective': 'multi:softprob'
             }
    xgb_model = xgb.train(params=param, dtrain=dtrain)
    model_1_path = os.path.join(model_1_dir, BST_FILE)
    xgb_model.save_model(model_1_path)
    return model_1_path


def train_sklearn_model(X, y):
    model_2_dir = "test/pyfunc/"
    MODEL_FILE = "model_2.joblib"
    model_2_path = os.path.join(model_2_dir, MODEL_FILE)

    clf = svm.SVC(gamma='scale', probability=True)
    clf.fit(X, y)
    joblib.dump(clf, model_2_path)
    return model_2_path
