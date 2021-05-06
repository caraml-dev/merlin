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
import { JsonOutputCard } from "./components/JsonOutputCard";
import { JsonOutput } from "../../../../../../services/transformer/TransformerConfig";

export const OutputPanel = ({
  outputs,
  onChangeHandler,
  errors = {} // TODO
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const onAddOutput = useCallback((field, input) => {
    onChangeHandler([...outputs, { [field]: input }]);
  });

  const onDeleteOutput = idx => () => {
    outputs.splice(idx, 1);
    onChangeHandler([...outputs]);
  };

  const onDragEnd = ({ source, destination }) => {
    if (source && destination) {
      const items = euiDragDropReorder(
        outputs,
        source.index,
        destination.index
      );
      onChangeHandler([...items]);
    }
  };

  return (
    <Panel title="Output" contentWidth="75%">
      <EuiDragDropContext onDragEnd={onDragEnd}>
        <EuiFlexGroup direction="column" gutterSize="s">
          <EuiDroppable droppableId="OUTPUTS_DROPPABLE_AREA" spacing="m">
            {outputs.map((output, idx) => (
              <EuiDraggable
                key={`${idx}`}
                index={idx}
                draggableId={`${idx}`}
                customDragHandle={true}
                disableInteractiveElementBlocking>
                {provided => (
                  <EuiFlexItem key={`output-${idx}`}>
                    {output.jsonOutput && (
                      <JsonOutputCard
                        index={idx}
                        data={output}
                        onChangeHandler={onChange(`${idx}.jsonOutput`)}
                        onDelete={
                          outputs.length > 1 ? onDeleteOutput(idx) : undefined
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
            <AddButton
              title="+ Add JSON Output"
              // TODO:
              // description="Use Feast features as input"
              onClick={() => onAddOutput("jsonOutput", new JsonOutput())}
            />
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiDragDropContext>
    </Panel>
  );
};
