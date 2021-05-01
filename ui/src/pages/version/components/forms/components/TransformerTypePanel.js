import React from "react";
import { EuiForm, EuiFormRow, EuiSuperSelect } from "@elastic/eui";
import {
  DescribedFormGroup,
  FormLabelWithToolTip,
  useOnChangeHandler
} from "@gojek/mlp-ui";
import { Panel } from "./Panel";

export const TransformerTypePanel = ({
  type,
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const setValue = value => {
    if (value === "disabled") {
      onChange("enabled")(false);
      onChange("transformer_type")("");
    } else {
      onChange("enabled")(true);
      onChange("transformer_type")(value);
    }
  };

  const options = [
    {
      value: "disabled",
      inputDisplay: "None",
      description:
        "No transformation gonna happen. Original request will be sent to the model."
    },
    {
      value: "standard",
      inputDisplay: "Standard Transformer",
      description:
        "Standard Transformer enables you to specify the preprocess and postprocess steps."
    },
    {
      value: "custom",
      inputDisplay: "Custom Transformer",
      description:
        "Merlin will deploy your Docker image as the transformer. The incoming request will be sent to it first for the preprocess."
    },
    {
      value: "feast",
      inputDisplay: "Feast Enricher",
      description:
        "Feast Enricher enriches the incoming request with Feast features."
    }
  ];

  const selectedOption = options.find(option =>
    type !== "" ? option.value === type : option.value === "disabled"
  );

  return (
    <Panel title="Transformer">
      <EuiForm>
        <DescribedFormGroup description={(selectedOption || {}).description}>
          <EuiFormRow
            fullWidth
            label={
              <FormLabelWithToolTip
                label="Transformer Type*"
                content="Select the type of transformer to be deployed alongside your model"
              />
            }
            isInvalid={!!errors.transformer_type}
            error={errors.transformer_type}
            display="row">
            <EuiSuperSelect
              fullWidth
              options={options}
              valueOfSelected={
                selectedOption ? selectedOption.value : "disabled"
              }
              onChange={value => setValue(value)}
              itemLayoutAlign="top"
              hasDividers
              isInvalid={!!errors.transformer_type}
            />
          </EuiFormRow>
        </DescribedFormGroup>
      </EuiForm>
    </Panel>
  );
};
