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
from typing import Iterable, List

from pyspark.sql import DataFrame, SparkSession

from merlinpyspark.config import BigQuerySourceConfig, SourceConfig


def create_source(spark_session: SparkSession,
                  source_config: SourceConfig) -> "Source":
    if source_config.source_type() == BigQuerySourceConfig.TYPE:
        if not isinstance(source_config, BigQuerySourceConfig):
            raise ValueError('source_config is not BigQuerySourceConfig '
                             'instance')

        return BigQuerySource(spark_session, source_config)

    raise NotImplementedError(f"{source_config.source_type()} source type "
                              f"is not implemented")


class Source(ABC):
    @abstractmethod
    def load(self) -> DataFrame:
        """
        Loads a data stream from a data source and returns it as a :class`DataFrame`.
        :return: Spark DataFrame
        """
        pass

    @abstractmethod
    def features(self) -> Iterable[str]:
        """
        List of columns used for prediction
        :return: list of string
        """
        pass


class BigQuerySource(Source):
    READ_FORMAT = "bigquery"
    OPTION_TABLE = "table"
    OPTION_PARALLELISM = "maxParallelism"

    def __init__(self, spark_session: SparkSession,
                 bq_source_config: BigQuerySourceConfig):
        self._spark = spark_session
        self._config = bq_source_config

        if bq_source_config.table() is None:
            raise ValueError("table field is empty")

    def load(self) -> DataFrame:
        cfg = self._config
        features = self.features()

        reader = self._spark.read.format(self.READ_FORMAT) \
            .option(self.OPTION_TABLE, cfg.table())

        if cfg.options() is not None:
            reader.options(**cfg.options())

        df = reader.load()
        if features is not None:
            return df.select(*features)

        return df

    def features(self) -> Iterable[str]:
        return self._config.features()
