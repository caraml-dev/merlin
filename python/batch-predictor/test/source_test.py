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

from merlinpyspark.config import BigQuerySourceConfig
from merlinpyspark.source import create_source
from merlinpyspark.spec.prediction_job_pb2 import BigQuerySource
import pytest

@pytest.mark.ci
def test_bq_source(spark_session):
    bq_src_proto = BigQuerySource(
        table="bigquery-public-data:samples.shakespeare",
        features=["word", "word_count"],
        options={"filter": "word_count > 100"}
    )
    src_cfg = BigQuerySourceConfig(bq_src_proto)

    bq_source = create_source(spark_session, src_cfg)
    df = bq_source.load()
    assert len(df.columns) == 2
    assert df.columns[0] == "word"
    assert df.columns[1] == "word_count"
    assert df.sort("word_count").take(1)[0]["word_count"] > 100
