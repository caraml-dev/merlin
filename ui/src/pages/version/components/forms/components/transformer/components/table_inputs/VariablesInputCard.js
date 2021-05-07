import React from "react";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiPanel,
  EuiSpacer,
  EuiText
} from "@elastic/eui";
import { DraggableHeader } from "../../../DraggableHeader";
import { VariablesInput } from "./VariablesInput";

export const VariablesInputCard = ({
  index = 0,
  variables,
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

      <EuiFlexGroup direction="column" gutterSize="s">
        <EuiFlexItem>
          <EuiText size="s">
            <h4>#{index + 1} - Variables</h4>
          </EuiText>
        </EuiFlexItem>

        <EuiFlexItem>
          <VariablesInput
            variables={variables}
            onChangeHandler={onChangeHandler}
            errors={errors}
          />
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiPanel>
  );
};
