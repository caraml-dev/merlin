import React, { useState } from "react";
import {
  EuiSpacer,
  EuiDragDropContext,
  euiDragDropReorder,
  EuiDraggable,
  EuiDroppable,
  EuiFlexGroup,
  EuiFlexItem
} from "@elastic/eui";
import { get, useOnChangeHandler } from "@gojek/mlp-ui";
import { AddButton } from "../AddButton";
import { ConditionalUpdateCard } from "./ConditionalUpdateCard";

export const ConditionalUpdatePanel = ({
  conditions = [],
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const items = conditions.length === 0 ? [...conditions, {}] : conditions;

  const onAddIfCondition = () => {
    onChangeHandler([...conditions, { rowSelector: "", expression: "" }]);
  };

  const onAddDefault = () => {
    setDefaultExist(true);
    onChangeHandler([...conditions, { default: { expression: "" } }]);
  };

  const allConditions = conditions.length === 0 ? [] : conditions;

  const [defaultExist, setDefaultExist] = useState(
    allConditions.filter(cond => cond.default !== undefined).length > 0
  );
  const onDeleteCondition = idx => () => {
    var deletedCond = conditions[idx];
    if (deletedCond.default !== undefined) {
      setDefaultExist(false);
    }
    conditions.splice(idx, 1);
    onChange("update")(conditions);
  };

  const onDragEnd = ({ source, destination }) => {
    if (source && destination) {
      const items = euiDragDropReorder(
        conditions,
        source.index,
        destination.index
      );
      onChangeHandler([...items]);
    }
  };

  return (
    <>
      <EuiSpacer size="s" />
      <EuiDragDropContext onDragEnd={onDragEnd}>
        <EuiFlexGroup direction="column" gutterSize="s">
          <EuiDroppable
            droppableId="TABLE_TRANSFORMATION_CONDITIONAL_UPDATE_DROPPABLE_AREA"
            spacing="m">
            {allConditions.map((condition, idx) => (
              <EuiDraggable
                key={`column-${idx}`}
                index={idx}
                draggableId={`${idx}`}
                customDragHandle={true}
                disableInteractiveElementBlocking>
                {provided => (
                  <EuiFlexItem key={`condition-${idx}`}>
                    <ConditionalUpdateCard
                      index={idx}
                      condition={condition}
                      onChangeHandler={onChange(`${idx}`)}
                      onDelete={
                        items.length > 0 ? onDeleteCondition(idx) : undefined
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
            <EuiFlexGroup direction="row" gutterSize="s">
              <EuiFlexItem>
                <AddButton
                  title="+ Add Row Selector"
                  description="Add condition and the execution"
                  onClick={() => onAddIfCondition()}
                />
              </EuiFlexItem>
              <EuiFlexItem>
                <AddButton
                  title="+ Add Default Statement"
                  description="Add default execution if no such condition is satisfied"
                  onClick={() => onAddDefault()}
                  disabled={defaultExist}
                />
              </EuiFlexItem>
            </EuiFlexGroup>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiDragDropContext>
    </>
  );
};
