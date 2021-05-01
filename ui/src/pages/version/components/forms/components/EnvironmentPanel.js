import React, { useContext } from "react";
import {
  EuiForm,
  EuiFormRow,
  EuiIcon,
  EuiSuperSelect,
  EuiToolTip
} from "@elastic/eui";
import sortBy from "lodash/sortBy";
import EnvironmentsContext from "../../../../../providers/environments/context";
import { Panel } from "./Panel";

export const EnvironmentPanel = ({ environment, onChange, errors = {} }) => {
  const environments = useContext(EnvironmentsContext);

  const environmentOptions = sortBy(environments, "name").map(environment => ({
    value: environment.name,
    inputDisplay: environment.name
  }));

  return (
    <Panel title="Environment">
      <EuiForm>
        <EuiFormRow
          fullWidth
          label={
            <EuiToolTip content="Specify the target environment your model version will be deployed to.">
              <span>
                Environment* <EuiIcon type="questionInCircle" color="subdued" />
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
            onChange={onChange("environment_name")}
            isInvalid={!!errors.environment_name}
            itemLayoutAlign="top"
            hasDividers
          />
        </EuiFormRow>
      </EuiForm>
    </Panel>
  );
};
