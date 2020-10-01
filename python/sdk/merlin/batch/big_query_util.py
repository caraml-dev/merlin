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

import re

GCP_PROJECT_ID_EXPRESSION = r'^[a-z]([-a-z0-9]*[a-z0-9])?'
WORD_CHARACTER_EXPRESSION = r'^\w+$'
DEFAULT_CHARACTER_LIMIT = 1024
COLUMN_NAME_PREFIX_EXCLUSIONS = ['_TABLE_', '_FILE_', '_PARTITION']


def valid_dataset(dataset: str) -> bool:
    """
    Validate BigQuery dataset name

    :param dataset: BigQuery dataset name
    :return: boolean

    Rules based on this page https://cloud.google.com/bigquery/docs/datasets#dataset-naming
    * May contain up to 1,024 characters
    * Can contain letters (upper or lower case), numbers, and underscores
    """
    return validate_text(dataset, WORD_CHARACTER_EXPRESSION, DEFAULT_CHARACTER_LIMIT)


def valid_column(column_name: str) -> bool:
    """
    Validate BigQuery column name

    :param column_name: BigQuery column name
    :return: boolean

    Rules based on this page https://cloud.google.com/bigquery/docs/schemas#column_names
    * A column name must contain only letters (a-z, A-Z), numbers (0-9), or underscores (_)
    * It must start with a letter or underscore
    * Maximum length 128
    """
    for prefix in COLUMN_NAME_PREFIX_EXCLUSIONS:
        if column_name.startswith(prefix):
            return False

    column_name_max_length = 128
    pattern = r'^[a-zA-Z_]\w*$'
    return validate_text(column_name, pattern, column_name_max_length)


def valid_table_name(table_name: str) -> bool:
    """
    Validate BigQuery table name

    :param table_name: BigQuery table name
    :return: boolean

    Rules based on this page https://cloud.google.com/bigquery/docs/tables#table_naming
    * A table name must contain only letters (a-z, A-Z), numbers (0-9), or underscores (_)
    * Maximum length 1024
    """
    return validate_text(table_name, WORD_CHARACTER_EXPRESSION, DEFAULT_CHARACTER_LIMIT)


def validate_text(text: str, pattern: str, max_length: int) -> bool:
    """
    Validate text based on regex pattern and maximum length allowed

    :param text: Text to validate
    :param pattern: Regular expression pattern to validate text
    :param max_length: Maximum length allowed
    :return: boolean
    """
    if len(text) > max_length:
        return False
    if re.search(pattern, text):
        return True
    return False


def valid_table_id(table_id: str) -> bool:
    """
    Validate BigQuery source_table which satisfied this format project_id.dataset.table

    :param table_id: Source table
    :return: boolean
    """
    components = table_id.split(".")
    if len(components) != 3:
        return False

    project_id = components[0]
    dataset = components[1]
    table = components[2]
    if not validate_text(project_id, GCP_PROJECT_ID_EXPRESSION, DEFAULT_CHARACTER_LIMIT):
        return False
    if not valid_dataset(dataset):
        return False
    if not valid_table_name(table):
        return False
    return True


def valid_columns(columns) -> bool:
    """
    Validate multiple BiqQuery columns

    :param columns: List of columns
    :return: boolean
    """
    for column in columns:
        if not valid_column(column):
            return False
    return True
