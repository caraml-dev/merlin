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
import os

from merlinpyspark.config import (
    BigQuerySourceConfig,
    SourceConfig,
    MaxComputeSourceConfig,
)


def create_source(spark_session: SparkSession, source_config: SourceConfig) -> "Source":
    source_type = source_config.source_type()

    if source_type == BigQuerySourceConfig.TYPE:
        if not isinstance(source_config, BigQuerySourceConfig):
            raise ValueError("source_config is not of BigQuerySourceConfig instance")
        return BigQuerySource(spark_session, source_config)

    if source_type == MaxComputeSourceConfig.TYPE:
        if not isinstance(source_config, MaxComputeSourceConfig):
            raise ValueError("source_config is not of MaxComputeSourceConfig instance")
        return MaxComputeSource(spark_session, source_config)

    raise NotImplementedError(f"{source_type} source type is not implemented")


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

    def __init__(
        self, spark_session: SparkSession, bq_source_config: BigQuerySourceConfig
    ):
        self._spark = spark_session
        self._config = bq_source_config

        if bq_source_config.table() is None:
            raise ValueError("table field is empty")

    def load(self) -> DataFrame:
        cfg = self._config
        features = self.features()

        reader = self._spark.read.format(self.READ_FORMAT).option(
            self.OPTION_TABLE, cfg.table()
        )

        if cfg.options() is not None:
            reader.options(**cfg.options())

        df = reader.load()
        if features is not None:
            return df.select(*features)

        return df

    def features(self) -> Iterable[str]:
        return self._config.features()


class MaxComputeSource(Source):
    READ_FORMAT = "jdbc"
    OPTION_URL = "url"
    OPTION_DRIVER = "driver"
    OPTION_QUERY_TIMEOUT = "queryTimeout"
    OPTION_DB_TABLE = "dbtable"
    OPTION_FETCH_SIZE = "fetchSize"

    def __init__(
        self,
        spark_session: SparkSession,
        maxcompute_source_config: MaxComputeSourceConfig,
    ):
        self._spark = spark_session
        self._config = maxcompute_source_config

        if maxcompute_source_config.table() is None:
            raise ValueError("table field is empty")

        if maxcompute_source_config.project() is None:
            raise ValueError("project field is empty")

        if maxcompute_source_config.schema() is None:
            raise ValueError("schema field is empty")

    def load(self) -> DataFrame:
        from py4j.java_gateway import java_import

        gw = self._spark.sparkContext._gateway # type: ignore
        java_import(gw.jvm, self._get_custom_dialect_class())
        gw.jvm.org.apache.spark.sql.jdbc.JdbcDialects.registerDialect(
            gw.jvm.dev.caraml.spark.odps.CustomDialect()
        )
        cfg = self._config
        reader = (
            self._spark.read.format(self.READ_FORMAT)
            .option(self.OPTION_DRIVER, self.get_jdbc_driver())
            .option(self.OPTION_URL, self.get_jdbc_url())
            .option(self.OPTION_QUERY_TIMEOUT, self.get_query_timeout())
            .option(self.OPTION_DB_TABLE, cfg.table())
            .option(self.OPTION_FETCH_SIZE, self.get_fetch_size())
        )
        if cfg.options() is not None:
            reader.options(**cfg.options())
        df = reader.load()
        if self.features() is not None:
            return df.select(*self.features())
        return df

    def features(self) -> Iterable[str]:
        return self._config.features()

    def get_jdbc_url(self):
        return f"jdbc:odps:{self._config.endpoint()}?project={self._config.project()}&accessId={self.get_access_id()}&accessKey={self.get_access_key()}&interactiveMode={self.get_interactive_mode()}&odpsNamespaceSchema=true&schema={self._config.schema()}&enableLimit=false"

    def get_query_timeout(self):
        return self._config.options().get("queryTimeout", "300")

    def get_interactive_mode(self):
        return self._config.options().get("interactiveMode", "true")    

    def get_fetch_size(self):
        return self._config.options().get("fetchSize", "0")

    def get_jdbc_driver(self):
        return os.environ.get("ODPS_JDBC_DRIVER", "com.aliyun.odps.jdbc.OdpsDriver")

    def get_access_id(self):
        # NOTE: access id and key will not be part of the PredictionConfig
        # since these are mounted from a configmap
        # these should be passed in via environment variable
        return os.environ.get("ODPS_ACCESS_ID")

    def get_access_key(self):
        # NOTE: access id and key will not be part of the PredictionConfig
        # since these are mounted from a configmap
        # these should be passed in via environment variable
        return os.environ.get("ODPS_SECRET_KEY")

    def _get_custom_dialect_class(self):
        # NOTE: this is hardcoded because of how it should be imported in the spark context
        return "dev.caraml.spark.odps.CustomDialect"
