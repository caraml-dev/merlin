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

from merlinpyspark.config import SinkConfig, BigQuerySinkConfig


def create_sink(sink_config: SinkConfig) -> "Sink":
    if sink_config.sink_type() == BigQuerySinkConfig.TYPE:
        if not isinstance(sink_config, BigQuerySinkConfig):
            raise ValueError("sink_config is not BigQuerySink")

        return BigQuerySink(sink_config)

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
        df.write \
            .mode(self._config.save_mode()) \
            .format(self.WRITE_FORMAT) \
            .option(self.OPTION_TABLE, self._config.table()) \
            .option(self.OPTION_STAGING_BUCKET, self._config.staging_bucket()) \
            .options(**self._config.options()) \
            .save()
