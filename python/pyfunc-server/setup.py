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
from setuptools import setup, find_packages

tests_require = [
    'pytest',
    'pytest-tornasync',
    'requests',
    'types-requests',
    'mypy'
]

with open('requirements.txt') as f:
    REQUIRE = f.read().splitlines()


# replace merlin relative path in requirements.txt into absolute path
# setuptools could not install relative path requirements
merlin_path = os.path.join(os.getcwd(), "../sdk")
merlin_sdk_package = "merlin-sdk"
for index, item in enumerate(REQUIRE):
    if merlin_sdk_package in item:
        REQUIRE[index] = f"{merlin_sdk_package} @ file://localhost/{merlin_path}#egg={merlin_sdk_package}"

setup(
    name='pyfuncserver',
    version='0.5.2',
    author_email='merlin-dev@gojek.com',
    description='Model Server implementation for mlflow pyfunc model. \
                 Not intended for use outside KFServing Frameworks Images',
    long_description=open('README.md').read(),
    python_requires='>=3.7',
    packages=find_packages(exclude=["tests"]),
    install_requires=REQUIRE,
    tests_require=tests_require,
    extras_require={'test': tests_require}
)
