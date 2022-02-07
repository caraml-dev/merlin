import React from "react";
import { EuiForm, EuiFormRow, EuiSuperSelect } from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@gojek/mlp-ui";

export const SelectEncoder = ({
  encoderConfig, //current encoder config (if any)
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const setValue = value => {
    //set new value only if there is changes
    if (encoderConfig[value] === undefined) {
      let newEncoder = {
        name: encoderConfig.name || ""
      };

      switch (value) {
        case "ordinalEncoderConfig":
          newEncoder[value] = {
            mapping: {},
            targetValueType: "INT"
          };
          break;
        case "cyclicalEncoderConfig":
          newEncoder[value] = {
            byEpochTime: {
              period: "DAY"
            }
          };
          break;
        default:
          break;
      }
      onChange()(newEncoder);
    }
  };

  const options = [
    {
      value: "ordinalEncoderConfig",
      inputDisplay: "Ordinal Encoder"
    },
    {
      value: "cyclicalEncoderConfig",
      inputDisplay: "Cyclical Encoder"
    }
  ];

  const selectedOption = options.find(option =>
    encoderConfig !== undefined
      ? encoderConfig[option.value] !== undefined
      : option.value === ""
  );

  return (
    <EuiForm>
      <EuiFormRow
        fullWidth
        label={
          <FormLabelWithToolTip
            label="Encoder Type *"
            content="Choose an encoder type"
          />
        }
        isInvalid={!!errors.encoderConfig}
        error={errors.encoderConfig}
        display="columnCompressed">
        <EuiSuperSelect
          fullWidth
          options={options}
          valueOfSelected={selectedOption ? selectedOption.value : ""}
          onChange={value => setValue(value)}
          itemLayoutAlign="top"
          hasDividers
          isInvalid={!!errors.encoderConfig}
        />
      </EuiFormRow>
    </EuiForm>
  );
};
