from test.pyfunc_integration_test import EnvVarModel

import pytest

import merlin
from merlin.model import ModelType


@pytest.mark.build_image
@pytest.mark.integration
@pytest.mark.dependency()
def test_model_version_with_labels(
    integration_test_url, project_name, use_google_oauth
):
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
