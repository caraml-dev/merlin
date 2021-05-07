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
import { Panel } from "../Panel";
import { AddButton } from "./components/AddButton";
import {
  TableJoin,
  TableTransformation
} from "../../../../../../services/transformer/TransformerConfig";
import { TableJoinCard } from "./components/TableJoinCard";
import { TableTransformationCard } from "./components/TableTransformationCard";
import { useOnChangeHandler } from "@gojek/mlp-ui";

export const TransformationPanel = ({
  transformations,
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const onAddInput = (field, transformation) => {
    onChangeHandler([...transformations, { [field]: transformation }]);
  };

  const onDeleteTransformation = idx => () => {
    transformations.splice(idx, 1);
    onChangeHandler([...transformations]);
  };

  const onDragEnd = ({ source, destination }) => {
    if (source && destination) {
      const items = euiDragDropReorder(
        transformations,
        source.index,
        destination.index
      );
      onChangeHandler([...items]);
    }
  };

  return (
    <Panel title="Transformation" contentWidth="75%">
      <EuiDragDropContext onDragEnd={onDragEnd}>
        <EuiFlexGroup direction="column" gutterSize="s">
          <EuiDroppable
            droppableId="TRANSFORMATIONS_DROPPABLE_AREA"
            spacing="m">
            {transformations.map((transformation, idx) => (
              <EuiDraggable
                key={`${idx}`}
                index={idx}
                draggableId={`${idx}`}
                customDragHandle={true}
                disableInteractiveElementBlocking>
                {provided => (
                  <EuiFlexItem key={`transformation-${idx}`}>
                    {transformation.tableTransformation && (
                      <TableTransformationCard
                        index={idx}
                        data={transformation.tableTransformation}
                        onChangeHandler={onChange(`${idx}.tableTransformation`)}
                        onDelete={
                          transformations.length > 1
                            ? onDeleteTransformation(idx)
                            : undefined
                        }
                        dragHandleProps={provided.dragHandleProps}
                      />
                    )}

                    {transformation.tableJoin && (
                      <TableJoinCard
                        index={idx}
                        data={transformation.tableJoin}
                        onChangeHandler={onChange(`${idx}.tableJoin`)}
                        onDelete={
                          transformations.length > 1
                            ? onDeleteTransformation(idx)
                            : undefined
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
                  title="+ Add Table Transformation"
                  description="Perform out-of-place transformation on a table"
                  onClick={() =>
                    onAddInput("tableTransformation", new TableTransformation())
                  }
                />
              </EuiFlexItem>

              <EuiFlexItem>
                <AddButton
                  title="+ Add Table Join"
                  description="Perform join operations on two tables"
                  onClick={() => onAddInput("tableJoin", new TableJoin())}
                />
              </EuiFlexItem>
            </EuiFlexGroup>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiDragDropContext>
    </Panel>
  );
};
