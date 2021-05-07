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
import {
  FromJson,
  FromTable
} from "../../../../../../../services/transformer/TransformerConfig";
import { useOnChangeHandler } from "@gojek/mlp-ui";

export const BaseJsonOutputCard = ({
  index = 0,
  baseJson,
  onChangeHandler,
  onColumnChangeHandler,
  onDelete,
  ...props
}) => {
  const onChange = (field, value) => {
    console.log("onChange to field " + field + " with value: " + value);
    const xx = { ...baseJson, [field]: value };
    console.log("xx ----- " + JSON.stringify(xx));
    console.log("current base Json " + JSON.stringify(baseJson));
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
            <h4>#{index + 1} - Base Json</h4>
          </EuiText>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFormRow
            label="Json Path"
            // isInvalid={!!errors.name}
            // error={errors.name}
            display="columnCompressed"
            fullWidth>
            <EuiFieldText
              placeholder="$.path"
              value={baseJson.jsonPath}
              onChange={e => onChange("jsonPath", e.target.value)}
              name={`jsonPath`}
              fullWidth
            />
          </EuiFormRow>
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiPanel>
  );
};
