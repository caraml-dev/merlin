import React from "react";
import { EuiForm, EuiFormRow, EuiSuperSelect } from "@elastic/eui";

export const SelectJsonFormat = ({ format, onChange, errors = {} }) => {
  const options = [
    {
      value: "RECORD",
      inputDisplay: "RECORD"
    },
    {
      value: "SPLIT",
      inputDisplay: "SPLIT"
    },
    {
      value: "VALUES",
      inputDisplay: "VALUES"
    }
  ];

  const selectedOption = options.find(option =>
    format !== "" ? option.value === format : option.value === ""
  );

  return (
    <EuiForm>
      <EuiFormRow
        fullWidth
        label="JSON Format *"
        isInvalid={!!errors.format}
        error={errors.format}
        display="columnCompressed">
        <EuiSuperSelect
          fullWidth
          options={options}
          valueOfSelected={selectedOption ? selectedOption.value : ""}
          onChange={onChange}
          itemLayoutAlign="top"
          hasDividers
          isInvalid={!!errors.format}
        />
      </EuiFormRow>
    </EuiForm>
  );
};
