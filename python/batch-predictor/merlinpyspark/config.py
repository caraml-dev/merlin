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
from typing import Iterable, MutableMapping, Dict

import yaml
from google.protobuf import json_format
from pyspark.sql.types import ArrayType, DataType, DoubleType, FloatType, \
    IntegerType, LongType, StringType

from merlinpyspark.spec.prediction_job_pb2 import BigQuerySink, BigQuerySource, MaxComputeSink, MaxComputeSource, Model, \
    PredictionJob, ResultType, ModelType, SaveMode


def printJobConfig(spec_path):
    with open(spec_path, "r") as f:
        print(f"""
===============================================================
Prediction job configuration:
{f.read()}
===============================================================
        """)


def load(spec_path: str) -> "JobConfig":
    """
    Read a job spec file (yaml file) and return nested dictionary
    containing the job config
    :param spec_path: path to yaml config file
    :return: JobConfig object
    """
    printJobConfig(spec_path)
    prediction_job = PredictionJob()
    with open(spec_path, "r") as f:
        return JobConfig(
            json_format.ParseDict(yaml.safe_load(f), prediction_job))


class JobConfig:
    BIGQUERY_SOURCE = "bigquery_source"
    BIGQUERY_SINK = "bigquery_sink"
    GCS_SOURCE = "gcs_source"
    GCS_SINK = "gcs_sink"
    MAXCOMPUTE_SOURCE = "maxcompute_source"
    MAXCOMPUTE_SINK = "maxcompute_sink"

    def __init__(self, config: PredictionJob):
        self._config = config

    def source(self) -> "SourceConfig":
        src = self._config.WhichOneof("source")
        if src is None:
            raise ValueError("missing mandatory source config")

        if src == self.BIGQUERY_SOURCE:
            return BigQuerySourceConfig(self._config.bigquery_source)
        elif src == self.MAXCOMPUTE_SOURCE:
            return MaxComputeSourceConfig(self._config.maxcompute_source)
        else:
            raise ValueError(f"{src} is not implemented")

    def model(self) -> "ModelConfig":
        return ModelConfig(self._config.model)

    def sink(self) -> "SinkConfig":
        sink = self._config.WhichOneof("sink")
        if sink is None:
            raise ValueError("missing mandatory sink config")

        if sink == self.BIGQUERY_SINK:
            return BigQuerySinkConfig(self._config.bigquery_sink)
        elif sink == self.MAXCOMPUTE_SINK:
            return MaxComputeSinkConfig(self._config.maxcompute_sink)
        else:
            raise ValueError(f"{sink} is not implemented")


class SourceConfig(ABC):
    @abstractmethod
    def source_type(self) -> str:
        pass

    @abstractmethod
    def options(self) -> MutableMapping[str, str]:
        pass

    @abstractmethod
    def features(self) -> Iterable[str]:
        pass


class BigQuerySourceConfig(SourceConfig):
    TYPE = "bigquery"

    def __init__(self, bq_src_proto: BigQuerySource):
        self._proto = bq_src_proto

    def source_type(self) -> str:
        return self.TYPE

    def options(self) -> MutableMapping[str, str]:
        return self._proto.options

    def features(self) -> Iterable[str]:
        return self._proto.features

    def table(self) -> str:
        return self._proto.table

class MaxComputeSourceConfig(SourceConfig):
    TYPE = "maxcompute"

    def __init__(self, mc_src_proto: MaxComputeSource):
        self._proto = mc_src_proto
        self._project, self._schema, self._table = self._proto.table.split(".")
        
    def source_type(self) -> str:
        return self.TYPE

    def options(self) -> MutableMapping[str, str]:
        return self._proto.options

    def features(self) -> Iterable[str]:
        return self._proto.features

    def table(self) -> str:
        return self._table
    
    def project(self) -> str:
        return self._project
    
    def schema(self) -> str:
        return self._schema
        
    def endpoint(self) -> str:
        return self._proto.endpoint

class ModelConfig:
    PRIMITIVE_TYPE_MAP = {
        ResultType.DOUBLE: DoubleType(),
        ResultType.FLOAT: FloatType(),
        ResultType.INTEGER: IntegerType(),
        ResultType.LONG: LongType(),
        ResultType.STRING: StringType(),
    }

    def __init__(self, model_proto: Model):
        self._proto = model_proto

    def model_type(self) -> str:
        return ModelType.Name(self._proto.type)

    def model_uri(self) -> str:
        return self._proto.uri

    def result_type(self) -> DataType:
        rt = self._proto.result.type
        if rt == ResultType.ARRAY:
            it = self._proto.result.item_type
            if it in self.PRIMITIVE_TYPE_MAP:
                return ArrayType(self.PRIMITIVE_TYPE_MAP[it])
            raise ValueError(f"unknown item type for array: {it}")

        if rt in self.PRIMITIVE_TYPE_MAP:
            return self.PRIMITIVE_TYPE_MAP[rt]

        raise ValueError(f"unknown result type: {rt}")


class SinkConfig(ABC):
    @abstractmethod
    def sink_type(self) -> str:
        pass

    @abstractmethod
    def result_column(self) -> str:
        pass

    @abstractmethod
    def options(self) -> MutableMapping[str, str]:
        pass

    @abstractmethod
    def save_mode(self) -> str:
        pass


class BigQuerySinkConfig(SinkConfig):
    TYPE = "bigquery"
    DEFAULT_RESULT_COLUMN = "prediction"
    
    def __init__(self, bq_sink_proto: BigQuerySink):
        self._proto = bq_sink_proto

    def sink_type(self) -> str:
        return self.TYPE

    def result_column(self) -> str:
        col = self._proto.result_column
        if col == "":
            return self.DEFAULT_RESULT_COLUMN
        return col

    def options(self) -> MutableMapping[str, str]:
        return self._proto.options

    def save_mode(self) -> str:
        return SaveMode.Name(self._proto.save_mode).lower()

    def table(self) -> str:
        return self._proto.table

    def staging_bucket(self) -> str:
        return self._proto.staging_bucket


class MaxComputeSinkConfig(SinkConfig):
    TYPE = "maxcompute"
    DEFAULT_RESULT_COLUMN = "prediction"

    def __init__(self, mc_sink_proto: MaxComputeSink):
        self._proto = mc_sink_proto
        self._project, self._schema, self._table = self._proto.table.split(".")
    
    def sink_type(self) -> str:
        return self.TYPE
    
    def result_column(self) -> str:
        col = self._proto.result_column
        if col == "":
            return self.DEFAULT_RESULT_COLUMN
        return col
    
    def options(self) -> MutableMapping[str, str]:
        return self._proto.options
    
    def save_mode(self) -> str:
        return SaveMode.Name(self._proto.save_mode).lower()
    
    def table(self) -> str:
        return self._table

    def project(self) -> str:
        return self._project
    
    def schema(self) -> str:
        return self._schema

    def endpoint(self) -> str:
        return self._proto.endpoint