import React from "react";
import { EuiForm, EuiFormRow, EuiSuperSelect } from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@gojek/mlp-ui";

export const SelectTableOperation = ({
  operation,
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const setValue = value => {
    onChange("operation")(value);
  };

  const options = [
    {
      value: "drop",
      inputDisplay: "Drop Columns"
    },
    {
      value: "rename",
      inputDisplay: "Rename Columns"
    },
    {
      value: "select",
      inputDisplay: "Select Columns"
    },
    {
      value: "sort",
      inputDisplay: "Sort Columns"
    },
    {
      value: "update",
      inputDisplay: "Update Columns"
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
            label="Choose Operation *"
            content="Choose the operation to be exectued on the table"
          />
        }
        isInvalid={!!errors.operation}
        error={errors.operation}
        display="columnCompressed">
        <EuiSuperSelect
          fullWidth
          options={options}
          valueOfSelected={
            selectedOption ? selectedOption.value : "dropColumns"
          }
          onChange={value => setValue(value)}
          itemLayoutAlign="top"
          hasDividers
          isInvalid={!!errors.operation}
        />
      </EuiFormRow>
    </EuiForm>
  );
};
