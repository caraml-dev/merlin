import React from "react";
import {
  EuiButton,
  EuiDragDropContext,
  euiDragDropReorder,
  EuiDroppable,
  EuiFlexGroup,
  EuiFlexItem,
  EuiSpacer,
  EuiText
} from "@elastic/eui";
import { useOnChangeHandler } from "@gojek/mlp-ui";
import { Panel } from "../Panel";
import { VariableInputCard } from "./VariableInputCard";
import { TableCreationCard } from "./TableCreationCard";

export const InputPanel = ({
  inputs,
  onChangeHandler,
  errors = {} // TODO
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const onAddInput = () => {
    onChangeHandler([...inputs, {}]);
  };

  const onDragEnd = ({ source, destination }) => {
    if (source && destination) {
      const items = euiDragDropReorder(
        // feastConfig,
        source.index,
        destination.index
      );
      onChangeHandler([...items]);
    }
  };

  return (
    <Panel title="Input" contentWidth="75%">
      <EuiDragDropContext onDragEnd={onDragEnd}>
        <EuiFlexGroup direction="column" gutterSize="s">
          <EuiDroppable droppableId="CUSTOM_HANDLE_DROPPABLE_AREA" spacing="m">
            <EuiFlexItem>
              <TableCreationCard />
              <EuiSpacer size="s" />
            </EuiFlexItem>
            <EuiFlexItem>
              <VariableInputCard />
              <EuiSpacer size="s" />
            </EuiFlexItem>
          </EuiDroppable>

          <EuiFlexItem>
            <EuiButton onClick={onAddInput} size="s">
              <EuiText size="s">+ Add Input</EuiText>
            </EuiButton>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiDragDropContext>
    </Panel>
  );
};
