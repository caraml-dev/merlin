import React, { Fragment, useContext } from "react";
import {
  EuiForm,
  EuiFormRow,
  EuiIcon,
  EuiSuperSelect,
  EuiText,
  EuiTextColor,
  EuiToolTip
} from "@elastic/eui";
import sortBy from "lodash/sortBy";
import { Panel } from "./Panel";
import EnvironmentsContext from "../../../../../providers/environments/context";

const EnvironmentDropdownOption = ({
  environment,
  version,
  endpoint,
  disabled
}) => {
  const option = (
    <Fragment>
      <strong>{environment.name}</strong>
      <EuiText size="s">
        <EuiTextColor color="subdued">{environment.cluster}</EuiTextColor>
      </EuiText>
      {endpoint && (
        <EuiText size="s">
          <EuiTextColor color="subdued">{endpoint.status}</EuiTextColor>
        </EuiText>
      )}
    </Fragment>
  );

  return disabled ? (
    <EuiToolTip
      position="left"
      content={
        <Fragment>
          <p>
            Model version {version.id} already has a <b>{endpoint.status}</b>{" "}
            endpoint on <b>{environment.name}</b>.
          </p>
          <p>
            <br />
          </p>
          <p>
            If you want to do redeployment, go to model version details page and
            click Redeploy button.
          </p>
        </Fragment>
      }>
      {option}
    </EuiToolTip>
  ) : (
    option
  );
};

const isOptionDisabled = endpoint => {
  return (
    endpoint &&
    (endpoint.status === "serving" ||
      endpoint.status === "running" ||
      endpoint.status === "pending")
  );
};

const getEndpointByEnvironment = (version, environmentName) => {
  return version.endpoints.find(ve => {
    return ve.environment_name === environmentName;
  });
};

export const EnvironmentPanel = ({
  environment,
  version,
  onChange,
  errors = {},
  isEnvironmentDisabled = false
}) => {
  const environments = useContext(EnvironmentsContext);

  const environmentOptions = sortBy(environments, "name").map(environment => {
    const versionEndpoint = getEndpointByEnvironment(version, environment.name);
    const isDisabled = isOptionDisabled(versionEndpoint);

    return {
      value: environment.name,
      disabled: isDisabled,
      inputDisplay: environment.name,
      dropdownDisplay: (
        <EnvironmentDropdownOption
          environment={environment}
          version={version}
          endpoint={versionEndpoint}
          disabled={isDisabled}
        />
      )
    };
  });

  const onEnvironmentChange = value => {
    onChange("environment_name")(value);

    const versionEndpoint = getEndpointByEnvironment(version, value);
    if (versionEndpoint) {
      Object.keys(versionEndpoint).forEach(key => {
        onChange(key)(versionEndpoint[key]);
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
