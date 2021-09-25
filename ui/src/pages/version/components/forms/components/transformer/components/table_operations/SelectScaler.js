import React from "react";
import { EuiForm, EuiFormRow, EuiSuperSelect } from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@gojek/mlp-ui";

export const SelectScaler = ({
  column,
  operation, //current selected operation (if any)
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const setValue = value => {
    let newOperation = {
      column: column || "",
      operation: value
    };

    //storage for info about transformation (eg selected col)
    switch (value) {
      case "standardScalerConfig":
        newOperation[value] = {};
        break;
      case "minMaxScalerConfig":
        newOperation[value] = {};
        break;
      default:
        break;
    }
    onChange()(newOperation);
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
    operation !== "" ? option.value === operation : option.value === ""
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
        isInvalid={!!errors.operation}
        error={errors.operation}
        display="columnCompressed">
        <EuiSuperSelect
          fullWidth
          options={options}
          valueOfSelected={selectedOption ? selectedOption.value : ""}
          onChange={value => setValue(value)}
          itemLayoutAlign="top"
          hasDividers
          isInvalid={!!errors.operation}
        />
      </EuiFormRow>
    </EuiForm>
  );
};
