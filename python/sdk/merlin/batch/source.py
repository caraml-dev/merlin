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
from typing import Iterable, MutableMapping, Mapping, Any, Optional
from merlin.batch.big_query_util import valid_table_id, valid_columns


class Source(ABC):
    @abstractmethod
    def to_dict(self) -> Mapping[str, Any]:
        pass


class BigQuerySource(Source):
    """
        Source contract for BigQuery to create prediction job
    """

    def __init__(self, table: str, features: Iterable[str],
                 options: MutableMapping[str, str] = None):
        """

        :param table: table id if the source in format of `gcp-project.dataset.table_name`
        :param features: list of features to be used for prediction, it has to match the column name in the source table.
        :param options: additional option to configure source.
        """
        self._table = table
        self._features = features
        self._options = options

    @property
    def table(self) -> str:
        return self._table

    @table.setter
    def table(self, table):
        self._table = table

    @property
    def features(self) -> Iterable[str]:
        return self._features

    @features.setter
    def features(self, features):
        self._features = features

    @property
    def options(self) -> Optional[MutableMapping[str, str]]:
        return self._options

    @options.setter
    def options(self, options):
        self._options = options

    def _validate(self):
        if not self._valid_types():
            raise ValueError("invalid input type")
        if not valid_table_id(self.table):
            raise ValueError(f"invalid table: {self.table}")
        if not valid_columns(self.features):
            raise ValueError(f"invalid features column: {self.features}")

    def _valid_types(self) -> bool:
        if not isinstance(self._table, str):
            return False
        if not isinstance(self._features, list):
            return False
        if self._options is not None and not isinstance(self._options,
                                                        MutableMapping):
            return False

        for feature in self._features:
            if not isinstance(feature, str):
                return False

        return True

    def to_dict(self) -> Mapping[str, Any]:
        self._validate()

        opts = self.options
        if opts is None:
            opts = {}
        return {
            'table': self._table,
            'features': self._features,
            'options': opts
        }
