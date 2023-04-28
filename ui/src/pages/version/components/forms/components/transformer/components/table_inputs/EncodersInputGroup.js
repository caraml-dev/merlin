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
import { get, useOnChangeHandler } from "@caraml-dev/ui-lib";
import { DraggableHeader } from "../../../DraggableHeader";
import { AddButton } from "../../../transformer/components/AddButton";
import { EncoderInputCard } from "./EncoderInputCard";

export const EncodersInputGroup = ({
  groupIndex,
  encoders,
  onChangeHandler,
  onDelete,
  errors = {},
  ...props
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const onAddEncoder = () => {
    onChange(`${groupIndex}.encoders`)([...encoders, { name: "" }]);
  };

  const onDeleteEncoder = idx => () => {
    encoders.splice(idx, 1);
    onChange(`${groupIndex}.encoders`)([...encoders]);
  };

  const onDragEnd = ({ source, destination }) => {
    if (source && destination) {
      const items = euiDragDropReorder(
        encoders,
        source.index,
        destination.index
      );
      onChange(`${groupIndex}.encoders`)([...items]);
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
            <EuiDroppable
              droppableId="ENCODERS_INPUT_DROPPABLE_AREA"
              spacing="m">
              {encoders.map((encoder, encoderIdx) => (
                <EuiDraggable
                  key={`${encoderIdx}`}
                  index={encoderIdx}
                  draggableId={`${encoderIdx}`}
                  customDragHandle={true}
                  disableInteractiveElementBlocking>
                  {provided => (
                    <EuiFlexItem key={`${encoderIdx}`}>
                      <EncoderInputCard
                        index={encoderIdx}
                        encoder={encoder}
                        onChangeHandler={onChange(
                          `${groupIndex}.encoders.${encoderIdx}`
                        )}
                        onDelete={
                          encoders.length > 1
                            ? onDeleteEncoder(encoderIdx)
                            : undefined
                        }
                        dragHandleProps={provided.dragHandleProps}
                        errors={get(
                          errors,
                          `${groupIndex}.encoders.${encoderIdx}`
                        )}
                      />
                      <EuiSpacer size="s" />
                    </EuiFlexItem>
                  )}
                </EuiDraggable>
              ))}
            </EuiDroppable>
          </EuiDragDropContext>
        </EuiFlexGroup>

        <EuiFlexItem>
          <AddButton
            title="+ Add Another Encoder"
            description="Declare a new encoder function that can be used on table transformation step."
            titleSize="xs"
            onClick={() => onAddEncoder()}
          />
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiPanel>
  );
};
