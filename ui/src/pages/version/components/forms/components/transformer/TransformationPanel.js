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
import { Panel } from "../Panel";
import { AddButton } from "./components/AddButton";
import { TableJoinCard } from "./components/table_operations/TableJoinCard";
import { VariablesInputCard } from "./components/table_inputs/VariablesInputCard";
import { TableTransformationCard } from "./components/table_operations/TableTransformationCard";
import {
  TableJoin,
  TableTransformation
} from "../../../../../../services/transformer/TransformerConfig";

export const TransformationPanel = ({
  transformations = [],
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
    <Panel title="Transformation" contentWidth="92%">
      <EuiSpacer size="s" />

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
                        onDelete={onDeleteTransformation(idx)}
                        dragHandleProps={provided.dragHandleProps}
                        errors={get(errors, `${idx}.tableTransformation`)}
                      />
                    )}

                    {transformation.tableJoin && (
                      <TableJoinCard
                        index={idx}
                        data={transformation.tableJoin}
                        onChangeHandler={onChange(`${idx}.tableJoin`)}
                        onDelete={onDeleteTransformation(idx)}
                        dragHandleProps={provided.dragHandleProps}
                        errors={get(errors, `${idx}.tableJoin`)}
                      />
                    )}

                    {transformation.variables && (
                      <VariablesInputCard
                        index={idx}
                        variables={transformation.variables}
                        onChangeHandler={onChange(`${idx}.variables`)}
                        onDelete={onDeleteTransformation(idx)}
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
                  description="Perform join operation on two tables"
                  onClick={() => onAddInput("tableJoin", new TableJoin())}
                />
              </EuiFlexItem>

              <EuiFlexItem>
                <AddButton
                  title="+ Add Variables"
                  description="Perform storing variable in transformation stage"
                  onClick={() => onAddInput("variables", [{}])}
                />
              </EuiFlexItem>
            </EuiFlexGroup>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiDragDropContext>
    </Panel>
  );
};
