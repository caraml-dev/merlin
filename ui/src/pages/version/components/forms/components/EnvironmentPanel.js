import React, { useContext } from "react";
import {
  EuiForm,
  EuiFormRow,
  EuiIcon,
  EuiSuperSelect,
  EuiToolTip
} from "@elastic/eui";
import sortBy from "lodash/sortBy";
import { Panel } from "./Panel";
import EnvironmentsContext from "../../../../../providers/environments/context";

export const EnvironmentPanel = ({
  environment,
  version,
  onChange,
  errors = {},
  isEnvironmentDisabled = false
}) => {
  const environments = useContext(EnvironmentsContext);

  const environmentOptions = sortBy(environments, "name").map(environment => ({
    value: environment.name,
    inputDisplay: environment.name
  }));

  const onEnvironmentChange = value => {
    onChange("environment_name")(value);

    const existingEndpoint = version.endpoints.find(
      e => e.environment_name === value
    );
    if (existingEndpoint) {
      Object.keys(existingEndpoint).forEach(key => {
        onChange(key)(existingEndpoint[key]);
      });
    }
  };

  return (
    <Panel title="Environment">
      <EuiForm>
        <EuiFormRow
          fullWidth
          label={
            <EuiToolTip content="Specify the target environment your model version will be deployed to.">
              <span>
                Environment *{" "}
                <EuiIcon type="questionInCircle" color="subdued" />
              </span>
            </EuiToolTip>
          }
          isInvalid={!!errors.environment_name}
          error={errors.environment_name}
          display="row">
          <EuiSuperSelect
            fullWidth
            options={environmentOptions}
            valueOfSelected={environment}
            onChange={onEnvironmentChange}
            isInvalid={!!errors.environment_name}
            itemLayoutAlign="top"
            hasDividers
            disabled={isEnvironmentDisabled}
          />
        </EuiFormRow>
      </EuiForm>
    </Panel>
  );
};
