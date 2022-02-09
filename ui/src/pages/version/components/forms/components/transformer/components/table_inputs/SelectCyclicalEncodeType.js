import React from "react";
import { EuiForm, EuiFormRow, EuiSuperSelect } from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@gojek/mlp-ui";

export const SelectCyclicalEncodeType = ({
  cyclicType, //type of input to be encoded
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const setValue = value => {
    //set new value only if there is changes
    if (cyclicType[value] === undefined) {
      let newEncodeType = {};

      switch (value) {
        case "byEpochTime":
          newEncodeType[value] = {
            periodType: "DAY"
          };
          break;
        case "byRange":
          newEncodeType[value] = {
            min: "0", //str to make it appear on UI, it will be converted to float later
            max: 7
          };
          break;
        default:
          break;
      }
      onChange()(newEncodeType);
    }
  };

  const options = [
    {
      value: "byEpochTime",
      inputDisplay: "Epoch Time"
    },
    {
      value: "byRange",
      inputDisplay: "Range"
    }
  ];

  const selectedOption = options.find(option =>
    cyclicType !== undefined
      ? cyclicType[option.value] !== undefined
      : option.value === ""
  );

  return (
    <EuiForm>
      <EuiFormRow
        fullWidth
        label={
          <FormLabelWithToolTip
            label="Encode by *"
            content="Choose input type to encode. Epoch time: unix epoch time; Range: Range of input values for representing 1 period"
          />
        }
        isInvalid={!!errors.cyclicType}
        error={errors.cyclicType}
        display="columnCompressed">
        <EuiSuperSelect
          fullWidth
          options={options}
          valueOfSelected={selectedOption ? selectedOption.value : ""}
          onChange={value => setValue(value)}
          itemLayoutAlign="top"
          hasDividers
          isInvalid={!!errors.cyclicType}
        />
      </EuiFormRow>
    </EuiForm>
  );
};
