import os

import pytest


@pytest.fixture
def bq_project() -> str:
    return os.environ.get("INTEGRATION_TEST_BQ_PROJECT")


@pytest.fixture
def bq_dataset() -> str:
    return os.environ.get("INTEGRATION_TEST_BQ_DATASET")
