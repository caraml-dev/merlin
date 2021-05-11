import React from "react";
import { EuiForm, EuiFormRow, EuiSuperSelect } from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@gojek/mlp-ui";

export const SelectTableJoin = ({ how, onChangeHandler, errors = {} }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const setValue = value => {
    onChange("how")(value);
    if (value === "CROSS" || value === "CONCAT") {
      onChange("onColumn")(undefined);
    }
  };

  const options = [
    {
      value: "LEFT",
      inputDisplay: "LEFT"
    },
    {
      value: "RIGHT",
      inputDisplay: "RIGHT"
    },
    {
      value: "INNER",
      inputDisplay: "INNER"
    },
    {
      value: "OUTER",
      inputDisplay: "OUTER"
    },
    {
      value: "CROSS",
      inputDisplay: "CROSS"
    },
    {
      value: "CONCAT",
      inputDisplay: "CONCAT"
    }
  ];

  const selectedOption = options.find(option =>
    how !== "" ? option.value === how : option.value === ""
  );

  return (
    <EuiForm>
      <EuiFormRow
        fullWidth
        label={
          <FormLabelWithToolTip
            label="Join Method *"
            content="Choose the join method"
          />
        }
        isInvalid={!!errors.how}
        error={errors.how}>
        <EuiSuperSelect
          fullWidth
          options={options}
          valueOfSelected={selectedOption ? selectedOption.value : ""}
          onChange={value => setValue(value)}
          itemLayoutAlign="top"
          hasDividers
          isInvalid={!!errors.how}
        />
      </EuiFormRow>
    </EuiForm>
  );
};
