import pytest
import yaml
from merlin.requirements import _process_conda_env


@pytest.mark.parametrize(
    "input,output,additional_merlin_reqs",
    [
        # Test using file
        (
            "test/requirements/empty_in.yaml",
            "test/requirements/empty_out.yaml",
            ["merlin-pyfunc-server==0.41.0"],
        ),
        (
            "test/requirements/no_pip_reqs_in.yaml",
            "test/requirements/no_pip_reqs_out.yaml",
            ["merlin-pyfunc-server==0.41.0"],
        ),
        (
            "test/requirements/pyfunc_server_req_with_version_in.yaml",
            "test/requirements/pyfunc_server_req_with_version_out.yaml",
            ["merlin-pyfunc-server==0.41.0"],
        ),
        (
            "test/requirements/pyfunc_server_req_without_version_in.yaml",
            "test/requirements/pyfunc_server_req_without_version_out.yaml",
            ["merlin-pyfunc-server==0.41.0"],
        ),
        # Test using dictionary
        # Empty conda env
        (
            {},
            {
                "dependencies": [
                    "python=3.10.*",
                    {"pip": ["merlin-pyfunc-server==0.41.0"]},
                ]
            },
            ["merlin-pyfunc-server==0.41.0"],
        ),
        # Using old python version, no pip reqs
        (
            {
                "dependencies": [
                    "python=3.7.*",
                ]
            },
            {
                "dependencies": [
                    "python=3.10.*",
                    {"pip": ["merlin-pyfunc-server==0.41.0"]},
                ]
            },
            ["merlin-pyfunc-server==0.41.0"],
        ),
        # Using old python version, with empty pip reqs
        (
            {
                "dependencies": [
                    "python=3.7.*",
                    {"pip": []},
                ]
            },
            {
                "dependencies": [
                    "python=3.10.*",
                    {"pip": ["merlin-pyfunc-server==0.41.0"]},
                ]
            },
            ["merlin-pyfunc-server==0.41.0"],
        ),
        # pip reqs has merlin-pyfunc-server without version
        (
            {
                "dependencies": [
                    "python=3.10.*",
                    {"pip": ["merlin-pyfunc-server"]},
                ]
            },
            {
                "dependencies": [
                    "python=3.10.*",
                    {"pip": ["merlin-pyfunc-server"]},
                ]
            },
            ["merlin-pyfunc-server==0.41.0"],
        ),
        # pip reqs has merlin-pyfunc-server with older version, should not be updated
        (
            {
                "dependencies": [
                    "python=3.10.*",
                    {"pip": ["merlin-pyfunc-server==0.20.0"]},
                ]
            },
            {
                "dependencies": [
                    "python=3.10.*",
                    {"pip": ["merlin-pyfunc-server==0.20.0"]},
                ]
            },
            ["merlin-pyfunc-server==0.41.0"],
        ),
        # pip reqs has merlin-pyfunc-server with newer version, should not be updated
        (
            {
                "dependencies": [
                    "python=3.10.*",
                    {"pip": ["merlin-pyfunc-server==0.50.0"]},
                ]
            },
            {
                "dependencies": [
                    "python=3.10.*",
                    {"pip": ["merlin-pyfunc-server==0.50.0"]},
                ]
            },
            ["merlin-pyfunc-server==0.41.0"],
        ),
    ],
)
def test_process_conda_env(input, output, additional_merlin_reqs):
    default_python_version = "3.10.*"

    actual_conda_env = _process_conda_env(
        conda_env=input,
        python_version=default_python_version,
        additional_merlin_reqs=additional_merlin_reqs,
    )

    expected_conda_env = output
    if isinstance(output, str):
        with open(output, "r") as f:
            expected_conda_env = yaml.safe_load(f)

    print(actual_conda_env)
    assert actual_conda_env == expected_conda_env
