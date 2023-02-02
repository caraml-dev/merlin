import React, { Fragment } from "react";
import { EuiSpacer } from "@elastic/eui";
import { CostEstimationPanel } from "./CostEstimationPanel";

export const DeploymentSummary = ({
  actionTitle = "Deploy",
  modelName,
  versionId,
  versionEndpoint,
}) => {
  return (
    <Fragment>
      <p>
        You're about to {actionTitle.toLowerCase()} a new endpoint for model{" "}
        <b>{modelName}</b> version <b>{versionId}</b>.
      </p>
      <EuiSpacer size="m" />
      <CostEstimationPanel versionEndpoint={versionEndpoint} />
    </Fragment>
  );
};
