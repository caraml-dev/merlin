import React from "react";
import {
  EuiFieldText,
  EuiFieldNumber,
  EuiFlexGroup,
  EuiFlexItem,
  EuiPanel,
  EuiSpacer,
  EuiText,
  EuiFormRow
} from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@gojek/mlp-ui";
import { DraggableHeader } from "../../../DraggableHeader";
import { SelectEncoder } from "./SelectEncoder";

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
        <EuiFlexGroup direction="row">
          <EuiFlexItem>
            <EuiFormRow
              label={
                <FormLabelWithToolTip
                  label="Encoder Name *"
                  content="Name of encoder"
                />
              }
              isInvalid={!!errors.name}
              error={errors.name}
              display="columnCompressed"
              fullWidth>
              <EuiFieldText
                placeholder="encoder name"
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
              name={encoder.name}
              encoder={encoder.encoder}
              onChangeHandler={onChangeHandler}
              errors={errors}
            />
          </EuiFlexItem>
        </EuiFlexGroup>
        <EuiSpacer size="s" />
        {encoder.encoder === "ordinalEncoderConfig" && (
          <EuiFlexGroup direction="row">
            <EuiFlexItem>
              <EuiFormRow
                label={
                  <FormLabelWithToolTip
                    label="Default Value *"
                    content="Default value to map to if mapping undefined"
                  />
                }
                isInvalid={!!errors.defaultValue}
                error={errors.defaultValue}
                display="columnCompressed"
                fullWidth>
                <EuiFieldNumber
                  placeholder="default value"
                  value={encoder.ordinalEncoderConfig.defaultValue || ""}
                  onChange={e =>
                    onChange("ordinalEncoderConfig.defaultValue")(
                      e.target.value
                    )
                  }
                  isInvalid={!!errors.defaultValue}
                  name={`defaultValue-${index}`}
                  fullWidth
                />
              </EuiFormRow>
            </EuiFlexItem>
            <EuiFlexItem>
              <EuiFormRow
                label={
                  <FormLabelWithToolTip
                    label="Value Type *"
                    content="Value type to map to"
                  />
                }
                isInvalid={!!errors.targetValueType}
                error={errors.targetValueType}
                display="columnCompressed"
                fullWidth>
                <EuiFieldNumber
                  placeholder="value type"
                  value={encoder.ordinalEncoderConfig.targetValueType || ""}
                  onChange={e =>
                    onChange("ordinalEncoderConfig.targetValueType")(
                      e.target.value
                    )
                  }
                  isInvalid={!!errors.targetValueType}
                  name={`targetValueType-${index}`}
                  fullWidth
                />
              </EuiFormRow>
            </EuiFlexItem>
          </EuiFlexGroup>
        )}
      </EuiFlexGroup>
    </EuiPanel>
  );
};
