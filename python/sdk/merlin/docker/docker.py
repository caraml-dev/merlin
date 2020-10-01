# Copyright 2020 The Merlin Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import os
import shutil
import pathlib

try:
    import importlib.resources as pkg_resources
except ImportError:
    # Try backported to PY<37 `importlib_resources`.

    import importlib_resources as pkg_resources  # type: ignore


def copy_pyfunc_dockerfile(dst):
    """
    Copy pyfunc dockerfile to the destination path

    :param dst: destination path
    :return: full path to pyfunc Dockerfile
    """
    return _copy_dockerfile(dst, "pyfunc.Dockerfile")


def copy_standard_dockerfile(dst):
    """
    Copy dockerfile for standarnd model to the destination path

    :param dst: detination path
    :return: full path to standard model Dockerfile
    """
    return _copy_dockerfile(dst, "standard.Dockerfile")


def _copy_dockerfile(dst, dockerfile_name):
    dst_dockerfile_path = os.path.join(dst, dockerfile_name)
    if os.path.isfile(dst_dockerfile_path):
        return dst_dockerfile_path

    absolute_package_path = pathlib.Path(__file__).absolute()
    package_path = os.path.dirname(absolute_package_path)
    full_path = os.path.join(package_path, dockerfile_name)
    shutil.copy(full_path, dst_dockerfile_path)
    return dst_dockerfile_path
