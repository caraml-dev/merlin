import re
from merlin.batch.big_query_util import validate_text


def mc_valid_table_id(table_id: str) -> bool:
    """
    Validate MaxCompute table which satisfied this format project.schema.table

    Ref: 
    * https://www.alibabacloud.com/help/en/maxcompute/user-guide/manage-projects-in-the-new-maxcompute-console
    * https://www.alibabacloud.com/help/en/maxcompute/user-guide/table-operations#section-ymq-wfc-2bg


    :param table_id: table
    :return: boolean
    """
    components = table_id.split(".")
    if len(components) != 3:
        return False

    project_id = components[0]
    _ = components[1]
    table = components[2]
    # TODO: check regex for valid project and schema
    # for now, just validate the table name
    # validate project name
    if not validate_text(project_id, r'^[a-zA-Z][a-zA-Z0-9_]*$', 28, 3):
        return False
    if not validate_text(table, r'^[a-zA-Z][a-zA-Z0-9_]*$', 128):
        return False
    return True

def mc_valid_column(column_name: str) -> bool:
    """
    Validate MaxCompute column

    :param column_name: column name
    :return: boolean
    """
    column_name_max_length = 128
    pattern = r'^[a-zA-Z_]\w*$'
    return validate_text(column_name, pattern, column_name_max_length)

def mc_valid_columns(columns) -> bool:
    """
    Validate MaxCompute columns

    :param columns: columns
    :return: boolean
    """
    for column in columns:
        if not mc_valid_column(column):
            return False
    return True