import React from "react";
import { EuiForm, EuiFormRow, EuiSuperSelect } from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@gojek/mlp-ui";

export const SelectScaler = ({
  scalerConfig, //current scaler config (if any)
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const setValue = value => {
    if (scalerConfig[value] === undefined) {
      let newScalerConfig = {
        column: scalerConfig.column || ""
      };

      //storage for info about transformation (eg selected col)
      switch (value) {
        case "standardScalerConfig":
          newScalerConfig[value] = {};
          break;
        case "minMaxScalerConfig":
          newScalerConfig[value] = {};
          break;
        default:
          break;
      }
      onChange()(newScalerConfig);
    }
  };

  const options = [
    {
      value: "standardScalerConfig",
      inputDisplay: "Standard Scaler"
    },
    {
      value: "minMaxScalerConfig",
      inputDisplay: "Min-Max Scaler"
    }
  ];

  const selectedOption = options.find(option =>
    scalerConfig !== undefined
      ? scalerConfig[option.value] !== undefined
      : option.value === ""
  );

  return (
    <EuiForm>
      <EuiFormRow
        fullWidth
        label={
          <FormLabelWithToolTip
            label="Scaler *"
            content="Choose a scaler type to apply to a column"
          />
        }
        isInvalid={!!errors.scalerConfig}
        error={errors.scalerConfig}
        display="columnCompressed">
        <EuiSuperSelect
          fullWidth
          options={options}
          valueOfSelected={selectedOption ? selectedOption.value : ""}
          onChange={value => setValue(value)}
          itemLayoutAlign="top"
          hasDividers
          isInvalid={!!errors.scalerConfig}
        />
      </EuiFormRow>
    </EuiForm>
  );
};
