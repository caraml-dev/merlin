import {
  EuiButton,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiIcon,
  EuiSpacer,
  EuiTextArea,
  EuiToolTip
} from "@elastic/eui";
import React from "react";
import { Panel } from "../../../Panel";

export const TransformerSimulationInput = ({
  simulationPayload,
  onChange,
  onSubmit,
  errors
}) => {
  const requestItems = [
    {
      label: "Request Payload",
      helpText:
        "Specify the request payload for your standard transformer, must be in JSON object format.",
      value: !!simulationPayload.payload ? simulationPayload.payload : "",
      onChange: val => onChange("payload", val),
      error: errors.payload
    },
    {
      label: "Request Headers",
      helpText:
        "Specify the request header for your standard transformer, must be in JSON format key value.",
      value: !!simulationPayload.headers ? simulationPayload.headers : "",
      onChange: val => onChange("headers", val),
      error: errors.headers
    },
    {
      label: "Mock Model Response Body",
      helpText:
        "Specify the mock for model prediction's response body, must be in JSON object format.",
      value: !!simulationPayload.mock_response_body
        ? simulationPayload.mock_response_body
        : "",
      onChange: val => onChange("mock_response_body", val),
      error: errors.mock_response_body
    },
    {
      label: "Mock Model Response Headers",
      helpText:
        "Specify the mock for model prediction's response headers, must be in JSON format key value.",
      value: !!simulationPayload.mock_response_headers
        ? simulationPayload.mock_response_headers
        : "",
      onChange: val => onChange("mock_response_headers", val),
      error: errors.mock_response_headers
    }
  ];

  return (
    <Panel title="Simulation Input" contentWidth="92%">
      <EuiSpacer size="s" />

      <EuiFlexGroup direction="column" gutterSize="m">
        {requestItems.map((item, idx) => (
          <EuiFlexItem key={`simulation-request-${idx}`}>
            <EuiFormRow
              fullWidth
              label={
                <EuiToolTip content={item.helpText}>
                  <span>
                    {item.label}
                    <EuiIcon type="questionInCircle" color="subdued" />
                  </span>
                </EuiToolTip>
              }
              display="columnCompressed"
              isInvalid={!!item.error}
              error={item.error}>
              <EuiTextArea
                placeholder="Request payload in JSON"
                value={item.value}
                onChange={e => item.onChange(e.target.value)}
                fullWidth
              />
            </EuiFormRow>
          </EuiFlexItem>
        ))}
      </EuiFlexGroup>

      <EuiSpacer size="s" />

      <EuiFlexGroup direction="row" justifyContent="flexEnd">
        <EuiFlexItem grow={false}>
          <EuiButton size="s" disabled={false} onClick={onSubmit}>
            Simulate
          </EuiButton>
        </EuiFlexItem>
      </EuiFlexGroup>
    </Panel>
  );
};
