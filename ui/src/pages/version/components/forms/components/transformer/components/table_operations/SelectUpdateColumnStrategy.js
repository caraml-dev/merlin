import React from "react";
import { EuiForm, EuiFormRow, EuiSuperSelect } from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@gojek/mlp-ui";

export const SelectUpdateColumnStrategy = ({
  strategy, //current selected strategy (if any)
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const setValue = value => {
    onChange("strategy")(value);
  };

  const options = [
    {
      value: "withoutCondition",
      inputDisplay: "Without Condition (Update column to whole row)"
    },
    {
      value: "withCondition",
      inputDisplay: "With Condition (Update column based on condition)"
    }
  ];

  const selectedOption = options.find(option =>
    strategy !== "" ? option.value === strategy : option.value === ""
  );

  return (
    <EuiForm>
      <EuiFormRow
        fullWidth
        label={
          <FormLabelWithToolTip
            label="Choose Strategy *"
            content="Choose the strategy to update a column"
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
