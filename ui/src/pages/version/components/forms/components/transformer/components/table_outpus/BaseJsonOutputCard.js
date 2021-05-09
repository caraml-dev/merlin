import React from "react";
import {
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiPanel,
  EuiSpacer,
  EuiText
} from "@elastic/eui";
import { DraggableHeader } from "../../../DraggableHeader";

export const BaseJsonOutputCard = ({
  baseJson,
  onChangeHandler,
  onColumnChangeHandler,
  onDelete,
  errors = {},
  ...props
}) => {
  const onChange = (field, value) => {
    if (JSON.stringify(value) !== JSON.stringify(baseJson[field])) {
      onChangeHandler({
        ...baseJson,
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
            <h4>Base JSON</h4>
          </EuiText>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFormRow
            label="JSONPath"
            isInvalid={!!errors.jsonPath}
            error={errors.jsonPath}
            display="columnCompressed"
            fullWidth>
            <EuiFieldText
              placeholder="$.path"
              value={baseJson.jsonPath}
              onChange={e => onChange("jsonPath", e.target.value)}
              name={`jsonPath`}
              isInvalid={!!errors.jsonPath}
              fullWidth
            />
          </EuiFormRow>
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiPanel>
  );
};
