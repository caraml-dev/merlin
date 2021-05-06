import React, { useCallback } from "react";
import {
  EuiButton,
  EuiCard,
  EuiDragDropContext,
  euiDragDropReorder,
  EuiDraggable,
  EuiDroppable,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
  EuiSpacer,
  EuiText
} from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@gojek/mlp-ui";
import { Panel } from "../../Panel";
import { TableTransformationStepCard } from "./TableTransformationStepCard";
import { AddButton } from "./AddButton";

export const TableTransformationStepPanel = ({
  steps,
  onChangeHandler,
  errors = {} // TODO
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const onAddInput = useCallback(() => {
    onChangeHandler([...steps, {}]);
  });

  const onDeleteStep = idx => () => {
    steps.splice(idx, 1);
    onChangeHandler([...steps]);
  };

  const onDragEnd = ({ source, destination }) => {
    if (source && destination) {
      const items = euiDragDropReorder(steps, source.index, destination.index);
      onChangeHandler([...items]);
    }
  };

  return (
    <Panel contentWidth="100%" color="subdued">
      <EuiDragDropContext onDragEnd={onDragEnd}>
        <EuiFlexGroup direction="column" gutterSize="s">
          <EuiDroppable
            droppableId="TABLE_TRANSFORMATION_STEPS_DROPPABLE_AREA"
            spacing="m">
            {steps.map((step, idx) => (
              <EuiDraggable
                key={`${idx}`}
                index={idx}
                draggableId={`${idx}`}
                customDragHandle={true}
                disableInteractiveElementBlocking>
                {provided => (
                  <EuiFlexItem key={`step-${idx}`}>
                    <TableTransformationStepCard
                      index={idx}
                      step={step}
                      onChangeHandler={onChange(`${idx}`)}
                      onDelete={
                        steps.length > 1 ? onDeleteStep(idx) : undefined
                      }
                      dragHandleProps={provided.dragHandleProps}
                    />

                    <EuiSpacer size="s" />
                  </EuiFlexItem>
                )}
              </EuiDraggable>
            ))}
          </EuiDroppable>

          <EuiFlexItem>
            <AddButton
              title="+ Add Step"
              description="Add another table transformation step"
              onClick={() => onAddInput()}
            />
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiDragDropContext>
    </Panel>
  );
};
