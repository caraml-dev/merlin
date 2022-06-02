import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiToolTip,
  EuiSpacer,
  EuiIcon,
  EuiCodeBlock
} from "@elastic/eui";
import React from "react";
import { Panel } from "../components/Panel";
import { TransformerSimulationOutputTracing } from "./TransformerSimulationOutputTracing";

export const TransformerSimulationOutput = ({ simulationResponse }) => {
  return simulationResponse.isLoaded ? (
    <EuiFlexItem grow={7}>
      <div id="simulation-output">
        <Panel title="Output" contentWidth="92%">
          <EuiSpacer size="s" />
          <EuiFlexGroup direction="column" gutterSize="m">
            <EuiFlexItem>
              <EuiFormRow
                fullWidth
                label={
                  <EuiToolTip content="Response of standard transformer simulation">
                    <span>
                      Response
                      <EuiIcon type="questionInCircle" color="subdued" />
                    </span>
                  </EuiToolTip>
                }
                display="columnCompressed">
                <EuiCodeBlock
                  language="json"
                  style={{ width: "100px" }}
                  overflowHeight={300}
                  isVirtualized={true}>
                  {JSON.stringify(simulationResponse.data.response, null, 2)}
                </EuiCodeBlock>
              </EuiFormRow>
            </EuiFlexItem>
            <EuiFlexItem>
              <EuiFormRow fullWidth label="Operation Tracing" display="row">
                <TransformerSimulationOutputTracing
                  tracingDetails={simulationResponse.data.operation_tracing}
                />
              </EuiFormRow>
            </EuiFlexItem>
          </EuiFlexGroup>
        </Panel>
      </div>
    </EuiFlexItem>
  ) : (
    <></>
  );
};
