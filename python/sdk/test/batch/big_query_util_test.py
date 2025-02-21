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

from merlin.batch.big_query_util import bq_valid_column, valid_dataset, valid_table_name, bq_valid_table_id, valid_columns


@pytest.mark.parametrize(
    "dataset_name,expected", [
        pytest.param(
            'dataset', True, id="Should valid for dataset_name which contains of letters"
        ),
        pytest.param(
            '_dataset', True, id="Should valid for dataset_name which begin with underscore and followed by letters"
        ),
        pytest.param(
            '1dataset', True, id="Should valid for dataset_name which begin with number and followed by letters"
        ),
        pytest.param(
            '+dataset', False, id="Should not valid for dataset_name which begin with +"
        ),
        pytest.param(
            'dataset?ok', False, id="Should not valid for dataset_name which contains of question mark"
        ),
        pytest.param(
            'BqUQ30d4Ree6HUK4JvgdtP7Im0diryYuTt8AfT0xt1t2oWM6wBDgRiGqpQpiHDV5Tdw4z8mH8MHT1o8OkztgoGQZPShobIrSnSBzLka9fDUnkOv3Fww0kK9zuHoIUlnV1aAvOfF3hZjbnHL4nmUyrhStA2BFGp5kgdD68iMGe9mlC8gB2dOH1Wq3vn4Zjd7ba6kVJeqUWBKic6zltrBCT07odjSYVGKHXEiF8g9tplPqVXd8Nc4POKVq0AwdohOV4T2HnJ5wJog30ZWV44ixVprj1B9di7Tc0AUiIzIetTLrbTijo0ZyUav7CcjXofRRWkOlCseGW3ytbl7RFZFKYutwoqnOKBXP6YImhMfXib44OequZOiFTz0Lx2Tke0rXFQrt3zGyx11Sgq51D22eseQQuR6w6z7xyIggisZ7zGHIBu4hTEig4Qdi4QGyOmtfObHHheZdFOvGjMisE04tqpd9tJ8gJUEkAN10SD58oADc1noJdgun4ShCagbWhoFyVUOMCaVEOTSwKhpBL0TSiHiCrUCathMgP0EnAx3ONTLNKMWaMhOUwkkU2Nc15WizW9jNUnlHcp0s3p9giKXtLJWyyOJK0poKDsewhsfRexcPMjDf5mrdYlDnFHncYjsXc1TZmeAc6GY7ZuAoFg4v5NrQ9FFEuvwYOiy4YBqijfovqGJDtX9pU9C74BRAzPstZcRvsBr1ayvriZbYNZc76Tw8xsgBVaogGm0L9NTdNsyGycS5Q5uerkIsaxJmWw2fXyurt4hXYdsijZN0GF6itzuPzt48iuyiPWMcAI3kguj3gBtKZZ0v1fzhfSe1MAwlvludHPE0fyQYXsVUY4JxmOjXFdQ5lZHltjs7JzjEa9wD3bHlJEQZFV0hyf5cNefaShwAqe4E4cc3tQ720A9MkC9mC6EichP3aRcd64I0GIUjuekUUvkS8v6YPDPitKyHEEYiqJNK1J0k2ZwPNqXPgtTQTZmvo3GBWmgNeovFx9GwgaQjdfx0cLSdQr46pZRLh',
            False,
            id="Should not valid for dataset_name which length more than 1024"
        )
    ]
)
def test_valid_dataset(dataset_name, expected):
    valid = valid_dataset(dataset_name)
    assert expected == valid


@pytest.mark.parametrize(
    "column_name,expected", [
        pytest.param(
            'column', True, id="Should valid if column_name contains all letters"
        ),
        pytest.param(
            '_column', True, id="Should valid if column_name begins with underscore and followed with letters"
        ),
        pytest.param(
            'Column', True, id="Should valid if column_name contains all letters regardless uppercase or lowercase"
        ),
        pytest.param(
            '1column', False, id="Should not valid if column_name begins with number"
        ),
        pytest.param(
            'column?1', False, id="Should not valid if column_name contains ? or any special characters"
        ),
        pytest.param(
            'column?', False, id="Should not valid if column_name contains ? or any special characters(2)"
        ),
        pytest.param(
            'YKbbvPahCfJQAFw7OQmx2LDfrUOoHcy5jWmJTWomeR3jyy2KdWsCCxUDo6heuOQdQtmHdL4fGt6Bjx8vdz2yZ8WCSLBEVK4jp89HZYhBcogbP2FLoS9ePdVlvWGv7ljJ2',
            False,
            id="Should not valid if column_name length more than 128"
        ),
        pytest.param(
            '_TABLE_1', False, id="Should not valid if column_name has prefix _TABLE"
        ),
        pytest.param(
            '_FILE_abc', False, id="Should not valid if column_name has prefix _FILE_"
        ),
        pytest.param(
            '_PARTITION2', False, id="Should not valid if column_name has prefix _PARTITION"
        ),
    ]
)
def test_valid_column(column_name, expected):
    valid = bq_valid_column(column_name)
    assert expected == valid


@pytest.mark.parametrize(
    "table_name,expected", [
        pytest.param(
            'table', True, id="Should valid if table_name is contains all letters"
        ),
        pytest.param(
            '_table', True, id="Should valid if table_name is begin with underscore and followed by letters"
        ),
        pytest.param(
            '1table', True, id="Should valid if table_name is begin with number and followed by letters"
        ),
        pytest.param(
            '+table', False, id="Should not valid if table_name contains of + or any special characters"
        ),
        pytest.param(
            'table?ok', False, id="Should not valid if table_name contains of ? or any special characters"
        ),
        pytest.param(
            'BqUQ30d4Ree6HUK4JvgdtP7Im0diryYuTt8AfT0xt1t2oWM6wBDgRiGqpQpiHDV5Tdw4z8mH8MHT1o8OkztgoGQZPShobIrSnSBzLka9fDUnkOv3Fww0kK9zuHoIUlnV1aAvOfF3hZjbnHL4nmUyrhStA2BFGp5kgdD68iMGe9mlC8gB2dOH1Wq3vn4Zjd7ba6kVJeqUWBKic6zltrBCT07odjSYVGKHXEiF8g9tplPqVXd8Nc4POKVq0AwdohOV4T2HnJ5wJog30ZWV44ixVprj1B9di7Tc0AUiIzIetTLrbTijo0ZyUav7CcjXofRRWkOlCseGW3ytbl7RFZFKYutwoqnOKBXP6YImhMfXib44OequZOiFTz0Lx2Tke0rXFQrt3zGyx11Sgq51D22eseQQuR6w6z7xyIggisZ7zGHIBu4hTEig4Qdi4QGyOmtfObHHheZdFOvGjMisE04tqpd9tJ8gJUEkAN10SD58oADc1noJdgun4ShCagbWhoFyVUOMCaVEOTSwKhpBL0TSiHiCrUCathMgP0EnAx3ONTLNKMWaMhOUwkkU2Nc15WizW9jNUnlHcp0s3p9giKXtLJWyyOJK0poKDsewhsfRexcPMjDf5mrdYlDnFHncYjsXc1TZmeAc6GY7ZuAoFg4v5NrQ9FFEuvwYOiy4YBqijfovqGJDtX9pU9C74BRAzPstZcRvsBr1ayvriZbYNZc76Tw8xsgBVaogGm0L9NTdNsyGycS5Q5uerkIsaxJmWw2fXyurt4hXYdsijZN0GF6itzuPzt48iuyiPWMcAI3kguj3gBtKZZ0v1fzhfSe1MAwlvludHPE0fyQYXsVUY4JxmOjXFdQ5lZHltjs7JzjEa9wD3bHlJEQZFV0hyf5cNefaShwAqe4E4cc3tQ720A9MkC9mC6EichP3aRcd64I0GIUjuekUUvkS8v6YPDPitKyHEEYiqJNK1J0k2ZwPNqXPgtTQTZmvo3GBWmgNeovFx9GwgaQjdfx0cLSdQr46pZRLh',
            False,
            id="Should not valid if table_name length is more than 1024"
        )
    ]
)
def test_valid_table(table_name, expected):
    valid = valid_table_name(table_name)
    assert expected == valid


@pytest.mark.parametrize(
    "source_table,expected", [
        pytest.param(
            'project.dataset.table', True,
            id="Should valid if source_table contains project_id dataset and table and all components are valid"
        ),
        pytest.param(
            'project-with-dash-and-number-007.dataset.table', True,
            id="Should valid if source_table contains project_id which contain number and dash"
        ),
        pytest.param(
            'project. dataset.table', False, id="Should not valid if source_table contains invalid dataset"
        ),
        pytest.param(
            'project.dataset .table', False, id="Should not valid if source_table contains invalid dataset(2)"
        ),
        pytest.param(
            'project.dataset?.table', False, id="Should not valid if source_table contains invalid dataset(3)"
        ),
        pytest.param(
            'project.dataset. table', False, id="Should not valid if source_table contains invalid table"
        ),
        pytest.param(
            'project.dataset.table ', False, id="Should not valid if source_table contains invalid table(2)"
        ),
        pytest.param(
            'project.dataset.table?table', False, id="Should not valid if source_table contains invalid table(3)"
        ),
        pytest.param(
            '.dataset.table', False, id="Should not valid if source_table doesnt have project_id info"
        ),
        pytest.param(
            '    abc.dataset.table', False, id="Should not valid if source_table contains invalid project_id"
        )
    ]
)
def test_valid_source_table(source_table, expected):
    valid = bq_valid_table_id(source_table)
    assert expected == valid


@pytest.mark.parametrize(
    "columns,expected", [
        pytest.param(
            ['column1', 'column2'], True, id="Should valid if all columns contains of letters"
        ),
        pytest.param(
            ['coLumn1', 'column2'], True, id="Should valid if all columns contains of letters(2)"
        ),
        pytest.param(
            ['?column1', 'column2'], False, id="Should valid if one of the columns contains invalid column"
        ),
        pytest.param(
            ['column1', '12column2'], False, id="Should valid if one of the columns contains invalid column"
        ),
    ]
)
def test_valid_columns(columns, expected):
    valid = valid_columns(columns)
    assert expected == valid
