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

from abc import ABC, abstractmethod
from enum import Enum
from typing import MutableMapping, Mapping, Any, Optional

import client
from merlin.batch.big_query_util import bq_valid_table_id, bq_valid_column
from merlin.batch.maxcompute_util import mc_valid_table_id, mc_valid_columns


class Sink(ABC):
    @abstractmethod
    def to_dict(self) -> Mapping[str, Any]:
        pass


class SaveMode(Enum):
    ERRORIFEXISTS = 0
    OVERWRITE = 1
    APPEND = 2
    IGNORE = 3
    ERROR = 4


class BigQuerySink(Sink):
    """
        Sink contract for BigQuery to create prediction job
    """

    def __init__(self, table: str,
                 staging_bucket: str,
                 result_column: str,
                 save_mode: SaveMode = SaveMode.ERRORIFEXISTS,
                 options: MutableMapping[str, str] = None):
        """
        :param table: table id of destination BQ table in format `gcp-project.dataset.table_name`
        :param staging_bucket: temporary GCS bucket for staging write into BQ table
        :param result_column: column name that will be used to store prediction result.
        :param save_mode: save mode. Default to SaveMode.ERRORIFEXISTS. Which will fail if destination table already exists
        :param options: additional sink option to configure the prediction job.
        """
        self._table = table
        self._staging_bucket = staging_bucket
        self._result_column = result_column
        self._save_mode = save_mode
        self._options = options

    @property
    def table(self) -> str:
        return self._table

    @table.setter
    def table(self, table):
        self._table = table

    @property
    def staging_bucket(self) -> str:
        return self._staging_bucket

    @staging_bucket.setter
    def staging_bucket(self, staging_bucket):
        self._staging_bucket = staging_bucket

    @property
    def result_column(self) -> str:
        return self._result_column

    @result_column.setter
    def result_column(self, result_column):
        self._result_column = result_column

    @property
    def save_mode(self) -> SaveMode:
        return self._save_mode

    @save_mode.setter
    def save_mode(self, save_mode):
        self._save_mode = save_mode

    @property
    def options(self) -> Optional[MutableMapping[str, str]]:
        return self._options

    @options.setter
    def options(self, options):
        self._options = options

    def _validate(self):
        if not self._valid_types():
            raise ValueError("invalid types")
        if not bq_valid_table_id(self._table):
            raise ValueError(f"invalid table id: {self.table}")
        if not bq_valid_column(self._result_column):
            raise ValueError(f"invalid result column: {self.result_column}")
        return True

    def _valid_types(self):
        if not isinstance(self._table, str):
            return False
        if not isinstance(self._staging_bucket, str):
            return False
        if not isinstance(self._result_column, str):
            return False
        if not isinstance(self._save_mode, SaveMode):
            return False
        if self._options is not None and not isinstance(self._options, MutableMapping):
            return False

        return True

    def to_dict(self) -> Mapping[str, Any]:
        self._validate()

        opts = self._options
        if opts is None:
            opts = {}
        return {
            'table': self._table,
            'staging_bucket': self._staging_bucket,
            'result_column': self._result_column,
            'save_mode': self._save_mode.value,
            'options': opts
        }

    def to_client_config(self) -> client.PredictionJobConfigBigquerySink:
        opts = {}
        if self.options is not None:
            for k, v in self.options.items():
                opts[k] = v

        return client.PredictionJobConfigBigquerySink(
            table=self._table,
            staging_bucket=self._staging_bucket,
            result_column=self._result_column,
            save_mode=client.SaveMode(self._save_mode.value),
            options=opts
        )

class MaxComputeSink(Sink):
    """
        Sink contract for MaxCompute to create prediction job
    """

    def __init__(self, 
                 endpoint: str,
                 table: str,
                 result_column: str,
                 save_mode: SaveMode = SaveMode.ERRORIFEXISTS,
                 options: MutableMapping[str, str] = None):
        """
        :param endpoint: MaxCompute endpoint
        :param table: table name in MaxCompute, in the format of `project_name.schema.table_name`
        :param result_column: column name that will be used to store prediction result.
        :param save_mode: save mode. Default to SaveMode.ERRORIFEXISTS. Which will fail if destination table already exists
        :param options: additional sink option to configure the prediction job.
        """
        self._endpoint = endpoint
        self._table = table
        self._result_column = result_column
        self._save_mode = save_mode
        self._options = options

    @property
    def endpoint(self) -> str:
        return self._endpoint

    @endpoint.setter
    def endpoint(self, endpoint):
        self._endpoint = endpoint

    @property
    def table(self) -> str:
        return self._table

    @table.setter
    def table(self, table):
        self._table = table

    @property
    def result_column(self) -> str:
        return self._result_column

    @result_column.setter
    def result_column(self, result_column):
        self._result_column = result_column

    @property
    def options(self) -> Optional[MutableMapping[str, str]]:
        return self._options

    @options.setter
    def options(self, options):
        self._options = options

    def _valid_types(self):
        if not isinstance(self._endpoint, str):
            return False
        if not isinstance(self._table, str):
            return False
        if not isinstance(self._result_column, str):
            return False
        if not isinstance(self._save_mode, SaveMode):
            return False
        if self._options is not None and not isinstance(self._options, MutableMapping):
            return False

        return


    def _validate(self):
        if not self._valid_types():
            raise ValueError("invalid input type")
        if not mc_valid_table_id(self.table):
            raise ValueError(f"invalid table: {self.table}")
        if not mc_valid_columns(self.result_column):
            raise ValueError(f"invalid result column: {self.result_column}")

    def to_dict(self) -> Mapping[str, Any]:
        self._validate()

        opts = self._options
        if opts is None:
            opts = {}
        return {
            'endpoint': self._endpoint,
            'table': self._table,
            'result_column': self._result_column,
            'save_mode': self._save_mode.value,
            'options': opts
        }

    def to_client_config(self) -> client.PredictionJobConfigMaxcomputeSink:
        opts = {}
        if self.options is not None:
            for k, v in self.options.items():
                opts[k] = v

        return client.PredictionJobConfigMaxcomputeSink(
            endpoint=self._endpoint,
            table=self._table,
            result_column=self._result_column,
            save_mode=client.SaveMode(self._save_mode.value),
            options=opts
        )