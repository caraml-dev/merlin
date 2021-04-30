import React, { Fragment } from "react";
import { EuiSpacer } from "@elastic/eui";

export const DeploymentSummary = () => {
  return (
    <Fragment>
      <p>
        You're about to deploy a new model version <b>WUT</b> into <b>WUT</b>{" "}
        environment.
      </p>

      <EuiSpacer size="s" />
    </Fragment>
  );
};
