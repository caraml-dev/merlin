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
import { DraggableHeader } from "../../DraggableHeader";

export const JsonOutputFieldCard = ({
  index = 0,
  field,
  onChangeHandler,
  onDelete,
  errors = {},
  ...props
}) => {
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
            <h4>#{index + 1} - JSON Field </h4>
          </EuiText>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFormRow
            label="Field"
            isInvalid={!!errors.name}
            error={errors.name}
            display="columnCompressed"
            fullWidth>
            <EuiFieldText
              placeholder="Field name"
              value={field.fieldName}
              // onChange={e => onChange("fieldName", e.target.value)}
              // isInvalid={!!errors.name}
              name={`table-name-${index}`}
              fullWidth
            />
          </EuiFormRow>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFormRow
            label="Field Value *"
            // isInvalid={!!errors.source}
            error={errors.source}
            display="columnCompressed"
            fullWidth>
            <EuiFlexGroup>
              <EuiFlexItem grow={false}>
                <EuiRadio
                  id={`fromJson-${index}`}
                  label="From JSON"
                  checked={(field.value && !!field.value.fromJson) || false}
                  onChange={() => {
                    // onChange("baseTable", { fromJson: new FromJson() });
                  }}
                />
              </EuiFlexItem>
              <EuiFlexItem>
                <EuiRadio
                  id={`fromTable-${index}`}
                  label="From Table"
                  checked={(field.value && !!field.value.fromTable) || false}
                  onChange={() => {
                    // onChange("baseTable", { fromTable: new FromTable() });
                  }}
                />
              </EuiFlexItem>
              <EuiFlexItem>
                <EuiRadio
                  id={`fromExpression-${index}`}
                  label="From Expression"
                  checked={(field.value && !!field.value.expression) || false}
                  onChange={() => {
                    // onChange("baseTable", { fromTable: new FromTable() });
                  }}
                />
              </EuiFlexItem>
            </EuiFlexGroup>
          </EuiFormRow>
        </EuiFlexItem>

        {field.value && field.value.fromJson && (
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
                  value={field.value.fromJson.jsonPath}
                  // onChange={e =>
                  //   onChange("baseTable", {
                  //     fromJson: {
                  //       ...table.baseTable.fromJson,
                  //       jsonPath: e.target.value
                  //     }
                  //   })
                  // }
                  // isInvalid={!!errors.name}
                  name={`json-path-${index}`}
                  fullWidth
                />

                <EuiSpacer size="m" />
              </Fragment>
            </EuiFormRow>
          </EuiFlexItem>
        )}

        {field.value && field.value.fromTable && (
          <EuiFlexItem>
            <EuiFormRow
              label="Source Table Name *"
              // isInvalid={!!errors.name}
              // error={errors.name}
              display="columnCompressed"
              fullWidth>
              <Fragment>
                <EuiFieldText
                  placeholder="Table Name"
                  value={field.value.fromTable.tableName}
                  // onChange={e =>
                  //   onChange("baseTable", {
                  //     fromTable: {
                  //       ...table.baseTable.fromTable,
                  //       tableName: e.target.value
                  //     }
                  //   })
                  // }
                  // isInvalid={!!errors.name}
                  name={`field-from-table-${index}`}
                  fullWidth
                />
              </Fragment>
            </EuiFormRow>

            {/* <EuiFormRow
              label="Columns *"
              // isInvalid={!!errors.name}
              // error={errors.name}
              fullWidth>
              <VariablesInput
                variables={table.columns || []}
                onChangeHandler={onColumnChangeHandler}
                errors={errors}
              />
            </EuiFormRow> */}
          </EuiFlexItem>
        )}
      </EuiFlexGroup>
    </EuiPanel>
  );
};
