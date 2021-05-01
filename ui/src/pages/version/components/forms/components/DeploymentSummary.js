import React, { Fragment } from "react";
import { EuiSpacer } from "@elastic/eui";

export const DeploymentSummary = ({ modelName, versionId }) => {
  return (
    <Fragment>
      <p>
        You're about to deploy a new endpoint for model <b>{modelName}</b>{" "}
        version <b>{versionId}</b>.
      </p>
      <EuiSpacer size="s" />
    </Fragment>
  );
};
