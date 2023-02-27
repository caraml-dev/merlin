from iris_model import IrisModel
from train_models import train_models


def test_iris_model():
    xgb_path, sklearn_path = train_models()

    iris_model = IrisModel()
    iris_model.initialize(
        artifacts={"xgb_model": xgb_path, "sklearn_model": sklearn_path}
    )
    response = iris_model.infer({"instances": [[2.8, 1.0, 6.8, 0.4]]})

    assert len(response["predictions"]) == 1
    assert len(response["predictions"][0]) == 3
