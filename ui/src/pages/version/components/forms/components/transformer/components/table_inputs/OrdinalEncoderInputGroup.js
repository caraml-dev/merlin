import React from "react";
import {
  EuiFormRow,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiText,
  EuiSpacer
} from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@caraml-dev/ui-lib";
import { SelectValueType } from "./SelectValueType";
import { OrdinalEncoderMapper } from "./OrdinalEncoderMapper";

export const OrdinalEncoderInputGroup = ({
  index,
  ordinalEncoderConfig,
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  return (
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
              <EuiFieldText
                placeholder="Default value"
                value={ordinalEncoderConfig.defaultValue || ""}
                onChange={e =>
                  onChange("ordinalEncoderConfig.defaultValue")(e.target.value)
                }
                isInvalid={!!errors.defaultValue}
                name={`defaultValue-${index}`}
                fullWidth
              />
            </EuiFormRow>
          </EuiFlexItem>
          <EuiFlexItem>
            <SelectValueType
              valueType={ordinalEncoderConfig.targetValueType}
              onChangeHandler={value =>
                onChange("ordinalEncoderConfig.targetValueType")(value)
              }
              errors={errors}
            />
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiFlexItem>

      <EuiSpacer size="s" />

      {ordinalEncoderConfig.targetValueType && (
        <EuiFlexItem>
          <EuiFlexGroup direction="column" gutterSize="s">
            <EuiFlexItem>
              <EuiText size="xs">
                <h4>Mapping *</h4>
              </EuiText>
            </EuiFlexItem>
            <EuiFlexItem>
              <EuiFlexGroup direction="row">
                <EuiFlexItem>
                  <OrdinalEncoderMapper
                    mappings={ordinalEncoderConfig.mapping}
                    onChangeHandler={onChange("ordinalEncoderConfig")}
                  />
                </EuiFlexItem>
              </EuiFlexGroup>
            </EuiFlexItem>
          </EuiFlexGroup>
        </EuiFlexItem>
      )}
    </EuiFlexGroup>
  );
};
