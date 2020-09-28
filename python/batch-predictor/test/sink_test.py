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

import pandas as pd
import pytest

from merlinpyspark.config import BigQuerySinkConfig
from merlinpyspark.sink import create_sink
from merlinpyspark.spec.prediction_job_pb2 import BigQuerySink


@pytest.mark.ci
def test_bq_sink(spark_session, bq):
    bq_sink_proto = BigQuerySink(table='project.dataset.table',
                                 result_column="prediction",
                                 staging_bucket="bucket-name",
                                 options={"project": "project"})

    sink_cfg = BigQuerySinkConfig(bq_sink_proto)
    table_id = sink_cfg.table()

    bq.delete_table(table=table_id, not_found_ok=True)

    test_data = {
        "word": ["lorem", "ipsum"],
        "word_count": [1, 2]
    }
    pdf = pd.DataFrame(test_data)
    df = spark_session.createDataFrame(pdf)

    bq_sink = create_sink(sink_cfg)
    bq_sink.save(df)

    t = bq.get_table(table_id)
    assert len(t.schema) == 2
    assert t.schema[0].name == 'word'
    assert t.schema[0].field_type == 'STRING'
    assert t.schema[1].name == 'word_count'
    assert t.schema[1].field_type == 'INTEGER'
