import React from "react";
import {
  EuiDragDropContext,
  euiDragDropReorder,
  EuiDraggable,
  EuiDroppable,
  EuiFlexGroup,
  EuiFlexItem,
  EuiPanel,
  EuiSpacer
} from "@elastic/eui";
import { get, useOnChangeHandler } from "@gojek/mlp-ui";
import { DraggableHeader } from "../../DraggableHeader";
import { FeastResourcesContextProvider } from "../../../../../../../providers/feast/FeastResourcesContext";
import { AddButton } from "../../transformer/components/AddButton";
import { FeastInput } from "../../../../../../../services/transformer/TransformerConfig";
import { FeastInputCard } from "./FeastInputCard";

export const FeastInputGroup = ({
  groupIndex,
  tables,
  tableNameEditable = false,
  onChangeHandler,
  onDelete,
  errors = {},
  ...props
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const onAddTable = () => {
    onChange(`${groupIndex}.feast`)([...tables, new FeastInput(true)]);
  };

  const onDeleteTable = idx => () => {
    tables.splice(idx, 1);
    onChange(`${groupIndex}.feast`)([...tables]);
  };

  const onDragEnd = ({ source, destination }) => {
    if (source && destination) {
      const items = euiDragDropReorder(tables, source.index, destination.index);
      onChange(`${groupIndex}.feast`)([...items]);
    }
  };

  return (
    <EuiPanel>
      <DraggableHeader
        onDelete={onDelete}
        dragHandleProps={props.dragHandleProps}
      />

      <EuiSpacer size="m" />

      <EuiFlexGroup direction="column" gutterSize="s">
        <EuiFlexGroup direction="column" gutterSize="s">
          <EuiDragDropContext onDragEnd={onDragEnd}>
            <EuiDroppable droppableId="FEAST_TABLE_DROPPABLE_AREA" spacing="m">
              {tables.map((table, tableIdx) => (
                <EuiDraggable
                  key={`${tableIdx}`}
                  index={tableIdx}
                  draggableId={`${tableIdx}`}
                  customDragHandle={true}
                  disableInteractiveElementBlocking>
                  {provided => (
                    <EuiFlexItem>
                      <FeastResourcesContextProvider project={table.project}>
                        <FeastInputCard
                          index={tableIdx}
                          table={table}
                          tableNameEditable={true}
                          onChangeHandler={onChange(
                            `${groupIndex}.feast.${tableIdx}`
                          )}
                          onDelete={onDeleteTable(tableIdx)}
                          dragHandleProps={provided.dragHandleProps}
                          errors={get(
                            errors,
                            `${groupIndex}.feast.${tableIdx}`
                          )}
                        />

                        <EuiSpacer size="s" />
                      </FeastResourcesContextProvider>
                    </EuiFlexItem>
                  )}
                </EuiDraggable>
              ))}
            </EuiDroppable>
          </EuiDragDropContext>
        </EuiFlexGroup>

        <EuiFlexItem>
          <AddButton
            title="+ Add Feast Table"
            description="Add another Feast table that will be created asynchronously with other tables in this group."
            titleSize="xs"
            onClick={() => onAddTable()}
          />
          {/* <div>
            <EuiToolTip
              position="bottom"
              content="Add another Feast table that will be created asynchronously with other tables in this group.">
              <EuiButtonEmpty size="s">
                <EuiText size="xs">+ Add Feast Table</EuiText>
              </EuiButtonEmpty>
            </EuiToolTip>
          </div> */}
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiPanel>
  );
};
