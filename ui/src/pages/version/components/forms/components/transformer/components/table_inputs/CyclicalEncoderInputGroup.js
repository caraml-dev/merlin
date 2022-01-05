import React from "react";
import { EuiFlexGroup, EuiFlexItem, EuiText, EuiSpacer } from "@elastic/eui";
import { useOnChangeHandler } from "@gojek/mlp-ui";
import { SelectCyclicalEncodeType } from "./SelectCyclicalEncodeType";
import { OrdinalEncoderMapper } from "./OrdinalEncoderMapper";
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

      {cyclicalEncoderConfig.targetValueType && (
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
                    mappings={cyclicalEncoderConfig.mapping}
                    onChangeHandler={onChange("cyclicalEncoderConfig")}
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
