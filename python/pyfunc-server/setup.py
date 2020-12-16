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

from setuptools import setup, find_packages

tests_require = [
    'pytest',
    'pytest-tornasync',
    'mypy'
]
setup(
    name='pyfuncserver',
    version='0.5.2',
    author_email='merlin-dev@gojek.com',
    description='Model Server implementation for mlflow pyfunc model. \
                 Not intended for use outside KFServing Frameworks Images',
    long_description=open('README.md').read(),
    python_requires='>=3.7',
    packages=find_packages("pyfuncserver"),
    install_requires=[
        "kfserving==0.2.1.1",
        "argparse>=1.4.0",
        "numpy >= 1.8.2",
        "mlflow==1.6.0",
        "cloudpickle==1.2.2",
        "merlin-sdk==0.8.0",
        "prometheus_client==0.7.1",
        "uvloop>=0.14.0",
        "orjson==2.6.8"
    ],
    tests_require=tests_require,
    extras_require={'test': tests_require}
)
