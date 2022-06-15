import {
  EuiCodeBlock,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiIcon,
  EuiSpacer,
  EuiToolTip
} from "@elastic/eui";
import React from "react";
import { Panel } from "../../../Panel";
import { TransformerSimulationOutputTracing } from "./TransformerSimulationOutputTracing";

export const TransformerSimulationOutput = ({ simulationResponse }) => {
  return simulationResponse.isLoaded ? (
    <Panel title="Simulation Output" contentWidth="92%">
      <EuiSpacer size="s" />
      <EuiFlexGroup direction="column" gutterSize="m">
        <EuiFlexItem>
          <EuiFormRow
            fullWidth
            label={
              <EuiToolTip content="Response of standard transformer simulation">
                <span>
                  Response Body
                  <EuiIcon type="questionInCircle" color="subdued" />
                </span>
              </EuiToolTip>
            }>
            <EuiCodeBlock language="json" overflowHeight={300}>
              {JSON.stringify(simulationResponse.data.response, null, 2)}
            </EuiCodeBlock>
          </EuiFormRow>
        </EuiFlexItem>

        {simulationResponse.data.operation_tracing && (
          <EuiFlexItem>
            <EuiFormRow
              fullWidth
              label="Transformation Operation Trace"
              display="row">
              <TransformerSimulationOutputTracing
                tracingDetails={simulationResponse.data.operation_tracing}
              />
            </EuiFormRow>
          </EuiFlexItem>
        )}
      </EuiFlexGroup>
    </Panel>
  ) : (
    <></>
  );
};
