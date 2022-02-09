import React from "react";
import { EuiForm, EuiFormRow, EuiSuperSelect } from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@gojek/mlp-ui";

export const SelectPeriodType = ({
  periodType, //type of input to be encoded
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const setValue = value => {
    //set new value only if there is changes
    if (periodType[value] === undefined) {
      let newPeriodType = {};

      newPeriodType["periodType"] = value;
      onChange()(newPeriodType);
    }
  };

  const options = [
    {
      value: "HOUR",
      inputDisplay: "HOUR"
    },
    {
      value: "DAY",
      inputDisplay: "DAY"
    },
    {
      value: "WEEK",
      inputDisplay: "WEEK"
    },
    {
      value: "MONTH",
      inputDisplay: "MONTH"
    },
    {
      value: "QUARTER",
      inputDisplay: "QUARTER"
    },
    {
      value: "HALF",
      inputDisplay: "HALF"
    },
    {
      value: "YEAR",
      inputDisplay: "YEAR"
    }
  ];

  const selectedOption = options.find(option =>
    periodType !== undefined
      ? periodType["periodType"] === option.value
      : option.value === ""
  );

  return (
    <EuiForm>
      <EuiFormRow
        fullWidth
        label={
          <FormLabelWithToolTip
            label="Period Type *"
            content="Choose a cycle period type"
          />
        }
        isInvalid={!!errors.periodType}
        error={errors.periodType}
        display="columnCompressed">
        <EuiSuperSelect
          fullWidth
          options={options}
          valueOfSelected={selectedOption ? selectedOption.value : ""}
          onChange={value => setValue(value)}
          itemLayoutAlign="top"
          hasDividers
          isInvalid={!!errors.periodType}
        />
      </EuiFormRow>
    </EuiForm>
  );
};
