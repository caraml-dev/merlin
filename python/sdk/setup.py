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

REQUIRES = [
    "boto3>=1.9.84",
    "caraml-upi-protos>=0.3.1",
    "certifi>=2017.4.17",
    "Click>=7.0,<8.1.4",
    "cloudpickle==2.0.0",  # used by mlflow
    "cookiecutter>=1.7.2",
    "dataclasses-json>=0.5.2",  # allow Flyte version 1.2.0 or above to import Merlin SDK
    "docker<=6.1.3",
    "GitPython>=3.1.40",
    "google-cloud-storage>=1.19.0",
    "protobuf>=3.12.0,<5.0.0",  # Determined by the mlflow dependency
    "mlflow==1.26.1",
    "PyPrind>=2.11.2",
    "python_dateutil>=2.5.3",
    "PyYAML>=5.4",
    "six>=1.10",
    "urllib3>=1.26",
    "numpy<=1.23.5",  # Temporary pin numpy due to https://numpy.org/doc/stable/release/1.20.0-notes.html#numpy-1-20-0-release-notes
    "caraml-auth-google==0.0.0.post7",
    "pydantic==2.5.3"
]

TEST_REQUIRES = [
    "google-cloud-bigquery-storage>=0.7.0",
    "google-cloud-bigquery>=1.18.0",
    "joblib>=0.13.0,<1.2.0",  # >=1.2.0 upon upgrade of kserve's version
    "mypy>=0.812",
    "pytest-cov",
    "pytest-dependency",
    "pytest-xdist",
    "pytest",
    "recursive-diff>=1.0.0",
    "requests",
    "scikit-learn>=1.1.2",
    "types-python-dateutil",
    "types-PyYAML",
    "types-six",
    "types-protobuf",
    "urllib3-mock>=0.3.3",
    "xarray",
    "xgboost==1.6.2",
]

setup(
    name="merlin-sdk",
    version=version,
    description="Python SDK for Merlin",
    url="https://github.com/caraml-dev/merlin",
    author="Merlin",
    packages=find_packages(),
    package_data={"merlin": ["docker/pyfunc.Dockerfile", "docker/standard.Dockerfile"]},
    zip_safe=True,
    install_requires=REQUIRES,
    setup_requires=["setuptools_scm"],
    tests_require=TEST_REQUIRES,
    extras_require={"test": TEST_REQUIRES},
    python_requires=">=3.8,<3.11",
    long_description=open("README.md").read(),
    long_description_content_type="text/markdown",
    entry_points="""
        [console_scripts]
        merlin=merlin.merlin:cli
    """,
)
