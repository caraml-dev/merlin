import React from "react";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiSpacer,
  EuiFieldNumber,
  EuiFormRow
} from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@caraml-dev/ui-lib";
import { SelectCyclicalEncodeType } from "./SelectCyclicalEncodeType";
import { SelectPeriodType } from "./SelectPeriodType";

export const CyclicalEncoderInputGroup = ({
  index,
  cyclicalEncoderConfig,
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  return (
    <EuiFlexGroup direction="column" gutterSize="s">
      <EuiFlexItem>
        <EuiFlexGroup direction="row">
          <EuiFlexItem>
            <SelectCyclicalEncodeType
              cyclicType={cyclicalEncoderConfig}
              onChangeHandler={onChange("cyclicalEncoderConfig")}
              errors={errors}
            />
          </EuiFlexItem>
          <EuiFlexItem>
            {cyclicalEncoderConfig["byEpochTime"] !== undefined && (
              <SelectPeriodType
                periodType={cyclicalEncoderConfig.byEpochTime}
                onChangeHandler={onChange("cyclicalEncoderConfig.byEpochTime")}
                errors={errors}
              />
            )}
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiFlexItem>

      <EuiSpacer size="s" />

      {cyclicalEncoderConfig["byRange"] !== undefined && (
        <EuiFlexItem>
          <EuiFlexGroup direction="row">
            <EuiFlexItem>
              <EuiFormRow
                label={
                  <FormLabelWithToolTip
                    label="Min *"
                    content="Min of range (inclusive)"
                  />
                }
                isInvalid={!!errors.min}
                error={errors.min}
                display="columnCompressed"
                fullWidth>
                <EuiFieldNumber
                  placeholder="min (inclusive)"
                  value={cyclicalEncoderConfig["byRange"].min || ""}
                  onChange={e =>
                    onChange("cyclicalEncoderConfig.byRange.min")(
                      e.target.value
                    )
                  }
                  isInvalid={!!errors.min}
                  name={`min-${index}`}
                  fullWidth
                />
              </EuiFormRow>
            </EuiFlexItem>
            <EuiFlexItem>
              <EuiFormRow
                label={
                  <FormLabelWithToolTip
                    label="Max *"
                    content="Max of range (exclusive)"
                  />
                }
                isInvalid={!!errors.max}
                error={errors.max}
                display="columnCompressed"
                fullWidth>
                <EuiFieldNumber
                  placeholder="max (exclusive)"
                  value={cyclicalEncoderConfig["byRange"].max || ""}
                  onChange={e =>
                    onChange("cyclicalEncoderConfig.byRange.max")(
                      e.target.value
                    )
                  }
                  isInvalid={!!errors.max}
                  name={`max-${index}`}
                  fullWidth
                />
              </EuiFormRow>
            </EuiFlexItem>
          </EuiFlexGroup>
        </EuiFlexItem>
      )}
    </EuiFlexGroup>
  );
};
