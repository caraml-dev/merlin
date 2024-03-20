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
    "merlinpyspark.version", os.path.join("merlinpyspark", "version.py")
).VERSION

with open("requirements.txt") as f:
    REQUIRE = f.read().splitlines()

with open("requirements_test.txt") as f:
    TESTS_REQUIRE = f.read().splitlines()

setup(
    name="merlin-batch-predictor",
    version=version,
    author_email="merlin-dev@gojek.com",
    description="Base PySpark application for running Merlin prediction batch job",
    long_description=open("README.md").read(),
    long_description_content_type="text/markdown",
    python_requires=">=3.8,<3.11",
    packages=find_packages(exclude=["test"]),
    install_requires=REQUIRE,
    tests_require=TESTS_REQUIRE,
    extras_require={"test": TESTS_REQUIRE},
    entry_points="""
        [console_scripts]
        merlin-batch-predictor=merlinpyspark.__main__:main
    """,
)
