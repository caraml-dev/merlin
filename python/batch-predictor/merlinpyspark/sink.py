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

from pyspark.sql import DataFrame

from merlinpyspark.config import SinkConfig, BigQuerySinkConfig, MaxComputeSinkConfig
import os

from pyspark.sql import SparkSession

def create_sink(spark_session: SparkSession, sink_config: SinkConfig) -> "Sink":
    sink_type = sink_config.sink_type()
    if sink_type == BigQuerySinkConfig.TYPE:
        if not isinstance(sink_config, BigQuerySinkConfig):
            raise ValueError("sink_config is not BigQuerySink")

        return BigQuerySink(sink_config)

    if sink_type == MaxComputeSinkConfig.TYPE:
        if not isinstance(sink_config, MaxComputeSinkConfig):
            raise ValueError("sink_config is not MaxComputeSink")

        return MaxComputeSink(spark_session, sink_config)

    raise ValueError(f"sink type is not implemented: {sink_config.sink_type()}")


class Sink(ABC):
    @abstractmethod
    def save(self, df):
        pass


class BigQuerySink(Sink):
    WRITE_FORMAT = "bigquery"
    OPTION_TABLE = "table"
    OPTION_STAGING_BUCKET = "temporaryGcsBucket"

    def __init__(self, config: BigQuerySinkConfig):
        self._config = config

    def save(self, df: DataFrame):
        df.write.mode(self._config.save_mode()).format(self.WRITE_FORMAT).option(
            self.OPTION_TABLE, self._config.table()
        ).option(self.OPTION_STAGING_BUCKET, self._config.staging_bucket()).options(
            **self._config.options()
        ).save()


class MaxComputeSink(Sink):
    WRITE_FORMAT = "jdbc"
    OPTION_URL = "url"
    OPTION_DRIVER = "driver"
    OPTION_QUERY_TIMEOUT = "queryTimeout"
    OPTION_DB_TABLE = "dbtable"
    OPTION_BATCH_SIZE = "batchSize"

    def __init__(self, spark_session: SparkSession, config: MaxComputeSinkConfig):
        self._config = config
        self._spark = spark_session

    def get_jdbc_url(self):
        return f"jdbc:odps:{self._config.endpoint()}?project={self._config.project()}&accessId={self.get_access_id()}&accessKey={self.get_access_key()}&interactiveMode={self.get_interactive_mode()}&odpsNamespaceSchema=true&schema={self._config.schema()}&enableLimit=false"

    def get_query_timeout(self):
        return self._config.options().get("queryTimeout", "300")

    def get_interactive_mode(self):
        return self._config.options().get("interactiveMode", "true")    

    def get_batch_size(self):
        return self._config.options().get("batchSize", "10000")


    def get_jdbc_driver(self):
        return os.environ.get("ODPS_JDBC_DRIVER", "com.aliyun.odps.jdbc.OdpsDriver")

    def get_access_id(self):
        # NOTE: access id and key will not be part of the Prediction Job proto message
        # these should be passed in via environment variable
        return os.environ.get("ODPS_ACCESS_ID")

    def get_access_key(self):
        # NOTE: access id and key will not be part of the Prediction Job proto message
        # these should be passed in via environment variable
        return os.environ.get("ODPS_SECRET_KEY")
    
    def _get_custom_dialect_class(self):
        # NOTE: this is hardcoded because the gateway function that registers the dialect
        # cannot be passed in at runtime:
        # gw.jvm.org.apache.spark.sql.jdbc.JdbcDialects.registerDialect(
        #    gw.jvm.dev.caraml.spark.odps.CustomDialect()
        # )
        return "dev.caraml.spark.odps.CustomDialect"

    def save(self, df: DataFrame):
        from py4j.java_gateway import java_import
        cfg = self._config

        gw = self._spark.sparkContext._gateway # type: ignore
        java_import(gw.jvm, self._get_custom_dialect_class())
        gw.jvm.org.apache.spark.sql.jdbc.JdbcDialects.registerDialect(
            gw.jvm.dev.caraml.spark.odps.CustomDialect()
        )

        df.write.mode(self._config.save_mode()).format(self.WRITE_FORMAT).option(
            self.OPTION_DRIVER, self.get_jdbc_driver()
        ).option(self.OPTION_URL, self.get_jdbc_url()).option(
            # TODO: use query timeout from config.options() if present
            self.OPTION_QUERY_TIMEOUT,
            self.get_query_timeout(),
        ).option(
            self.OPTION_DB_TABLE, cfg.table()
        ).option(self.OPTION_BATCH_SIZE, self.get_batch_size()).save()
