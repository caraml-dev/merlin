import React from "react";
import {
  EuiButton,
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
import { FeastInputCard } from "./components/FeastInputCard";
import { FeastResourcesContextProvider } from "../../../../../../providers/feast/FeastResourcesContext";
import { FeastInput } from "../../../../../../services/transformer/TransformerConfig";

export const FeastEnricherPanel = ({
  feastConfig,
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const onAddTable = () => {
    onChangeHandler([...feastConfig, new FeastInput()]);
  };

  const onDeleteTable = idx => () => {
    feastConfig.splice(idx, 1);
    onChangeHandler([...feastConfig]);
  };

  const onDragEnd = ({ source, destination }) => {
    if (source && destination) {
      const items = euiDragDropReorder(
        feastConfig,
        source.index,
        destination.index
      );
      onChangeHandler([...items]);
    }
  };

  return (
    <Panel title="Feast Enricher" contentWidth="80%">
      <EuiSpacer size="m" />

      <EuiDragDropContext onDragEnd={onDragEnd}>
        <EuiFlexGroup direction="column" gutterSize="s">
          <EuiDroppable droppableId="FEAST_INPUT_DROPPABLE_AREA" spacing="m">
            {feastConfig.map((table, idx) => (
              <EuiDraggable
                key={`${idx}`}
                index={idx}
                draggableId={`${idx}`}
                customDragHandle={true}
                disableInteractiveElementBlocking>
                {provided => (
                  <EuiFlexItem key={`feast-table-${idx}`}>
                    <FeastResourcesContextProvider project={table.project}>
                      <FeastInputCard
                        index={idx}
                        table={table}
                        onChangeHandler={onChange(`${idx}`)}
                        onDelete={
                          feastConfig.length > 1
                            ? onDeleteTable(idx)
                            : undefined
                        }
                        dragHandleProps={provided.dragHandleProps}
                        errors={errors[idx]}
                      />
                    </FeastResourcesContextProvider>
                    <EuiSpacer size="s" />
                  </EuiFlexItem>
                )}
              </EuiDraggable>
            ))}
          </EuiDroppable>

          <EuiFlexItem>
            <EuiButton onClick={onAddTable} size="s">
              <EuiText size="s">+ Add Feast Input</EuiText>
            </EuiButton>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiDragDropContext>
    </Panel>
  );
};
