import React, { useContext } from "react";
import {
  EuiForm,
  EuiFormRow,
  EuiIcon,
  EuiSpacer,
  EuiSuperSelect,
  EuiToolTip
} from "@elastic/eui";
import sortBy from "lodash/sortBy";
import { Panel } from "./Panel";
import EnvironmentsContext from "../../../../../providers/environments/context";
import { EnvironmentDropdownOption } from "./EnvironmentDropdownOption";
import { DeploymentModeDropdown } from "./DeploymentModeDropdown";
import { AutoscalingPolicyFormGroup } from "./AutoscalingPolicyFormGroup";

const isEnvironmentOptionDisabled = endpoint => {
  return (
    endpoint &&
    (endpoint.status === "serving" ||
      endpoint.status === "running" ||
      endpoint.status === "pending")
  );
};

const getEndpointByEnvironment = (version, environmentName) => {
  const ve = version.endpoints.find(ve => {
    return ve.environment_name === environmentName;
  });
  return ve;
};

const DEFAULT_AUTOSCALING_POLICY = {
  metrics_type: "concurrency",
  target_value: 1.0
};

export const DeploymentConfigPanel = ({
  environment,
  endpoint,
  version,
  onChange,
  errors = {},
  isEnvironmentDisabled = false
}) => {
  const environments = useContext(EnvironmentsContext);

  const environmentOptions = sortBy(environments, "name").map(environment => {
    const versionEndpoint = getEndpointByEnvironment(version, environment.name);
    const isDisabled = isEnvironmentOptionDisabled(versionEndpoint);

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

  const onDeploymentModeChange = deploymentMode => {
    onChange("deployment_mode")(deploymentMode);
  };

  const onAutoscalingPolicyChange = policy => {
    onChange("autoscaling_policy")(policy);
  };

  return (
    <Panel title="Deployment Configuration">
      <EuiSpacer></EuiSpacer>
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
        <EuiFormRow
          fullWidth
          label={
            <EuiToolTip content="Specify the deployment mode of your model version">
              <span>
                Deployment Mode *{" "}
                <EuiIcon type="questionInCircle" color="subdued" />
              </span>
            </EuiToolTip>
          }
          display="row">
          <DeploymentModeDropdown
            fullWidth
            endpointStatus={endpoint.status}
            deploymentMode={endpoint.deployment_mode}
            onDeploymentModeChange={onDeploymentModeChange}
          />
        </EuiFormRow>
        <EuiSpacer></EuiSpacer>
        <AutoscalingPolicyFormGroup
          fullWidth
          deploymentMode={endpoint.deployment_mode}
          autoscalingPolicy={
            endpoint.autoscaling_policy || DEFAULT_AUTOSCALING_POLICY
          }
          onAutoscalingPolicyChange={onAutoscalingPolicyChange}
        />
      </EuiForm>
    </Panel>
  );
};
