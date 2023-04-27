import React, { Fragment } from "react";
import {
  EuiSwitch,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiPanel,
  EuiRadio,
  EuiSpacer,
  EuiText
} from "@elastic/eui";
import { DraggableHeader } from "../../../DraggableHeader";
import { FromJson } from "../../../../../../../../services/transformer/TransformerConfig";
import { TableColumnsInput } from "./TableColumnsInput";
import { TableFromFileSchema } from "./TableFromFileSchema";
import { FormLabelWithToolTip, get } from "@caraml-dev/ui-lib";

export const TableInputCard = ({
  index = 0,
  tableIdx = 0,
  table,
  onChangeHandler,
  onColumnChangeHandler,
  onDelete,
  errors = {},
  ...props
}) => {
  const onChange = (field, value) => {
    if (JSON.stringify(value) !== JSON.stringify(table[field])) {
      onChangeHandler({
        ...table,
        [field]: value
      });
    }
  };

  return (
    <EuiPanel>
      <DraggableHeader
        onDelete={onDelete}
        dragHandleProps={props.dragHandleProps}
      />

      <EuiSpacer size="s" />

      <EuiFlexGroup direction="column" gutterSize="m">
        <EuiFlexItem>
          <EuiText size="s">
            <h4>Generic Table</h4>
          </EuiText>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFormRow
            label="Table Name *"
            isInvalid={!!errors.name}
            error={errors.name}
            display="columnCompressed"
            fullWidth>
            <EuiFieldText
              placeholder="Table name"
              value={table.name}
              onChange={e => onChange("name", e.target.value)}
              isInvalid={!!errors.name}
              name={`table-name-${index}`}
              fullWidth
            />
          </EuiFormRow>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFormRow
            label="Base Table *"
            isInvalid={!!errors.source}
            error={errors.source}
            display="columnCompressed"
            fullWidth>
            <EuiFlexGroup>
              <EuiFlexItem grow={false}>
                <EuiRadio
                  id={`table-input-none-${index}-${tableIdx}`}
                  label="None"
                  checked={table.baseTable === undefined}
                  onChange={() => {
                    onChangeHandler({
                      ...table,
                      baseTable: undefined,
                      columns: []
                    });
                  }}
                />
              </EuiFlexItem>
              <EuiFlexItem grow={false}>
                <EuiRadio
                  id={`table-input-fromTable-${index}-${tableIdx}`}
                  label="From Table"
                  checked={
                    (table.baseTable && !!table.baseTable.fromTable) || false
                  }
                  onChange={() =>
                    onChangeHandler({
                      ...table,
                      baseTable: { fromTable: { tableName: "" } },
                      columns: []
                    })
                  }
                />
              </EuiFlexItem>
              <EuiFlexItem grow={false}>
                <EuiRadio
                  id={`table-input-fromFile-${index}-${tableIdx}`}
                  label="From File"
                  checked={
                    (table.baseTable && !!table.baseTable.fromFile) || false
                  }
                  onChange={() =>
                    onChangeHandler({
                      ...table,
                      baseTable: {
                        fromFile: { format: "", uri: "", schema: [] }
                      },
                      columns: undefined
                    })
                  }
                />
              </EuiFlexItem>
              <EuiFlexItem>
                <EuiRadio
                  id={`table-input-fromJson-${index}-${tableIdx}`}
                  label="From JSON"
                  checked={
                    (table.baseTable && !!table.baseTable.fromJson) || false
                  }
                  onChange={() =>
                    onChangeHandler({
                      ...table,
                      baseTable: { fromJson: new FromJson() },
                      columns: []
                    })
                  }
                />
              </EuiFlexItem>
            </EuiFlexGroup>
          </EuiFormRow>
        </EuiFlexItem>

        {table.baseTable === undefined && (
          <EuiFlexItem>
            <EuiFormRow label="Columns *" fullWidth>
              <TableColumnsInput
                columns={table.columns || []}
                onChangeHandler={onColumnChangeHandler}
                errors={errors.columns}
              />
            </EuiFormRow>
          </EuiFlexItem>
        )}

        {table.baseTable && table.baseTable.fromTable && (
          <EuiFlexItem>
            <EuiFormRow
              label="Base Table Name *"
              isInvalid={!!get(errors, "baseTable.fromTable.tableName")}
              error={get(errors, "baseTable.fromTable.tableName")}
              display="columnCompressed"
              fullWidth>
              <Fragment>
                <EuiFieldText
                  placeholder="Base Table Name"
                  value={table.baseTable.fromTable.tableName}
                  onChange={e =>
                    onChange("baseTable", {
                      fromTable: {
                        ...table.baseTable.fromTable,
                        tableName: e.target.value
                      }
                    })
                  }
                  isInvalid={!!errors.name}
                  name={`base-table-name-${index}`}
                  fullWidth
                />
              </Fragment>
            </EuiFormRow>

            <EuiFormRow label="Columns *" fullWidth>
              <TableColumnsInput
                columns={table.columns || []}
                onChangeHandler={onColumnChangeHandler}
                errors={errors.columns}
              />
            </EuiFormRow>
          </EuiFlexItem>
        )}

        {table.baseTable && table.baseTable.fromFile && (
          <EuiFlexItem>
            <EuiFormRow
              label="File Type *"
              isInvalid={!!get(errors, "baseTable.fromFile.format")}
              error={get(errors, "baseTable.fromFile.format")}
              display="columnCompressed"
              fullWidth>
              <EuiFlexGroup>
                <EuiFlexItem grow={false}>
                  <EuiRadio
                    id={`table-file-format-csv-${index}-${tableIdx}`}
                    label="CSV"
                    checked={
                      (table.baseTable &&
                        table.baseTable.fromFile &&
                        table.baseTable.fromFile.format === "CSV") ||
                      false
                    }
                    onChange={e =>
                      onChange("baseTable", {
                        fromFile: {
                          ...table.baseTable.fromFile,
                          format: "CSV"
                        }
                      })
                    }
                  />
                </EuiFlexItem>
                <EuiFlexItem>
                  <EuiRadio
                    id={`table-file-format-parquet-${index}-${tableIdx}`}
                    label="Parquet"
                    checked={
                      (table.baseTable &&
                        table.baseTable.fromFile &&
                        table.baseTable.fromFile.format === "PARQUET") ||
                      false
                    }
                    onChange={e =>
                      onChange("baseTable", {
                        fromFile: {
                          ...table.baseTable.fromFile,
                          format: "PARQUET"
                        }
                      })
                    }
                  />
                </EuiFlexItem>
              </EuiFlexGroup>
            </EuiFormRow>
            <EuiFormRow
              label={
                <FormLabelWithToolTip
                  label="URI to file *"
                  content="Global GCS path e.g. gs://my-bucket/my-csv.csv OR
                    local path relative to model storage e.g. myfiles/myparquet.parquet"
                />
              }
              isInvalid={!!get(errors, "baseTable.fromFile.uri")}
              error={get(errors, "baseTable.fromFile.uri")}
              display="columnCompressed"
              fullWidth>
              <Fragment>
                <EuiFieldText
                  placeholder="URI to file"
                  value={table.baseTable.fromFile.uri}
                  onChange={e =>
                    onChange("baseTable", {
                      fromFile: {
                        ...table.baseTable.fromFile,
                        uri: e.target.value
                      }
                    })
                  }
                  isInvalid={!!errors.name}
                  name={`base-table-uri-${index}`}
                  fullWidth
                />
              </Fragment>
            </EuiFormRow>
            <EuiFormRow
              label={
                <FormLabelWithToolTip
                  label="Schema (optional)"
                  content="Column type is auto-detected by default. Create schema below to manually define the type for specific column."
                />
              }
              fullWidth>
              <TableFromFileSchema
                columns={table.baseTable.fromFile.schema || []}
                onChangeHandler={onChangeHandler}
                errors={errors.schema}
              />
            </EuiFormRow>
          </EuiFlexItem>
        )}

        {table.baseTable && table.baseTable.fromJson && (
          <Fragment>
            <EuiFlexItem>
              <EuiFormRow
                label="JSONPath *"
                isInvalid={!!get(errors, "baseTable.fromJson.jsonPath")}
                error={get(errors, "baseTable.fromJson.jsonPath")}
                display="columnCompressed"
                fullWidth>
                <EuiFieldText
                  placeholder="JSONPath"
                  value={table.baseTable.fromJson.jsonPath}
                  onChange={e =>
                    onChange("baseTable", {
                      fromJson: {
                        ...table.baseTable.fromJson,
                        jsonPath: e.target.value
                      }
                    })
                  }
                  isInvalid={!!errors.name}
                  name={`table-name-${index}`}
                  fullWidth
                />
              </EuiFormRow>
            </EuiFlexItem>

            <EuiFlexItem>
              <EuiFormRow
                label="Row Number"
                display="columnCompressed"
                fullWidth>
                <EuiSwitch
                  id={`addRowNumber-${index}`}
                  label="Add row number"
                  checked={table.baseTable.fromJson.addRowNumber}
                  onChange={e =>
                    onChange("baseTable", {
                      fromJson: {
                        ...table.baseTable.fromJson,
                        addRowNumber: e.target.checked
                      }
                    })
                  }
                />
              </EuiFormRow>

              <EuiFormRow label="Columns *" fullWidth>
                <TableColumnsInput
                  columns={table.columns || []}
                  onChangeHandler={onColumnChangeHandler}
                  errors={errors.columns}
                />
              </EuiFormRow>
            </EuiFlexItem>
          </Fragment>
        )}
      </EuiFlexGroup>
    </EuiPanel>
  );
};
