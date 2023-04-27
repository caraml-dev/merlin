import React from "react";
import {
  EuiSpacer,
  EuiDragDropContext,
  euiDragDropReorder,
  EuiDraggable,
  EuiDroppable,
  EuiFlexGroup,
  EuiFlexItem,
  EuiPanel
} from "@elastic/eui";
import { get, useOnChangeHandler } from "@caraml-dev/ui-lib";
import { AddButton } from "../AddButton";
import { UpdateColumnCard } from "./UpdateColumnCard";

export const UpdateColumnPanel = ({
  columns = [],
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const items = columns.length === 0 ? [...columns, {}] : columns;

  const onAddColumn = () => {
    onChangeHandler([...columns, {}]);
  };

  const onDeleteColumn = idx => () => {
    columns.splice(idx, 1);
    onChange("update")(columns);
  };

  const onDragEnd = ({ source, destination }) => {
    if (source && destination) {
      const items = euiDragDropReorder(
        columns,
        source.index,
        destination.index
      );
      onChangeHandler([...items]);
    }
  };

  return (
    <EuiPanel title="Columns">
      <EuiSpacer size="s" />
      <EuiDragDropContext onDragEnd={onDragEnd}>
        <EuiFlexGroup direction="column" gutterSize="s">
          <EuiDroppable
            droppableId="TABLE_TRANSFORMATION_STEPS_UPDATE_COLUMN_DROPPABLE_AREA"
            spacing="m">
            {items.map((column, idx) => (
              <EuiDraggable
                key={`column-${idx}`}
                index={idx}
                draggableId={`${idx}`}
                customDragHandle={true}
                disableInteractiveElementBlocking>
                {provided => (
                  <EuiFlexItem key={`column-${idx}`}>
                    <UpdateColumnCard
                      index={idx}
                      column={column}
                      onChangeHandler={onChange(`${idx}`)}
                      onDelete={
                        items.length > 1 ? onDeleteColumn(idx) : undefined
                      }
                      dragHandleProps={provided.dragHandleProps}
                      errors={get(errors, `${idx}`)}
                    />
                    <EuiSpacer size="s" />
                  </EuiFlexItem>
                )}
              </EuiDraggable>
            ))}
          </EuiDroppable>

          <EuiFlexItem>
            <AddButton
              title="+ Add Column"
              description="Add column to updated or created"
              onClick={() => onAddColumn()}
            />
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiDragDropContext>
    </EuiPanel>
  );
};
