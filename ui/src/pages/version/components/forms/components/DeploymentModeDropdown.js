import React, { Fragment } from "react";
import { EuiSuperSelect, EuiText } from "@elastic/eui";

export const DeploymentModeDropdown = ({
  endpointStatus,
  deploymentMode,
  onDeploymentModeChange
}) => {
  // Users should not be able to change the deployment mode when the model endpoint status is serving
  const isDisabled = endpointStatus => {
    return endpointStatus === "serving";
  };

  const onChange = value => {
    onDeploymentModeChange(value);
  };

  return (
    <EuiSuperSelect
      fullWidth
      options={[
        {
          value: "serverless",
          inputDisplay: "Serverless",
          dropdownDisplay: (
            <Fragment>
              <strong>Serverless</strong>
              <EuiText size="s" color="subdued">
                <p className="euiTextColor--subdued">
                  More advanced autoscaling and ability to scale down to 0
                  replica
                </p>
              </EuiText>
            </Fragment>
          )
        },
        {
          value: "raw_deployment",
          inputDisplay: "Raw Deployment",
          dropdownDisplay: (
            <Fragment>
              <strong>Raw Deployment</strong>
              <EuiText size="s" color="subdued">
                <p className="euiTextColor--subdued">
                  Less infrastructure overhead and more performant
                </p>
              </EuiText>
            </Fragment>
          )
        }
      ]}
      valueOfSelected={deploymentMode}
      onChange={onChange}
      disabled={isDisabled(endpointStatus)}
      hasDividers
    />
  );
};
