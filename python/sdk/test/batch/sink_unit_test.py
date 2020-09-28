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

from merlin.batch.sink import BigQuerySink, SaveMode


@pytest.mark.parametrize(
    "table,staging_bucket,result_column,save_mode,options,expected_dict,expected_valid", [
        pytest.param(
            'project.dataset.table',
            'bucket',
            'column',
            SaveMode.APPEND,
            {'project': 'project'},
            {
                'table': 'project.dataset.table',
                'staging_bucket': 'bucket',
                'result_column': 'column',
                'save_mode': SaveMode.APPEND.value,
                'options': {'project': 'project'}
            },
            True,
            id="Should return correct dictionary and valid if sink has valid_types and valid property values"
        ),
        pytest.param(
            'project.dataset.table',
            'bucket',
            'column',
            SaveMode.APPEND,
            {},
            {
                'table': 'project.dataset.table',
                'staging_bucket': 'bucket',
                'result_column': 'column',
                'save_mode': SaveMode.APPEND.value,
                'options': {}
            },
            True,
            id="Should return correct dictionary and valid if sink has valid_types and valid property values(2)"
        ),
        pytest.param(
            'project.dataset.table',
            'bucket',
            'column',
            SaveMode.APPEND,
            None,
            {
                'table': 'project.dataset.table',
                'staging_bucket': 'bucket',
                'result_column': 'column',
                'save_mode': SaveMode.APPEND.value,
                'options': {}
            },
            True,
            id="Should return correct dictionary and valid if sink has valid_types and valid property values(3)"
        ),
        pytest.param(
            'project',
            'bucket',
            'column',
            SaveMode.APPEND,
            None,
            {},
            False,
            id="Should return empty dictionary and not valid if table is invalid"
        ),
        pytest.param(
            'project.dataset.table',
            'bucket',
            '_TABLE_abc',
            SaveMode.APPEND,
            None,
            {},
            False,
            id="Should return empty dictionary and not valid if result_column is invalid"
        )
    ]
)
def test_dictionary_generated(table, staging_bucket, result_column, save_mode, options, expected_dict, expected_valid):
    bq_sink = BigQuerySink(table, staging_bucket, result_column, save_mode, options)
    if expected_valid:
        assert bq_sink.to_dict() == expected_dict
    else:
        with pytest.raises(ValueError):
            bq_sink.to_dict()


@pytest.mark.parametrize(
    "table,staging_bucket,result_column,save_mode,options,expected", [
        pytest.param(
            'project.dataset.table',
            'bucket',
            'column',
            SaveMode.APPEND,
            {'project': 'project'},
            True,
            id="Should valid if sink has valid types and valid property values"
        ),
        pytest.param(
            'project.dataset.table',
            1,
            'column',
            SaveMode.APPEND,
            {},
            False,
            id="Should not valid if staging_bucket has wrong data type, it should be string"
        ),
        pytest.param(
            'project.dataset.table',
            'bucket',
            ['column'],
            SaveMode.APPEND,
            {'project': 'project'},
            False,
            id="Should not valid if result_column has wrong data type, it should be string"
        ),
        pytest.param(
            'project.dataset.table',
            'bucket',
            '_PARTITION_result',
            SaveMode.APPEND,
            None,
            False,
            id="Should not valid if result_column is invalid"
        ),
        pytest.param(
            'project.dataset.table',
            'bucket',
            '_TABLE_abc',
            SaveMode.APPEND,
            None,
            False,
            id="Should not valid if result_column is invalid(2)"
        ),
    ]
)
def test_valid(table, staging_bucket, result_column, save_mode, options, expected):
    bq_sink = BigQuerySink(table, staging_bucket, result_column, save_mode, options)
    if expected:
        bq_sink._validate()
    else:
        with pytest.raises(ValueError):
            bq_sink._validate()
