import React from "react";
import {
  EuiForm,
  EuiFormRow,
  EuiIcon,
  EuiSuperSelect,
  EuiText,
  EuiToolTip
} from "@elastic/eui";
import { useOnChangeHandler } from "@gojek/mlp-ui";
import { Panel } from "../Panel";

export const OutputPanel = ({ onChangeHandler, errors = {} }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const jsonTypeOptions = [
    {
      value: "prediction-api-v1",
      inputDisplay: "Prediction API V1"
    }
  ];

  const inputTableOptions = [
    {
      value: "customer_table",
      inputDisplay: "customer_table"
    }
  ];

  return (
    <Panel title="Output">
      <EuiForm>
        <EuiFormRow
          fullWidth
          label="JSON Format Type*"
          isInvalid={!!errors.environment_name}
          error={errors.environment_name}
          display="row">
          <EuiSuperSelect
            fullWidth
            options={jsonTypeOptions}
            valueOfSelected="prediction-api-v1"
            onChange={onChange("environment_name")}
            isInvalid={!!errors.environment_name}
            itemLayoutAlign="top"
            hasDividers
          />
        </EuiFormRow>

        <EuiFormRow
          fullWidth
          label="Input Table*"
          isInvalid={!!errors.environment_name}
          error={errors.environment_name}
          display="row">
          <EuiSuperSelect
            fullWidth
            options={inputTableOptions}
            valueOfSelected="customer_table"
            onChange={onChange("environment_name")}
            isInvalid={!!errors.environment_name}
            itemLayoutAlign="top"
            hasDividers
          />
        </EuiFormRow>
      </EuiForm>
    </Panel>
  );
};
