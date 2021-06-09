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

from enum import Enum
import client
from merlin.util import autostr
from typing import Optional


@autostr
class CustomPredictor:

    def __init__(self, image: str, command: str, args: str):
        self._image = image
        self._command = command
        self._args = args

    @property
    def image(self) -> str:
        return self._image

    @property
    def enabled(self) -> bool:
        return self._enabled

    @property
    def command(self) -> Optional[str]:
        return self._command

    @property
    def args(self) -> Optional[str]:
        return self._args



