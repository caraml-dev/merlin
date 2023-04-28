import React from "react";
import { EuiForm, EuiFormRow, EuiSuperSelect } from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@caraml-dev/ui-lib";

export const SelectTableOperation = ({
  operation, //current selected operation (if any)
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const setValue = value => {
    let newOperation = {
      operation: value
    };

    //storage for info about transformation (eg selected col)
    switch (value) {
      case "dropColumns":
        newOperation[value] = [];
        break;
      case "encodeColumns":
        newOperation[value] = [];
        break;
      case "renameColumns":
        newOperation[value] = {};
        break;
      case "scaleColumns":
        newOperation[value] = [{ column: "" }];
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
      value: "encodeColumns",
      inputDisplay: "Encode Columns"
    },
    {
      value: "renameColumns",
      inputDisplay: "Rename Columns"
    },
    {
      value: "scaleColumns",
      inputDisplay: "Scale Columns"
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
    },
    {
      value: "filterRow",
      inputDisplay: "Filter Row"
    },
    {
      value: "sliceRow",
      inputDisplay: "Slice Row"
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
