import React, { Fragment } from "react";
import {
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
import { SelectJsonFormat } from "./SelectJsonFormat";

export const JsonOutputFieldCard = ({
  index = 0,
  field,
  onChange,
  onDelete,
  errors = {},
  ...props
}) => {
  const setSource = source => {
    let newField = { fieldName: field.fieldName };
    switch (source) {
      case "fromJson":
        newField[source] = {};
        break;
      case "fromTable":
        newField[source] = {};
        break;
      case "expression":
        newField[source] = "";
        break;
      default:
        break;
    }
    onChange(index, newField);
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
            <h4>#{index + 1} Field</h4>
          </EuiText>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFormRow
            label="Field Name"
            isInvalid={!!errors.fieldName}
            error={errors.fieldName}
            display="columnCompressed"
            fullWidth>
            <EuiFieldText
              placeholder="Field name"
              value={field.fieldName || ""}
              onChange={e =>
                onChange(index, { ...field, fieldName: e.target.value })
              }
              name={`field-name-${index}`}
              isInvalid={!!errors.fieldName}
              fullWidth
            />
          </EuiFormRow>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFormRow
            label="Field Value *"
            isInvalid={!!errors.source}
            error={errors.source}
            display="columnCompressed"
            fullWidth>
            <EuiFlexGroup>
              <EuiFlexItem grow={false}>
                <EuiRadio
                  id={`fromJson-${index}`}
                  label="From JSON"
                  checked={!!field.fromJson || false}
                  onChange={() => setSource("fromJson")}
                />
              </EuiFlexItem>
              <EuiFlexItem grow={false}>
                <EuiRadio
                  id={`fromTable-${index}`}
                  label="From Table"
                  checked={!!field.fromTable || false}
                  onChange={() => setSource("fromTable")}
                />
              </EuiFlexItem>
              <EuiFlexItem>
                <EuiRadio
                  id={`expression-${index}`}
                  label="From Expression"
                  checked={field.expression !== undefined || false}
                  onChange={() => setSource("expression")}
                />
              </EuiFlexItem>
            </EuiFlexGroup>
          </EuiFormRow>
        </EuiFlexItem>

        {field.fromJson && (
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
                  value={field.fromJson.jsonPath || ""}
                  onChange={e =>
                    onChange(index, {
                      ...field,
                      fromJson: { jsonPath: e.target.value }
                    })
                  }
                  // isInvalid={!!errors.name}
                  name={`json-path-${index}`}
                  fullWidth
                />

                <EuiSpacer size="m" />
              </Fragment>
            </EuiFormRow>
          </EuiFlexItem>
        )}

        {field.fromTable && (
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
                  value={field.fromTable.tableName || ""}
                  onChange={e =>
                    onChange(index, {
                      ...field,
                      fromTable: {
                        ...field.fromTable,
                        tableName: e.target.value
                      }
                    })
                  }
                  // isInvalid={!!errors.name}
                  name={`field-from-table-${index}`}
                  fullWidth
                />
              </Fragment>
            </EuiFormRow>

            <EuiSpacer size="m" />

            <SelectJsonFormat
              format={field.fromTable.format || ""}
              onChange={value =>
                onChange(index, {
                  ...field,
                  fromTable: {
                    ...field.fromTable,
                    format: value
                  }
                })
              }
              errors={errors}
            />
          </EuiFlexItem>
        )}

        {field.expression !== undefined && (
          <EuiFlexItem>
            <EuiFormRow
              label="Expression *"
              // isInvalid={!!errors.name}
              // error={errors.name}
              display="columnCompressed"
              fullWidth>
              <Fragment>
                <EuiFieldText
                  placeholder="Expression"
                  value={field.expression || ""}
                  onChange={e =>
                    onChange(index, { ...field, expression: e.target.value })
                  }
                  // isInvalid={!!errors.name}
                  name={`field-expression-${index}`}
                  fullWidth
                />
              </Fragment>
            </EuiFormRow>
          </EuiFlexItem>
        )}
      </EuiFlexGroup>
    </EuiPanel>
  );
};
