import React from "react";
import {
  EuiDragDropContext,
  euiDragDropReorder,
  EuiDraggable,
  EuiDroppable,
  EuiFlexGroup,
  EuiFlexItem,
  EuiSpacer
} from "@elastic/eui";
import { get, useOnChangeHandler } from "@caraml-dev/ui-lib";
import { AddButton } from "../AddButton";
import { ScaleColumnCard } from "./ScaleColumnCard";
import { Panel } from "../../../Panel";

export const ScaleColumnsGroup = ({
  columns,
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const onAddCol = () => {
    onChangeHandler([...columns, {}]);
  };

  const onDeleteCol = idx => () => {
    columns.splice(idx, 1);
    onChangeHandler([...columns]);
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
    <Panel contentWidth="100%" color="subdued">
      <EuiDragDropContext onDragEnd={onDragEnd}>
        <EuiFlexGroup direction="column" gutterSize="s">
          <EuiDroppable droppableId="SCALE_COLUMNS_DROPPABLE_AREA" spacing="m">
            {columns.map((col, idx) => (
              <EuiDraggable
                key={`${idx}`}
                index={idx}
                draggableId={`${idx}`}
                customDragHandle={true}
                disableInteractiveElementBlocking>
                {provided => (
                  <EuiFlexItem key={`${idx}`}>
                    <ScaleColumnCard
                      index={idx}
                      col={col}
                      onChangeHandler={onChange(`${idx}`)}
                      onDelete={
                        columns.length > 1 ? onDeleteCol(idx) : undefined
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
              title="+ Add Column to Scale"
              description="Add another column to apply scaling"
              onClick={() => onAddCol()}
            />
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiDragDropContext>
    </Panel>
  );
};
