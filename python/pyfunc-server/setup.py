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

import importlib.util
import os

from setuptools import find_packages, setup

# get version from version.py
spec = importlib.util.spec_from_file_location(
    "pyfuncserver.version", os.path.join("pyfuncserver","version.py")
)

v_module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(v_module)

version = v_module.VERSION

with open("requirements.txt") as f:
    REQUIRE = f.read().splitlines()

TESTS_REQUIRE = [
    "joblib>=1.2.0",
    "mypy>=1.5.4",
    "pytest>=8.1",
    "pytest-benchmark>=5.1.0",
    "pytest-tornasync",
    "requests>=2.31.0",
    "scikit-learn>=1.3.1",
    "types-protobuf",
    "types-requests",
    "xgboost>=1.7.6",
]

setup(
    name="merlin-pyfunc-server",
    version=version,
    author_email="merlin-dev@gojek.com",
    description="Model Server implementation for Merlin PyFunc model",
    long_description=open("README.md").read(),
    long_description_content_type="text/markdown",
    python_requires=">=3.8,<3.14",
    packages=find_packages(exclude=["test"]),
    install_requires=REQUIRE,
    tests_require=TESTS_REQUIRE,
    extras_require={"test": TESTS_REQUIRE},
    entry_points="""
        [console_scripts]
        merlin-pyfunc-server=pyfuncserver.__main__:main
    """,
)
