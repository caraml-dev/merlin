import os
from collections import namedtuple
from itertools import filterfalse
from pathlib import Path
from typing import Any, List, Optional

import yaml
from packaging.requirements import Requirement, InvalidRequirement

_CONSTRAINTS_FILE_NAME = "constraints.txt"


def process_conda_env(
    conda_env: Any,
    python_version: Optional[str] = "3.10.*",
    additional_merlin_reqs: Optional[List[str]] = None,
):
    """
    Processes `conda_env` passed to `merlin.log_pyfunc_model` and `merlin.log_model`.
    This function will do the following:
    1. Add/replace Python version on the conda_env with `python_version` argument.
    2. Add `additional_merlin_reqs` onto the conda_env's pip dependencies.

    :param conda_env: Either a dictionary representation of a conda environment or the path to a conda environment yaml file
    :param python_version: The python version to replace the conda environment with
    :param additonal_merlin_reqs: A list of additional Merlin requirements to be added to the conda environment
    :return: dict representation of the conda environment
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
    pip_reqs, constraints = _parse_pip_requirements(pip_reqs)

    if additional_merlin_reqs:
        for additional_merlin_req in additional_merlin_reqs:
            exist = False
            for pip_req in pip_reqs:
                # Skip requirement comparison for pip reqs that is not a package, which can happen if the user specifies
                # repository or trusted host as part of the pip file
                if pip_req.strip().startswith("--"):
                    continue

                pip_req_obj = Requirement(pip_req)
                additional_merlin_req_obj = Requirement(additional_merlin_req)
                if pip_req_obj.name.lower() == additional_merlin_req_obj.name.lower():
                    exist = True
                    break

            if not exist:
                pip_reqs.insert(0, additional_merlin_req)

    if constraints:
        pip_reqs.append(f"-c {_CONSTRAINTS_FILE_NAME}")

    conda_env = _overwrite_pip_deps(conda_env, pip_reqs)

    return conda_env


def _overwrite_python_version(conda_env, python_version):
    """
    Overwrites the python version in the given conda env dictionary.

    {
        "name": "env",
        "channels": [...],
        "dependencies": [
            ...,
            "python=3.10.*",  <- Overwrite this
        ],
    }
    """
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


def _parse_pip_requirements(pip_requirements):
    """
    Parses an iterable of pip requirement strings or a pip requirements file.

    :param pip_requirements: Either an iterable of pip requirement strings
        (e.g. ``["scikit-learn", "-r requirements.txt"]``) or the string path to a pip requirements
        file on the local filesystem (e.g. ``"requirements.txt"``). If ``None``, an empty list will
        be returned.
    :return: A tuple of parsed requirements and constraints.
    """
    if pip_requirements is None:
        return [], []

    def _is_string(x):
        return isinstance(x, str)

    def _is_iterable(x):
        try:
            iter(x)
            return True
        except Exception:
            return False

    if _is_string(pip_requirements):
        with open(pip_requirements) as f:
            return _parse_pip_requirements(f.read().splitlines())
    elif _is_iterable(pip_requirements) and all(map(_is_string, pip_requirements)):
        requirements = []
        constraints = []
        for req_or_con in _parse_requirements(pip_requirements, is_constraint=False):
            if req_or_con.is_constraint:
                constraints.append(req_or_con.req_str)
            else:
                requirements.append(req_or_con.req_str)

        return requirements, constraints
    else:
        raise TypeError(
            "`pip_requirements` must be either a string path to a pip requirements file on the "
            "local filesystem or an iterable of pip requirement strings, but got `{}`".format(
                type(pip_requirements)
            )
        )


# Represents a pip requirement.
#
# :param req_str: A requirement string (e.g. "scikit-learn == 0.24.2").
# :param is_constraint: A boolean indicating whether this requirement is a constraint.
_Requirement = namedtuple("_Requirement", ["req_str", "is_constraint"])


def _parse_requirements(requirements, is_constraint, base_dir=None):
    """
    A simplified version of `pip._internal.req.parse_requirements` which performs the following
    operations on the given requirements file and yields the parsed requirements.

    - Remove comments and blank lines
    - Join continued lines
    - Resolve requirements file references (e.g. '-r requirements.txt')
    - Resolve constraints file references (e.g. '-c constraints.txt')

    :param requirements: A string path to a requirements file on the local filesystem or
                         an iterable of pip requirement strings.
    :param is_constraint: Indicates the parsed requirements file is a constraint file.
    :param base_dir: If specified, resolve relave file references (e.g. '-r requirements.txt')
                     against the specified directory.
    :return: A list of ``_Requirement`` instances.

    References:
    - `pip._internal.req.parse_requirements`:
      https://github.com/pypa/pip/blob/7a77484a492c8f1e1f5ef24eaf71a43df9ea47eb/src/pip/_internal/req/req_file.py#L118
    - Requirements File Format:
      https://pip.pypa.io/en/stable/cli/pip_install/#requirements-file-format
    - Constraints Files:
      https://pip.pypa.io/en/stable/user_guide/#constraints-files
    """
    if base_dir is None:
        if isinstance(requirements, (str, Path)):
            base_dir = os.path.dirname(requirements)
            with open(requirements) as f:
                requirements = f.read().splitlines()
        else:
            base_dir = os.getcwd()

    lines = map(str.strip, requirements)
    lines = map(_strip_inline_comment, lines)
    lines = _join_continued_lines(lines)
    lines = filterfalse(_is_comment, lines)
    lines = filterfalse(_is_empty, lines)

    for line in lines:
        if _is_requirements_file(line):
            req_file = line.split(maxsplit=1)[1]
            # If `req_file` is an absolute path, `os.path.join` returns `req_file`:
            # https://docs.python.org/3/library/os.path.html#os.path.join
            abs_path = os.path.join(base_dir, req_file)
            yield from _parse_requirements(abs_path, is_constraint=False)
        elif _is_constraints_file(line):
            req_file = line.split(maxsplit=1)[1]
            abs_path = os.path.join(base_dir, req_file)
            yield from _parse_requirements(abs_path, is_constraint=True)
        else:
            yield _Requirement(line, is_constraint)


def _is_comment(line):
    return line.startswith("#")


def _is_empty(line):
    return line == ""


def _strip_inline_comment(line):
    return line[: line.find(" #")].rstrip() if " #" in line else line


def _is_requirements_file(line):
    return line.startswith("-r ") or line.startswith("--requirement ")


def _is_constraints_file(line):
    return line.startswith("-c ") or line.startswith("--constraint ")


def _join_continued_lines(lines):
    """
    Joins lines ending with '\\'.

    >>> _join_continued_lines["a\\", "b\\", "c"]
    >>> 'abc'
    """
    continued_lines = []

    for line in lines:
        if line.endswith("\\"):
            continued_lines.append(line.rstrip("\\"))
        else:
            continued_lines.append(line)
            yield "".join(continued_lines)
            continued_lines.clear()

    # The last line ends with '\'
    if continued_lines:
        yield "".join(continued_lines)


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
