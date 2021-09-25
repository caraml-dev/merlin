import React from "react";
import { EuiForm, EuiFormRow, EuiSuperSelect } from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@gojek/mlp-ui";

export const SelectEncoder = ({
  name,
  encoder, //current selected encoder (if any)
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const setValue = value => {
    let newEncoder = {
      name: name || "",
      encoder: value
    };

    //storage for info about transformation (eg selected col)
    switch (value) {
      case "ordinalEncoderConfig":
        newEncoder[value] = {};
        break;
      default:
        break;
    }
    onChange()(newEncoder);
  };

  const options = [
    {
      value: "ordinalEncoderConfig",
      inputDisplay: "Ordinal Encoder"
    }
  ];

  const selectedOption = options.find(option =>
    encoder !== "" ? option.value === encoder : option.value === ""
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
        isInvalid={!!errors.encoder}
        error={errors.encoder}
        display="columnCompressed">
        <EuiSuperSelect
          fullWidth
          options={options}
          valueOfSelected={selectedOption ? selectedOption.value : ""}
          onChange={value => setValue(value)}
          itemLayoutAlign="top"
          hasDividers
          isInvalid={!!errors.encoder}
        />
      </EuiFormRow>
    </EuiForm>
  );
};
