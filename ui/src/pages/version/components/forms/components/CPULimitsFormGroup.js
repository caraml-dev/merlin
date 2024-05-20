import React, { Fragment } from "react";
import { FormLabelWithToolTip, useOnChangeHandler } from "@caraml-dev/ui-lib";
import { EuiDescribedFormGroup, EuiFieldText, EuiFormRow } from "@elastic/eui";


export const CPULimitsFormGroup = ({
  resourcesConfig,
  onChangeHandler,
  errors = {},
}) => {
  const {onChange} = useOnChangeHandler(onChangeHandler);

  return (
    <EuiDescribedFormGroup
      title={<p>CPU Limits</p>}
      description={
        <Fragment>
          CPU limits are not set at the platform level. Use this field to limit the
          amount of CPU that the model is able to consume.
        </Fragment>
      }
      fullWidth
    >
      <EuiFormRow
        label={
          <FormLabelWithToolTip
            label="CPU Limits"
            content="Specify the maximum amount of CPU available for your model"
          />
        }
        isInvalid={!!errors.cpu_limit}
        error={errors.cpu_limit}
        fullWidth
      >
        <EuiFieldText
          placeholder="500m"
          value={resourcesConfig?.cpu_limit}
          onChange={(e) => onChange("cpu_limit")(e.target.value)}
          isInvalid={!!errors.cpu_limit}
          name="cpu_limit"
          fullWidth
        />
      </EuiFormRow>
    </EuiDescribedFormGroup>
  )
}