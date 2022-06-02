import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiTextArea,
  EuiButton,
  EuiToolTip,
  EuiSpacer,
  EuiIcon
} from "@elastic/eui";
import React from "react";
import { Panel } from "../components/Panel";

export const TransformerSimulationRequest = ({
  simulationPayload,
  onChange,
  onSubmit
}) => {
  return (
    <Panel title="Request" contentWidth="92%">
      <EuiSpacer size="s" />
      <EuiFlexGroup direction="column" gutterSize="m">
        <EuiFlexItem>
          <EuiFormRow
            fullWidth
            label={
              <EuiToolTip content="Specify the request payload for your standard transformer, must be in JSON format">
                <span>
                  Payload
                  <EuiIcon type="questionInCircle" color="subdued" />
                </span>
              </EuiToolTip>
            }
            display="columnCompressed">
            <EuiTextArea
              placeholder="Request payload in JSON"
              value={
                !!simulationPayload.payload ? simulationPayload.payload : ""
              }
              onChange={e => onChange("payload", e.target.value)}
            />
          </EuiFormRow>
        </EuiFlexItem>
        <EuiFlexItem>
          <EuiFormRow
            fullWidth
            label={
              <EuiToolTip content="Specify the request header for your standard transformer, must be in JSON format key value">
                <span>
                  Headers
                  <EuiIcon type="questionInCircle" color="subdued" />
                </span>
              </EuiToolTip>
            }
            display="columnCompressed">
            <EuiTextArea
              placeholder="Request headers in JSON"
              value={
                !!simulationPayload.headers ? simulationPayload.headers : ""
              }
              onChange={e => onChange("headers", e.target.value)}
            />
          </EuiFormRow>
        </EuiFlexItem>
        <EuiFlexItem>
          <EuiFormRow
            fullWidth
            label={
              <EuiToolTip content="Specify the mock for model prediction's response headers, must be in JSON format key value">
                <span>
                  Mock Prediction Payload
                  <EuiIcon type="questionInCircle" color="subdued" />
                </span>
              </EuiToolTip>
            }
            display="columnCompressed">
            <EuiTextArea
              placeholder="Mock response body for model prediction"
              value={
                !!simulationPayload.mock_response_body
                  ? simulationPayload.mock_response_body
                  : ""
              }
              onChange={e => onChange("mock_response_body", e.target.value)}
            />
          </EuiFormRow>
        </EuiFlexItem>
        <EuiFlexItem>
          <EuiFormRow
            fullWidth
            label={
              <EuiToolTip content="Specify the mock for model prediction's response headers , must be in JSON format key value">
                <span>
                  Mock Prediction Headers
                  <EuiIcon type="questionInCircle" color="subdued" />
                </span>
              </EuiToolTip>
            }
            display="columnCompressed">
            <EuiTextArea
              placeholder="Mock response headers for model prediction"
              value={
                !!simulationPayload.mock_response_headers
                  ? simulationPayload.mock_response_headers
                  : ""
              }
              onChange={e => onChange("mock_response_headers", e.target.value)}
            />
          </EuiFormRow>
        </EuiFlexItem>
      </EuiFlexGroup>
      <EuiSpacer size="s" />
      <EuiFlexGroup direction="row" justifyContent="flexEnd">
        <EuiFlexItem grow={false}>
          <EuiButton
            size="s"
            color="primary"
            disabled={false}
            onClick={onSubmit}>
            Try
          </EuiButton>
        </EuiFlexItem>
      </EuiFlexGroup>
    </Panel>
  );
};
