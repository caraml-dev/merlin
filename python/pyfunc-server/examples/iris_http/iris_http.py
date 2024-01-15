import os

import joblib
import merlin
import numpy as np
import xgboost as xgb
from joblib import dump
from merlin.model import PyFuncModel
from sklearn import svm
from sklearn.datasets import load_iris

XGB_PATH = os.path.join("models/", "model_1.bst")
SKLEARN_PATH = os.path.join("models/", "model_2.joblib")


class IrisModel(PyFuncModel):
    def initialize(self, artifacts):
        self._model_1 = xgb.Booster(model_file=artifacts["xgb_model"])
        self._model_2 = joblib.load(artifacts["sklearn_model"])

    def infer(self, model_input):
        inputs = np.array(model_input["instances"])
        dmatrix = xgb.DMatrix(model_input["instances"])
        result_1 = self._model_1.predict(dmatrix)
        result_2 = self._model_2.predict_proba(inputs)
        return {"predictions": ((result_1 + result_2) / 2).tolist()}


def train_models():
    iris = load_iris()
    y = iris["target"]
    X = iris["data"]

    # train xgboost model
    dtrain = xgb.DMatrix(X, label=y)
    param = {
        "max_depth": 6,
        "eta": 0.1,
        "silent": 1,
        "nthread": 4,
        "num_class": 3,
        "objective": "multi:softprob",
    }
    xgb_model = xgb.train(params=param, dtrain=dtrain)
    xgb_model.save_model(XGB_PATH)

    # train sklearn model
    clf = svm.SVC(gamma="scale", probability=True)
    clf.fit(X, y)
    dump(clf, SKLEARN_PATH)


if __name__ == "__main__":
    train_models()

    # test pyfunc model locally
    iris_model = IrisModel()
    iris_model.initialize(
        artifacts={
            "xgb_model": XGB_PATH,
            "sklearn_model": SKLEARN_PATH,
        }
    )
    pred = iris_model.infer({"instances": [[2.8, 1.0, 6.8, 0.4], [3.1, 1.4, 4.5, 1.6]]})
    print(pred)

    # run pyfunc model locally
    merlin.run_pyfunc_model(
        model_instance=IrisModel(),
        conda_env="env.yaml",
        artifacts={
            "xgb_model": XGB_PATH,
            "sklearn_model": SKLEARN_PATH,
        },
    )
