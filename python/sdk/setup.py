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
from setuptools import setup, find_packages

version = imp.load_source(
    'merlin.version', os.path.join('merlin', 'version.py')).VERSION

REQUIRES = [
    "certifi>=2017.4.17",
    "python-dateutil>=2.1",
    "six>=1.10",
    "mlflow>=1.2.0,<=1.23.0", #  for py3.11 due to proto -> "mlflow>=1.26.1",
    "google-cloud-storage>=1.19.0",
    "boto3>=1.9.84",
    "urllib3>=1.23",
    "PyPrind>=2.11.2",
    "google-auth>=1.11.0",
    "Click>=7.0",
    "cloudpickle==2.0.0",  # used by mlflow
    "cookiecutter>=1.7.2",
    "docker>=4.2.1",
    "PyYAML>=5.4",
    "protobuf<4.0.0", #  for py3.11 due to proto -> "protobuf>=4.0.0,<5.0dev",
    "caraml-upi-protos>=0.3.1"
]

TEST_REQUIRES = [
    "pytest",
    "pytest-dependency",
    "pytest-cov",
    "pytest-xdist",
    "urllib3-mock>=0.3.3",
    "requests",
    "xgboost==1.6.2",
    "scikit-learn==1.0.1",  # >=1.1.2 upon python 3.7 deprecation
    "joblib>=0.13.0,<1.2.0",  # >=1.2.0 upon upgrade of kserve's version
    "mypy>=0.812",
    "google-cloud-bigquery>=1.18.0",
    "google-cloud-bigquery-storage>=0.7.0",
    "grpcio>=1.31.0,<1.49.0",
    "recursive-diff>=1.0.0",
    "xarray"
]

setup(
    name="merlin-sdk",
    version=version,
    description="Python SDK for Merlin",
    url="https://github.com/gojek/merlin",
    author="Merlin",
    packages=find_packages(),
    package_data={"merlin": [
        "docker/pyfunc.Dockerfile", "docker/standard.Dockerfile"]},
    zip_safe=True,
    install_requires=REQUIRES,
    setup_requires=["setuptools_scm"],
    tests_require=TEST_REQUIRES,
    extras_require={'test': TEST_REQUIRES},
    python_requires='>=3.7,<3.11',
    long_description=open("README.md").read(),
    long_description_content_type='text/markdown',
    entry_points='''
        [console_scripts]
        merlin=merlin.merlin:cli
    '''
)
