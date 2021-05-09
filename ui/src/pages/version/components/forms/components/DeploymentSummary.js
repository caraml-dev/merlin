import React, { Fragment } from "react";
import { EuiSpacer } from "@elastic/eui";

export const DeploymentSummary = ({
  actionTitle = "Deploy",
  modelName,
  versionId
}) => {
  return (
    <Fragment>
      <p>
        You're about to {actionTitle.toLowerCase()} a new endpoint for model{" "}
        <b>{modelName}</b> version <b>{versionId}</b>.
      </p>
      <EuiSpacer size="s" />
    </Fragment>
  );
};
