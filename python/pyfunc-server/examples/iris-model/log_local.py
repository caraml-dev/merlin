import os

import mlflow
from iris_model import IrisModel
from train_models import train_models

dirname = os.path.dirname(__file__)

if __name__ == "__main__":
    xgb_path, sklearn_path = train_models()

    mlflow.pyfunc.log_model(
        artifact_path="model",
        python_model=IrisModel(),
        conda_env=dirname + "/env.yaml",
        code_path=[dirname],
        artifacts={"xgb_model": xgb_path, "sklearn_model": sklearn_path},
    )

    print("PyFunc model have logged locally.")
    print("To run PyFunc Server locally, execute this command:\n")
    print(f"\tpython -m pyfuncserver --model_dir {mlflow.get_artifact_uri()}/model\n")
