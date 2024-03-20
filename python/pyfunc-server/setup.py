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

import imp
import os

from setuptools import find_packages, setup

version = imp.load_source(
    "pyfuncserver.version", os.path.join("pyfuncserver", "version.py")
).VERSION

with open("requirements.txt") as f:
    REQUIRE = f.read().splitlines()

TESTS_REQUIRE = [
    "joblib>=0.13.0,<1.2.0",  # >=1.2.0 upon upgrade of kserve's version
    "mypy",
    "pytest-benchmark",
    "pytest-tornasync",
    "pytest",
    "requests",
    "scikit-learn>=1.1.2",
    "types-protobuf",
    "types-requests",
    "xgboost==1.6.2",
]

setup(
    name="merlin-pyfunc-server",
    version=version,
    author_email="merlin-dev@gojek.com",
    description="Model Server implementation for Merlin PyFunc model",
    long_description=open("README.md").read(),
    long_description_content_type="text/markdown",
    python_requires=">=3.8,<3.11",
    packages=find_packages(exclude=["test"]),
    install_requires=REQUIRE,
    tests_require=TESTS_REQUIRE,
    extras_require={"test": TESTS_REQUIRE},
    entry_points="""
        [console_scripts]
        merlin-pyfunc-server=pyfuncserver.__main__:main
    """,
)
