import os

import joblib
import xgboost as xgb
from sklearn import svm
from sklearn.datasets import load_iris

dirname = os.path.dirname(__file__)


def train_xgboost_model(X, y):
    model_1_dir = dirname + "/xgboost-model"
    BST_FILE = "model_1.bst"
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
    model_1_path = os.path.join(model_1_dir, BST_FILE)
    xgb_model.save_model(model_1_path)
    return model_1_path


def train_sklearn_model(X, y):
    model_2_dir = dirname + "/sklearn-model"
    MODEL_FILE = "model_2.joblib"
    model_2_path = os.path.join(model_2_dir, MODEL_FILE)

    clf = svm.SVC(gamma="scale", probability=True)
    clf.fit(X, y)
    joblib.dump(clf, model_2_path)
    return model_2_path


def train_models():
    iris = load_iris()
    y = iris["target"]
    X = iris["data"]

    xgb_path = train_xgboost_model(X, y)
    sklearn_path = train_sklearn_model(X, y)

    return xgb_path, sklearn_path
