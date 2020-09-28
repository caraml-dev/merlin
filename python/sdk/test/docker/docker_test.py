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
from merlin.docker.docker import copy_pyfunc_dockerfile, copy_standard_dockerfile


def test_copy_pyfunc_dockerfile():
    path = copy_pyfunc_dockerfile(".")
    assert os.path.isfile(path) == True
    assert os.path.basename(path) == "pyfunc.Dockerfile"


def test_copy_standard_dockerfile():
    path = copy_standard_dockerfile(".")
    assert os.path.isfile(path) == True
    assert os.path.basename(path) == "standard.Dockerfile"
