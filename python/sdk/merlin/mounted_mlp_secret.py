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
from __future__ import annotations

import client


class MountedMLPSecret:
    """
    The MLP secret mounted to the Merlin model/transformer as an environment variable.
    """

    def __init__(
        self,
        mlp_secret_name: str,
        env_var_name: str,
    ):
        self._mlp_secret_name = mlp_secret_name
        self._env_var_name = env_var_name

    @classmethod
    def from_response(cls, response: client.MountedMLPSecret) -> MountedMLPSecret:
        return MountedMLPSecret(
            mlp_secret_name=response.mlp_secret_name,
            env_var_name=response.env_var_name
        )

    @property
    def mlp_secret_name(self) -> str:
        return self._mlp_secret_name

    @mlp_secret_name.setter
    def mlp_secret_name(self, mlp_secret_name):
        self._mlp_secret_name = mlp_secret_name

    @property
    def env_var_name(self) -> str:
        return self._env_var_name

    @env_var_name.setter
    def env_var_name(self, env_var_name):
        self._env_var_name = env_var_name
