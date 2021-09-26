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
import { SelectOrdinalValueType } from "./SelectOrdinalValueType";
import { OrdinalEncoderMapper } from "./OrdinalEncoderMapper";

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
            <EuiFlexGroup direction="column" gutterSize="s">
              <EuiFlexItem>
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
                    <SelectOrdinalValueType
                      defaultValue={encoder.ordinalEncoderConfig.defaultValue}
                      valueType={encoder.ordinalEncoderConfig.targetValueType}
                      mapping={encoder.ordinalEncoderConfig.mapping}
                      onChangeHandler={onChange("ordinalEncoderConfig")}
                      errors={errors}
                    />
                  </EuiFlexItem>
                </EuiFlexGroup>
              </EuiFlexItem>

              <EuiSpacer size="s" />

              {encoder.ordinalEncoderConfig.targetValueType && (
                <EuiFlexItem>
                  <EuiFlexGroup direction="column" gutterSize="s">
                    <EuiFlexItem>
                      <EuiText size="s">
                        <h4>Mapping:</h4>
                      </EuiText>
                    </EuiFlexItem>
                    <EuiFlexItem>
                      <EuiFlexGroup direction="row">
                        <EuiFlexItem>
                          <OrdinalEncoderMapper
                            mappings={encoder.ordinalEncoderConfig.mapping}
                            onChangeHandler={onChange("ordinalEncoderConfig")}
                          />
                        </EuiFlexItem>
                      </EuiFlexGroup>
                    </EuiFlexItem>
                  </EuiFlexGroup>
                </EuiFlexItem>
              )}
            </EuiFlexGroup>
          )}
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiPanel>
  );
};
