import os

import merlin
from iris_model import IrisModel
from merlin.model import ModelType
from train_models import train_models

MERLIN_URL = "console.d.ai.golabs.io/api/merlin"
PROJECT_NAME = "sample"

dirname = os.path.dirname(__file__)

if __name__ == "__main__":
    merlin.set_url(MERLIN_URL)
    merlin.set_project(PROJECT_NAME)
    merlin.set_model("iris-model", ModelType.PYFUNC)

    with merlin.new_model_version() as v:
        xgb_path, sklearn_path = train_models()

        v.log_pyfunc_model(
            model_instance=IrisModel(),
            conda_env=dirname + "/env.yaml",
            code_dir=[dirname],
            artifacts={"xgb_model": xgb_path, "sklearn_model": sklearn_path},
        )

    merlin.deploy(v)
