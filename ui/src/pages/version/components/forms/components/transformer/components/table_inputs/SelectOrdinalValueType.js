import React from "react";
import { EuiForm, EuiFormRow, EuiSuperSelect } from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@gojek/mlp-ui";

export const SelectOrdinalValueType = ({
  defaultValue,
  mapping,
  valueType, //current selected value type (if any)
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const setValue = value => {
    //set new value only if there is changes
    if (value !== valueType) {
      let newValueType = {
        defaultValue: defaultValue || "",
        targetValueType: value,
        mapping: mapping || {}
      };

      onChange()(newValueType);
    }
  };

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
        display="columnCompressed">
        <EuiSuperSelect
          fullWidth
          options={options}
          valueOfSelected={selectedOption ? selectedOption.value : ""}
          onChange={value => setValue(value)}
          itemLayoutAlign="top"
          hasDividers
          isInvalid={!!errors.valueType}
        />
      </EuiFormRow>
    </EuiForm>
  );
};
