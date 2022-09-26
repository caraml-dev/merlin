import React, { useContext } from "react";
import {
  EuiForm,
  EuiFormRow,
  EuiIcon,
  EuiSpacer,
  EuiSuperSelect,
  EuiToolTip,
} from "@elastic/eui";
import sortBy from "lodash/sortBy";
import { Panel } from "./Panel";
import EnvironmentsContext from "../../../../../providers/environments/context";
import { EnvironmentDropdownOption } from "./EnvironmentDropdownOption";
import { DeploymentModeDropdown } from "./DeploymentModeDropdown";
import { AutoscalingPolicyFormGroup } from "./AutoscalingPolicyFormGroup";
import { ProtocolDropdown } from "./ProtocolDropdown";

const isEnvironmentOptionDisabled = (endpoint) => {
  return (
    endpoint &&
    (endpoint.status === "serving" ||
      endpoint.status === "running" ||
      endpoint.status === "pending")
  );
};

const getEndpointByEnvironment = (version, environmentName) => {
  const ve = version.endpoints.find((ve) => {
    return ve.environment_name === environmentName;
  });
  return ve;
};

const DEFAULT_AUTOSCALING_POLICY = {
  metrics_type: "concurrency",
  target_value: 1.0,
};

/**
 * Deployment config panel provides user interface to configure deployment-wide configurations that affect both
 * Model and Transformer
 * @param {*} environment. Selected environment to deploy to
 * @param {*} endpoint. Existing endpoint
 * @param {*} version. Model version to be deployed
 * @param {*} onChange. Callback to be made when configuration is changed
 * @param {*} errors. Why pass errors?
 * @param {*} isEnvironmentDisabled. Disable deployment to the environment if the flag is true.
 */
export const DeploymentConfigPanel = ({
  environment,
  endpoint,
  version,
  onChange,
  errors = {},
  isEnvironmentDisabled = false,
}) => {
  const environments = useContext(EnvironmentsContext);

  const environmentOptions = sortBy(environments, "name").map((environment) => {
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
      ),
    };
  });

  const onEnvironmentChange = (value) => {
    onChange("environment_name")(value);

    const versionEndpoint = getEndpointByEnvironment(version, value);

    if (versionEndpoint) {
      Object.keys(versionEndpoint).forEach((key) => {
        onChange(key)(versionEndpoint[key]);
      });
    }
  };

  const onDeploymentModeChange = (deploymentMode) => {
    onChange("deployment_mode")(deploymentMode);
  };

  const onAutoscalingPolicyChange = (policy) => {
    onChange("autoscaling_policy")(policy);
  };

  const onProtocolChange = (protocol) => {
    onChange("protocol")(protocol);
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
          display="row"
        >
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
            <EuiToolTip content="Protocol affects the server type and interface exposed by the deployment">
              <span>
                Protocol * <EuiIcon type="questionInCircle" color="subdued" />
              </span>
            </EuiToolTip>
          }
          display="row"
        >
          <ProtocolDropdown
            fullWidth
            endpointStatus={endpoint.status}
            protocol={endpoint.protocol}
            onProtocolChange={onProtocolChange}
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
          display="row"
        >
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
