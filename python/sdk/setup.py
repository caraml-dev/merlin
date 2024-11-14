#!/usr/bin/env python
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
    "merlin.version", os.path.join("merlin", "version.py")
).VERSION

with open("requirements.txt") as f:
    REQUIRE = f.read().splitlines()

with open("requirements_test.txt") as f:
    TESTS_REQUIRE = f.read().splitlines()

setup(
    name="merlin-sdk",
    version=version,
    description="Python SDK for Merlin",
    url="https://github.com/caraml-dev/merlin",
    author="Merlin",
    packages=find_packages(),
    package_data={"merlin": ["docker/pyfunc.Dockerfile", "docker/standard.Dockerfile"]},
    zip_safe=True,
    install_requires=REQUIRE,
    setup_requires=["setuptools_scm"],
    tests_require=TESTS_REQUIRE,
    extras_require={"test": TESTS_REQUIRE},
    python_requires=">=3.8,<3.11",
    long_description=open("README.md").read(),
    long_description_content_type="text/markdown",
    entry_points="""
        [console_scripts]
        merlin=merlin.merlin:cli
    """,
)
