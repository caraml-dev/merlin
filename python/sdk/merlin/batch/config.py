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
from typing import Optional, Dict

from merlin.batch.sink import Sink
from merlin.batch.source import Source


class ResultType(Enum):
    DOUBLE = "DOUBLE"
    FLOAT = "FLOAT"
    INTEGER = "INTEGER"
    LONG = "LONG"
    STRING = "STRING"
    ARRAY = "ARRAY"


class PredictionJobResourceRequest:
    """
    Resource request configuration for starting prediction job
    """

    def __init__(self, driver_cpu_request: str,
                 driver_memory_request: str,
                 executor_cpu_request: str,
                 executor_memory_request: str,
                 executor_replica: int):
        """
        Create resource request object

        :param driver_cpu_request: driver's cpu request in kubernetes request format (e.g. : 500m, 1, 2, etc)
        :param driver_memory_request: driver's memory request in kubernetes format (e.g.: 512Mi, 1Gi, 2Gi, etc)
        :param executor_cpu_request: executors's cpu request in kubernetes request format (e.g. : 500m, 1, 2, etc)
        :param executor_memory_request: executors's memory request in kubernetes format (e.g.: 512Mi, 1Gi, 2Gi, etc)
        :param executor_replica: number of executor to be used
        """
        self._driver_cpu_request = driver_cpu_request
        self._driver_memory_request = driver_memory_request
        self._executor_cpu_request = executor_cpu_request
        self._executor_memory_request = executor_memory_request
        self._executor_replica = executor_replica

    def to_dict(self):
        return {
            "driver_cpu_request": self._driver_cpu_request,
            "driver_memory_request": self._driver_memory_request,
            "executor_cpu_request": self._executor_cpu_request,
            "executor_memory_request": self._executor_memory_request,
            "executor_replica": self._executor_replica
        }


class PredictionJobConfig:
    def __init__(self, source: Source, sink: Sink,
                 service_account_name: str,
                 result_type: ResultType = ResultType.DOUBLE,
                 item_type: ResultType = ResultType.DOUBLE,
                 resource_request: PredictionJobResourceRequest = None,
                 env_vars: Dict[str, str] = None):
        """
        Create configuration for starting a prediction job

        :param source: source configuration. See merlin.batch.source package.
        :param sink: sink configuration. See merlin.batch.sink package
        :param service_account_name: secret name containing the service account for executing the prediction job.
        :param result_type: type of the prediction result (default to ResultType.DOUBLE).
        :param item_type: item type of the prediction result if the result_type is ResultType.ARRAY. Otherwise will be ignored.
        :param resource_request: optional resource request for starting the prediction job. If not given the system default will be used.
        :param env_vars: optional environment variables in the form of a key value pair in a list.
        """

        self._source = source
        self._sink = sink
        self._service_account_name = service_account_name
        self._resource_request = resource_request
        self._result_type = result_type
        self._item_type = item_type
        self._env_vars = env_vars

    @property
    def source(self) -> Source:
        return self._source

    @property
    def sink(self) -> Sink:
        return self._sink

    @property
    def service_account_name(self) -> str:
        return self._service_account_name

    @property
    def resource_request(self) -> Optional[PredictionJobResourceRequest]:
        return self._resource_request

    @property
    def result_type(self) -> ResultType:
        return self._result_type

    @property
    def item_type(self) -> ResultType:
        return self._item_type

    @property
    def env_vars(self) -> Optional[Dict[str, str]]:
        return self._env_vars
