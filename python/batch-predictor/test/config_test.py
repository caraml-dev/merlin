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

import pytest
from pyspark.sql.types import ArrayType, DoubleType, FloatType, IntegerType, \
    LongType, StringType

from merlinpyspark.config import BigQuerySinkConfig, BigQuerySourceConfig, load
from test.util.test_utils import write_config

sample_1_yaml = """
kind: PredictionJob
version: v1
name: sample-1
bigquerySource:
  table: 'bigquery-public-data:samples.shakespeare'
  features:
    - word
    - word_count
  options:
    filter: "word_count > 100"
model:
  type: PYFUNC_V2
  uri: "gs://bucket-name/mlflow/6/2c3703fbbf9f4866b26e4cf91641f02c/artifacts/model"
  result:
    type: ARRAY
    item_type: DOUBLE
bigquerySink:
  table: 'project.dataset.table'
  result_column: 'prediction'
  save_mode: OVERWRITE
  options:
    project: "project"
    temporaryGcsBucket: "bucket-name"
"""

sample_2_yaml = """
kind: PredictionJob
version: v1
name: sample-2
bigquerySource:
  table: 'bigquery-public-data:samples.shakespeare'
model:
  type: PYFUNC_V2
  uri: "gs://bucket-name/mlflow/6/2c3703fbbf9f4866b26e4cf91641f02c/artifacts/model"
bigquerySink:
  table: 'project.dataset.table'
"""

test_type_yaml = """
kind: PredictionJob
version: v1
name: test-type
bigquerySource:
  table: 'bigquery-public-data:samples.shakespeare'
model:
  type: PYFUNC_V2
  uri: "gs://bucket-name/mlflow/6/2c3703fbbf9f4866b26e4cf91641f02c/artifacts/model"
  result:
    type: {}
    item_type: {}
bigquerySink:
  table: 'project.dataset.table'
"""


@pytest.fixture
def sample_1_path():
    return write_config("test-config/sample_1.yaml", sample_1_yaml)


@pytest.fixture
def sample_2_path():
    return write_config("test-config/sample_2.yaml", sample_2_yaml)


def test_load_1(sample_1_path):
    cfg = load(sample_1_path)
    assert cfg is not None
    assert cfg.model() is not None
    assert cfg.source() is not None
    assert cfg.sink() is not None

    src_cfg = cfg.source()
    assert src_cfg.source_type() == "bigquery"
    assert src_cfg.options()["filter"] == "word_count > 100"
    assert len(src_cfg.features()) == 2
    assert src_cfg.features()[0] == "word"
    assert src_cfg.features()[1] == "word_count"
    assert isinstance(src_cfg, BigQuerySourceConfig)
    assert src_cfg.table() == "bigquery-public-data:samples.shakespeare"

    model_cfg = cfg.model()
    assert model_cfg.model_type() == "PYFUNC_V2"
    assert model_cfg.result_type() == ArrayType(DoubleType())
    assert model_cfg.model_uri() == "gs://bucket-name/mlflow/6/2c3703fbbf9f4866b26e4cf91641f02c/artifacts/model"

    sink_cfg = cfg.sink()
    assert sink_cfg.sink_type() == "bigquery"
    assert sink_cfg.save_mode() == "overwrite"
    assert sink_cfg.result_column() == "prediction"
    assert sink_cfg.options()["temporaryGcsBucket"] == "bucket-name"
    assert isinstance(sink_cfg, BigQuerySinkConfig)
    assert sink_cfg.table() == "project.dataset.table"


def test_load_2(sample_2_path):
    cfg = load(sample_2_path)
    assert cfg.model() is not None
    assert cfg.source() is not None
    assert cfg.sink() is not None

    src_cfg = cfg.source()
    assert src_cfg.source_type() == "bigquery"
    assert len(src_cfg.options()) == 0
    assert len(src_cfg.features()) == 0
    assert isinstance(src_cfg, BigQuerySourceConfig)
    assert src_cfg.table() == "bigquery-public-data:samples.shakespeare"

    model_cfg = cfg.model()
    assert model_cfg.model_type() == "PYFUNC_V2"
    assert model_cfg.result_type() == DoubleType()
    assert model_cfg.model_uri() == "gs://bucket-name/mlflow/6/2c3703fbbf9f4866b26e4cf91641f02c/artifacts/model"

    sink_cfg = cfg.sink()
    assert sink_cfg.sink_type() == "bigquery"
    assert sink_cfg.save_mode() == "errorifexists"
    assert sink_cfg.result_column() == "prediction"
    assert len(sink_cfg.options()) == 0
    assert isinstance(sink_cfg, BigQuerySinkConfig)
    assert sink_cfg.table() == "project.dataset.table"


@pytest.mark.parametrize(
    "return_type,item_type,expected",
    [
        ("FLOAT", "", FloatType()),
        ("DOUBLE", "", DoubleType()),
        ("STRING", "", StringType()),
        ("INTEGER", "", IntegerType()),
        ("LONG", "", LongType()),
        ("ARRAY", "FLOAT", ArrayType(FloatType())),
        ("ARRAY", "DOUBLE", ArrayType(DoubleType())),
        ("ARRAY", "STRING", ArrayType(StringType())),
        ("ARRAY", "INTEGER", ArrayType(IntegerType())),
        ("ARRAY", "LONG", ArrayType(LongType())),
    ]
)
def test_model_result_conversion(return_type, item_type, expected):
    cfg_file = write_config("test-config/test_type.yaml",
                            test_type_yaml.format(
                                return_type,
                                item_type))

    cfg = load(cfg_file)
    assert cfg.model().result_type() == expected


