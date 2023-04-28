import React from "react";
import { EuiForm, EuiFormRow, EuiSuperSelect } from "@elastic/eui";
import { FormLabelWithToolTip } from "@caraml-dev/ui-lib";

export const SelectValueType = ({
  valueType, //current selected value type (if any)
  onChangeHandler,
  displayInColumn = true,
  errors = {}
}) => {
  const options = [
    {
      value: "INT",
      inputDisplay: "Integer"
    },
    {
      value: "FLOAT",
      inputDisplay: "Float"
    },
    {
      value: "BOOL",
      inputDisplay: "Boolean"
    },
    {
      value: "STRING",
      inputDisplay: "String"
    }
  ];

  const selectedOption = options.find(option =>
    valueType !== "" ? option.value === valueType : option.value === ""
  );

  return (
    <EuiForm>
      <EuiFormRow
        fullWidth
        label={
          <FormLabelWithToolTip
            label="Value Type *"
            content="Choose a value type"
          />
        }
        isInvalid={!!errors.valueType}
        error={errors.valueType}
        display={displayInColumn ? "columnCompressed" : "rowCompressed"}>
        <EuiSuperSelect
          fullWidth
          options={options}
          valueOfSelected={selectedOption ? selectedOption.value : ""}
          onChange={onChangeHandler}
          itemLayoutAlign="top"
          hasDividers
          isInvalid={!!errors.valueType}
        />
      </EuiFormRow>
    </EuiForm>
  );
};
