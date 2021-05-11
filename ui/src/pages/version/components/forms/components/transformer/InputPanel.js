import React, { Fragment } from "react";
import {
  EuiDragDropContext,
  euiDragDropReorder,
  EuiDraggable,
  EuiDroppable,
  EuiFlexGroup,
  EuiFlexItem,
  EuiSpacer
} from "@elastic/eui";
import { get, useOnChangeHandler } from "@gojek/mlp-ui";
import { AddButton } from "./components/AddButton";
import { TableInputCard } from "./components/table_inputs/TableInputCard";
import { VariablesInputCard } from "./components/table_inputs/VariablesInputCard";
import { FeastInputCard } from "../feast_config/components/FeastInputCard";
import { Panel } from "../Panel";
import {
  FeastInput,
  TablesInput
} from "../../../../../../services/transformer/TransformerConfig";
import { FeastResourcesContextProvider } from "../../../../../../providers/feast/FeastResourcesContext";

export const InputPanel = ({ inputs = [], onChangeHandler, errors = {} }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const onAddInput = (field, input) => {
    onChangeHandler([...inputs, { [field]: input }]);
  };

  const onDeleteInput = idx => () => {
    inputs.splice(idx, 1);
    onChangeHandler([...inputs]);
  };

  const onDeleteInputChild = (idx, field, childIdx) => () => {
    inputs[idx][field].splice(childIdx, 1);
    onChangeHandler([...inputs]);
  };

  const onDragEnd = ({ source, destination }) => {
    if (source && destination) {
      const items = euiDragDropReorder(inputs, source.index, destination.index);
      onChangeHandler([...items]);
    }
  };

  return (
    <Panel title="Input" contentWidth="80%">
      <EuiSpacer size="xs" />

      <EuiFlexGroup direction="column" gutterSize="s">
        <EuiDragDropContext onDragEnd={onDragEnd}>
          <EuiDroppable droppableId="INPUTS_DROPPABLE_AREA" spacing="m">
            {inputs.map((input, idx) => (
              <EuiDraggable
                key={`input-${idx}`}
                index={idx}
                draggableId={`${idx}`}
                customDragHandle={true}
                disableInteractiveElementBlocking>
                {provided => (
                  <EuiFlexItem>
                    {input.feast &&
                      input.feast.map((feast, feastIdx) => (
                        <Fragment key={`input-${idx}-feast-${feastIdx}`}>
                          <FeastResourcesContextProvider
                            project={feast.project}>
                            <FeastInputCard
                              index={idx}
                              table={feast}
                              tableNameEditable={true}
                              onChangeHandler={onChange(
                                `${idx}.feast.${feastIdx}`
                              )}
                              onDelete={
                                input.feast.length > 1
                                  ? onDeleteInputChild(idx, "feast", feastIdx)
                                  : onDeleteInput(idx)
                              }
                              dragHandleProps={provided.dragHandleProps}
                              errors={get(errors, `${idx}.feast.${feastIdx}`)}
                            />
                          </FeastResourcesContextProvider>
                          {feastIdx < input.feast.length - 1 && (
                            <EuiSpacer size="s" />
                          )}
                        </Fragment>
                      ))}

                    {input.tables &&
                      input.tables.map((table, tableIdx) => (
                        <Fragment key={`input-${idx}-table-${tableIdx}`}>
                          <TableInputCard
                            index={idx}
                            table={table}
                            onChangeHandler={onChange(
                              `${idx}.tables.${tableIdx}`
                            )}
                            onColumnChangeHandler={onChange(
                              `${idx}.tables.${tableIdx}.columns`
                            )}
                            onDelete={
                              input.tables.length > 1
                                ? onDeleteInputChild(idx, "tables", tableIdx)
                                : onDeleteInput(idx)
                            }
                            dragHandleProps={provided.dragHandleProps}
                            errors={get(errors, `${idx}.tables.${tableIdx}`)}
                          />
                          {tableIdx < input.tables.length - 1 && (
                            <EuiSpacer size="s" />
                          )}
                        </Fragment>
                      ))}

                    {input.variables && (
                      <VariablesInputCard
                        index={idx}
                        variables={input.variables}
                        onChangeHandler={onChange(`${idx}.variables`)}
                        onDelete={onDeleteInput(idx)}
                        dragHandleProps={provided.dragHandleProps}
                        errors={get(errors, `${idx}.variables`)}
                      />
                    )}

                    <EuiSpacer size="s" />
                  </EuiFlexItem>
                )}
              </EuiDraggable>
            ))}
          </EuiDroppable>
        </EuiDragDropContext>

        <EuiFlexItem>
          <EuiFlexGroup>
            <EuiFlexItem>
              <AddButton
                title="+ Add Feast Features"
                description="Create a table by using features retrieved from Feast"
                onClick={() => onAddInput("feast", [new FeastInput(true)])}
              />
            </EuiFlexItem>

            <EuiFlexItem>
              <AddButton
                title="+ Add Generic Table"
                description="Create a table from request/response json payload or other inputs"
                onClick={() => onAddInput("tables", [new TablesInput()])}
              />
            </EuiFlexItem>

            <EuiFlexItem>
              <AddButton
                title="+ Add Variables"
                description="Declare a variable that can be used as input to expression"
                onClick={() => onAddInput("variables", [{}])}
              />
            </EuiFlexItem>
          </EuiFlexGroup>
        </EuiFlexItem>
      </EuiFlexGroup>
    </Panel>
  );
};
