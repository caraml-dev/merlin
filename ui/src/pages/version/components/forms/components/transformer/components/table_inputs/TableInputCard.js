import React, { Fragment } from "react";
import {
  EuiCheckbox,
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
import {
  FromJson,
  FromTable
} from "../../../../../../../../services/transformer/TransformerConfig";
import { VariablesInput } from "./VariablesInput";

export const TableInputCard = ({
  index = 0,
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
            <h4>#{index + 1} - Generic Table</h4>
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
            label="Table Source *"
            isInvalid={!!errors.source}
            error={errors.source}
            display="columnCompressed"
            fullWidth>
            <EuiFlexGroup>
              <EuiFlexItem grow={false}>
                <EuiRadio
                  id={`fromJson-${index}`}
                  label="From JSON"
                  checked={
                    (table.baseTable && !!table.baseTable.fromJson) || false
                  }
                  onChange={() => {
                    onChange("baseTable", { fromJson: new FromJson() });
                  }}
                />
              </EuiFlexItem>
              <EuiFlexItem>
                <EuiRadio
                  id={`fromTable-${index}`}
                  label="From Table"
                  checked={
                    (table.baseTable && !!table.baseTable.fromTable) || false
                  }
                  onChange={() => {
                    onChange("baseTable", { fromTable: new FromTable() });
                  }}
                />
              </EuiFlexItem>
            </EuiFlexGroup>
          </EuiFormRow>
        </EuiFlexItem>

        {table.baseTable && table.baseTable.fromJson && (
          <EuiFlexItem>
            <EuiFormRow
              label="JSONPath *"
              // isInvalid={!!errors.name}
              // error={errors.name}
              display="columnCompressed"
              fullWidth>
              <Fragment>
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

                <EuiSpacer size="m" />

                <EuiCheckbox
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
              </Fragment>
            </EuiFormRow>
          </EuiFlexItem>
        )}

        {table.baseTable && table.baseTable.fromTable && (
          <EuiFlexItem>
            <EuiFormRow
              label="Source Table Name *"
              // isInvalid={!!errors.name}
              // error={errors.name}
              display="columnCompressed"
              fullWidth>
              <Fragment>
                <EuiFieldText
                  placeholder="Source Table Name"
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
                  name={`source-table-name-${index}`}
                  fullWidth
                />
              </Fragment>
            </EuiFormRow>

            <EuiFormRow
              label="Columns *"
              // isInvalid={!!errors.name}
              // error={errors.name}
              fullWidth>
              <VariablesInput
                variables={table.columns || []}
                onChangeHandler={onColumnChangeHandler}
                errors={errors}
              />
            </EuiFormRow>
          </EuiFlexItem>
        )}
      </EuiFlexGroup>
    </EuiPanel>
  );
};
