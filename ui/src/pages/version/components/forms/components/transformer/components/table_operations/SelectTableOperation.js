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
    let newOperation = {
      operation: value
    };

    switch (value) {
      case "dropColumns":
        newOperation[value] = [];
        break;
      case "renameColumns":
        newOperation[value] = {};
        break;
      case "selectColumns":
        newOperation[value] = [];
        break;
      case "sort":
        newOperation[value] = [];
        break;
      case "updateColumns":
        newOperation[value] = [];
        break;
      default:
        break;
    }

    onChange()(newOperation);
  };

  const options = [
    {
      value: "dropColumns",
      inputDisplay: "Drop Columns"
    },
    {
      value: "renameColumns",
      inputDisplay: "Rename Columns"
    },
    {
      value: "selectColumns",
      inputDisplay: "Select Columns"
    },
    {
      value: "sort",
      inputDisplay: "Sort Columns"
    },
    {
      value: "updateColumns",
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
