from test.batch_integration_test import IrisClassifier
from test.pyfunc_integration_test import EnvVarModel

import pytest

import merlin
from merlin.model import ModelType


@pytest.mark.build_image
@pytest.mark.integration
@pytest.mark.dependency()
def test_build_image_pyfunc(integration_test_url, project_name, use_google_oauth):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("build-image", ModelType.PYFUNC)

    with merlin.new_model_version() as v:
        v.log_pyfunc_model(
            model_instance=EnvVarModel(),
            conda_env="test/pyfunc/env.yaml",
            code_dir=["test"],
            artifacts={},
        )

    image = merlin.build_image(v)
    assert image.existed == True


@pytest.mark.build_image
@pytest.mark.integration
@pytest.mark.dependency()
def test_build_image_batch(integration_test_url, project_name, use_google_oauth):
    merlin.set_url(integration_test_url, use_google_oauth=use_google_oauth)
    merlin.set_project(project_name)
    merlin.set_model("build-image-batch", ModelType.PYFUNC_V2)

    with merlin.new_model_version() as v:
        v.log_pyfunc_model(
            model_instance=IrisClassifier(),
            conda_env="test/batch/model/env.yaml",
            code_dir=["test"],
            artifacts={"model_path": "test/batch/model/model.joblib"},
        )

    image = merlin.build_image(v)
    assert image.existed == True
