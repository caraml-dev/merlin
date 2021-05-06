import React from "react";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiPanel,
  EuiSpacer,
  EuiText
} from "@elastic/eui";
import { DraggableHeader } from "../../DraggableHeader";
import { VariablesInput } from "./VariablesInput";

export const TablesInputCard = ({
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
            <h4>#{index + 1} - Generic Table</h4>
          </EuiText>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiText>TODO</EuiText>
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiPanel>
  );
};
