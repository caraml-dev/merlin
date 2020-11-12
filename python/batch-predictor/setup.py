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

with open('requirements_test.txt') as f:
    TEST_REQUIRE = f.read().splitlines()

with open('requirements.txt') as f:
    REQUIRE = f.read().splitlines()

setup(
    name='merlin-pyspark-app',
    version='0.1.0',
    author_email='merlin-dev@gojek.com',
    description='Base pyspark application for running merlin prediction job',
    long_description=open('README.md').read(),
    python_requires='>=3.6',
    packages=find_packages("merlinpyspark"),
    install_requires=REQUIRE,
    tests_require=TEST_REQUIRE,
    extras_require={'test': TEST_REQUIRE}
)
