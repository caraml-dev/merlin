import React from "react";
import {
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiPanel,
  EuiText,
  EuiSpacer,
  EuiFormRow
} from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@gojek/mlp-ui";
import { DraggableHeader } from "../../../DraggableHeader";
import { SelectEncoder } from "./SelectEncoder";
import { OrdinalEncoderInputGroup } from "./OrdinalEncoderInputGroup";

export const EncoderInputCard = ({
  index = 0,
  encoder,
  onChangeHandler,
  onDelete,
  errors = {},
  ...props
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  return (
    <EuiPanel>
      <DraggableHeader
        onDelete={onDelete}
        dragHandleProps={props.dragHandleProps}
      />

      <EuiSpacer size="s" />

      <EuiFlexGroup direction="column" gutterSize="s">
        <EuiFlexItem>
          <EuiText size="s">
            <h4>Encoder</h4>
          </EuiText>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFlexGroup direction="row">
            <EuiFlexItem>
              <EuiFormRow
                label={
                  <FormLabelWithToolTip
                    label="Encoder Name *"
                    content="The name of encoder that will be passed in the table transformation step."
                  />
                }
                isInvalid={!!errors.name}
                error={errors.name}
                display="columnCompressed"
                fullWidth>
                <EuiFieldText
                  placeholder="Encoder name"
                  value={encoder.name || ""}
                  onChange={e => onChange("name")(e.target.value)}
                  isInvalid={!!errors.name}
                  name={`name-${index}`}
                  fullWidth
                />
              </EuiFormRow>
            </EuiFlexItem>

            <EuiFlexItem>
              <SelectEncoder
                encoderConfig={encoder}
                onChangeHandler={onChangeHandler}
                errors={errors}
              />
            </EuiFlexItem>
          </EuiFlexGroup>

          <EuiSpacer size="s" />

          {encoder["ordinalEncoderConfig"] !== undefined && (
            <OrdinalEncoderInputGroup
              index={index}
              ordinalEncoderConfig={encoder.ordinalEncoderConfig}
              onChangeHandler={onChangeHandler}
              errors={errors}
            />
          )}
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiPanel>
  );
};
