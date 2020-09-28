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

from merlin.batch.source import BigQuerySource


@pytest.mark.parametrize(
    "table,features,options,expected_dict,expected_valid", [
        pytest.param(
            'project.dataset.table',
            ['driver_id', 'completed_orders'],
            {'project': 'project'},
            {
                'table': 'project.dataset.table',
                'features': ['driver_id', 'completed_orders'],
                'options': {'project': 'project'}
            },
            True,
            id="Should valid if source has valid_types and valid property values"
        ),
        pytest.param(
            'project.dataset.table',
            ['driver_id', 'completed_orders'],
            {},
            {
                'table': 'project.dataset.table',
                'features': ['driver_id', 'completed_orders'],
                'options': {}
            },
            True,
            id="Should valid if source has valid_types and valid property value(2)"
        ),
        pytest.param(
            'project.dataset.table',
            ['driver_id', 'completed_orders'],
            None,
            {
                'table': 'project.dataset.table',
                'features': ['driver_id', 'completed_orders'],
                'options': {}
            },
            True,
            id="Should valid if source has valid_types and valid property value(3)"
        ),
        pytest.param(
            'project.dataset',
            ['driver_id', 'completed_orders'],
            None,
            {},
            False,
            id="Should not valid and return empty dictionary if source table is not valid"
        ),
        pytest.param(
            1,
            ['driver_id', 'completed_orders'],
            None,
            {},
            False,
            id="Should not valid and return empty dictionary if source table doesn't have valid data type, it should be string"
        ),
        pytest.param(
            'project.dataset.table',
            [1, 2, 3],
            None,
            {},
            False,
            id="Should not valid and return empty dictionary if features doesn't have valid data type, it should be array of string"
        ),
        pytest.param(
            'project.dataset.table',
            'driver_id',
            None,
            {},
            False,
            id="Should not valid and return empty dictionary if features doesn't have valid data type, it should be array of string(2)"
        )
    ]
)
def test_dictionary_generated(table, features, options, expected_dict, expected_valid):
    biq_query_source = BigQuerySource(table, features, options)
    if expected_valid:
        assert biq_query_source.to_dict() == expected_dict
    else:
        with pytest.raises(ValueError):
            biq_query_source.to_dict()


@pytest.mark.parametrize(
    "table,features,options,expected", [
        pytest.param(
            'project.dataset.table', ['driver_id', 'completed_orders'], {'project': 'project'}, True,
            id="Should valid if source has valid_types and valid property value"
        ),
        pytest.param(
            'project.dataset.table', ['driver_id', 'completed_orders'], None, True,
            id="Should valid if source has valid_types and valid property value(2)"
        ),
        pytest.param(
            'project.dataset.table', ['driver_id', 'completed_orders'], ['project', 'environment'], False,
            id="Should not valid if options has wrong data type, it should be dictionary"
        ),
        pytest.param(
            'project', ['driver_id', 'completed_orders'], {'project': 'project'}, False,
            id="Should not valid if source_table is not valid"
        ),
        pytest.param(
            'project.dataset', ['driver_id', 'completed_orders'], {'project': 'project'}, False,
            id="Should not valid if source_table is not valid(2)"
        ),
        pytest.param(
            'project.dataset.table', [1, 2, 3], {'project': 'project'}, False,
            id="Should not valid if features doesn't have correct data types, it should be array of string"
        ),
        pytest.param(
            'project.dataset.table', 'driver_id', {'project': 'project'}, False,
            id="Should not valid if features doesn't have correct data types, it should be array of string"
        ),
        pytest.param(
            'project.dataset.table', 14, {'project': 'project'}, False,
            id="Should not valid, features doesn't have correct data types, it should be array of string"
        ),
        pytest.param(
            'project.dataset.table', ['driver_id', '_TABLE_driver'], {'project': 'project'}, False,
            id="Should not valid if features contains invalid column name"
        ),
    ]
)
def test_valid(table, features, options, expected):
    biq_query_source = BigQuerySource(table, features, options)
    if expected:
        biq_query_source._validate()
    else:
        with pytest.raises(ValueError):
            biq_query_source._validate()
