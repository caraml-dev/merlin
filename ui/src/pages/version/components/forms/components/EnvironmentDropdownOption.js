import React, { Fragment } from "react";
import { EuiText, EuiTextColor, EuiToolTip } from "@elastic/eui";

export const EnvironmentDropdownOption = ({
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
