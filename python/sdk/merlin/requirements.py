from typing import Dict, List, Optional

import yaml
from packaging.requirements import InvalidRequirement, Requirement


def get_default_merlin_requirements():
    return [
        "merlin-batch-job==0.41.0",
        "merlin-pyfunc-server==0.41.0",
    ]


def _process_conda_env(
    conda_env: Dict = None,
    python_version: str = "3.10.*",
    additional_merlin_reqs: Optional[List[str]] = None,
):
    """
    Processes `conda_env` passed to `merlin.log_pyfunc_model` and `merlin.log_model`.
    """

    if isinstance(conda_env, str):
        with open(conda_env, "r") as f:
            conda_env = yaml.safe_load(f)
    elif not isinstance(conda_env, dict):
        raise TypeError(
            "Expected a string path to a conda env yaml file or a `dict` representing a conda env, "
            "but got `{}`".format(type(conda_env).__name__)
        )

    if conda_env is None:
        conda_env = {}

    if "dependencies" not in conda_env:
        conda_env["dependencies"] = []

    conda_env = _overwrite_python_version(conda_env, python_version)

    pip_reqs = _get_pip_deps(conda_env)

    for additional_merlin_req in additional_merlin_reqs:
        exist = False
        for pip_req in pip_reqs:
            pip_req_obj = Requirement(pip_req)
            additional_merlin_req_obj = Requirement(additional_merlin_req)
            if pip_req_obj.name.lower() == additional_merlin_req_obj.name.lower():
                exist = True
                break

        if not exist:
            pip_reqs.append(additional_merlin_req)

    conda_env = _overwrite_pip_deps(conda_env, pip_reqs)

    return conda_env


def _overwrite_python_version(conda_env, python_version):
    # if "python" in conda_env["dependencies"]:
    #     conda_env["dependencies"].remove("python")
    # conda_env["dependencies"].append("python={}".format(python_version))

    for dep in conda_env["dependencies"]:
        if isinstance(dep, str) and "python" in dep:
            conda_env["dependencies"].remove(dep)

    conda_env["dependencies"].insert(0, "python={}".format(python_version))
    return conda_env


def _get_pip_deps(conda_env):
    """
    :return: The pip dependencies from the conda env
    """
    if conda_env is not None:
        for dep in conda_env["dependencies"]:
            if _is_pip_deps(dep):
                return dep["pip"]
    return []


def _is_pip_deps(dep):
    """
    Returns True if `dep` is a dict representing pip dependencies
    """
    return isinstance(dep, dict) and "pip" in dep


def _overwrite_pip_deps(conda_env, new_pip_deps):
    """
    Overwrites the pip dependencies section in the given conda env dictionary.

    {
        "name": "env",
        "channels": [...],
        "dependencies": [
            ...,
            "pip",
            {"pip": [...]},  <- Overwrite this
        ],
    }
    """
    deps = conda_env.get("dependencies", [])
    new_deps = []
    contains_pip_deps = False
    for dep in deps:
        if _is_pip_deps(dep):
            contains_pip_deps = True
            new_deps.append({"pip": new_pip_deps})
        else:
            new_deps.append(dep)

    if not contains_pip_deps:
        new_deps.append({"pip": new_pip_deps})

    return {**conda_env, "dependencies": new_deps}
