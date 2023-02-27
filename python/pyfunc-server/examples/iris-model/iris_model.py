import joblib
import numpy as np
import xgboost as xgb
from merlin.model import PyFuncModel


class IrisModel(PyFuncModel):
    def initialize(self, artifacts: dict):
        self._model_1 = xgb.Booster(model_file=artifacts["xgb_model"])
        self._model_2 = joblib.load(artifacts["sklearn_model"])

    def infer(self, request: dict, **kwargs) -> dict:
        inputs = np.array(request["instances"])
        dmatrix = xgb.DMatrix(request["instances"])
        result_1 = self._model_1.predict(dmatrix)
        result_2 = self._model_2.predict_proba(inputs)
        return {"predictions": ((result_1 + result_2) / 2).tolist()}
