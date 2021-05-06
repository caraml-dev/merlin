import React, { useCallback } from "react";
import {
  EuiDragDropContext,
  euiDragDropReorder,
  EuiDraggable,
  EuiDroppable,
  EuiFlexGroup,
  EuiFlexItem,
  EuiSpacer,
  EuiText
} from "@elastic/eui";
import { useOnChangeHandler } from "@gojek/mlp-ui";
import { Panel } from "../Panel";
import { AddButton } from "./components/AddButton";
import { VariablesInputCard } from "./components/VariablesInputCard";
import {
  FeastInput,
  TablesInput,
  VariablesInput
} from "../../../../../../services/transformer/TransformerConfig";
import { FeastResourcesContextProvider } from "../../../../../../providers/feast/FeastResourcesContext";
import { FeastInputCard } from "../feast_config/components/FeastInputCard";
import { TablesInputCard } from "./components/TablesInputCard";

export const InputPanel = ({
  inputs,
  onChangeHandler,
  errors = {} // TODO
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const onAddInput = useCallback((field, input) => {
    onChangeHandler([...inputs, { [field]: input }]);
  });

  const onDeleteInput = idx => () => {
    inputs.splice(idx, 1);
    onChangeHandler([...inputs]);
  };

  const onDragEnd = ({ source, destination }) => {
    if (source && destination) {
      const items = euiDragDropReorder(inputs, source.index, destination.index);
      onChangeHandler([...items]);
    }
  };

  return (
    <Panel title="Input" contentWidth="75%">
      <EuiDragDropContext onDragEnd={onDragEnd}>
        <EuiFlexGroup direction="column" gutterSize="s">
          <EuiDroppable droppableId="INPUTS_DROPPABLE_AREA" spacing="m">
            {inputs.map((input, idx) => (
              <EuiDraggable
                key={`${idx}`}
                index={idx}
                draggableId={`${idx}`}
                customDragHandle={true}
                disableInteractiveElementBlocking>
                {provided => (
                  <EuiFlexItem key={`input-${idx}`}>
                    {input.feast && (
                      <FeastResourcesContextProvider
                        project={input.feast.project}>
                        <FeastInputCard
                          index={idx}
                          table={input.feast}
                          tableNameEditable={true}
                          onChangeHandler={onChange(`${idx}.feast`)}
                          onDelete={
                            inputs.length > 1 ? onDeleteInput(idx) : undefined
                          }
                          dragHandleProps={provided.dragHandleProps}
                        />
                      </FeastResourcesContextProvider>
                    )}

                    {input.tables && (
                      <TablesInputCard
                        index={idx}
                        variables={input.tables}
                        onChangeHandler={onChange(`${idx}.tables`)}
                        onDelete={
                          inputs.length > 1 ? onDeleteInput(idx) : undefined
                        }
                        dragHandleProps={provided.dragHandleProps}
                      />
                    )}

                    {input.variables && (
                      <VariablesInputCard
                        index={idx}
                        variables={input.variables}
                        onChangeHandler={onChange(`${idx}.variables`)}
                        onDelete={
                          inputs.length > 1 ? onDeleteInput(idx) : undefined
                        }
                        dragHandleProps={provided.dragHandleProps}
                      />
                    )}

                    <EuiSpacer size="s" />
                  </EuiFlexItem>
                )}
              </EuiDraggable>
            ))}
          </EuiDroppable>

          <EuiFlexItem>
            <EuiFlexGroup>
              <EuiFlexItem>
                <AddButton
                  title="+ Add Feast Features"
                  description="Use Feast features as input"
                  onClick={() => onAddInput("feast", new FeastInput())}
                />
              </EuiFlexItem>

              <EuiFlexItem>
                <AddButton
                  title="+ Add Generic Table"
                  description="Create generic table from request body or other inputs (Feast features or variables)"
                  onClick={() => onAddInput("tables", new TablesInput())}
                />
              </EuiFlexItem>

              <EuiFlexItem>
                <AddButton
                  title="+ Add Variables"
                  description="Declare literal variable"
                  onClick={() => onAddInput("variables", [])}
                />
              </EuiFlexItem>
            </EuiFlexGroup>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiDragDropContext>
    </Panel>
  );
};
