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
import { simulationIcon } from "./constants";
import {PROTOCOL} from "../../../../../../../../services/version_endpoint/VersionEndpoint"

export const TransformerSimulationInput = ({
  simulationPayload,
  protocol,
  onChange,
  onSubmit,
  errors
}) => {
  const clientRequestItems = [
    {
      label: "Client Request Body",
      helpText: protocol === PROTOCOL.HTTP_JSON ?
        "Specify the request body for your standard transformer, must be in JSON object format.": "Specify the request body for your standard transformer, must be in UPI PredictionValuesRequest type",
      value: !!simulationPayload.payload ? simulationPayload.payload : "",
      onChange: val => onChange("payload", val),
      error: errors.payload
    },
    {
      label: "Client Request Headers",
      helpText:
        "Specify the request header for your standard transformer, must be in JSON key value format.",
      value: !!simulationPayload.headers ? simulationPayload.headers : "",
      onChange: val => onChange("headers", val),
      error: errors.headers
    }
  ];

  const modelRequestItems = [
    {
      label: "Model Response Body",
      helpText: protocol === PROTOCOL.HTTP_JSON ?
        "Specify the mock for model prediction's response body that can be used for the postprocess, must be in JSON object format." : "Specify the mock for model prediction's response body that can be used for the postprocess, must be in UPI PredictionValuesResponse type",
      value: !!simulationPayload.mock_response_body
        ? simulationPayload.mock_response_body
        : "",
      onChange: val => onChange("mock_response_body", val),
      error: errors.mock_response_body
    },
    {
      label: "Model Response Headers",
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
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem>
        <Panel title="Mock Client Request" contentWidth="92%">
          <EuiSpacer size="s" />

          <EuiFlexGroup direction="column" gutterSize="m">
            {clientRequestItems.map((item, idx) => (
              <EuiFlexItem key={`simulation-request-${idx}`}>
                <EuiFormRow
                  fullWidth
                  label={
                    <EuiToolTip content={item.helpText}>
                      <span>
                        {item.label}{" "}
                        <EuiIcon type="questionInCircle" color="subdued" />
                      </span>
                    </EuiToolTip>
                  }
                  isInvalid={!!item.error}
                  error={item.error}>
                  <EuiTextArea
                    placeholder={item.helpText}
                    value={item.value}
                    onChange={e => item.onChange(e.target.value)}
                    fullWidth
                  />
                </EuiFormRow>
              </EuiFlexItem>
            ))}
          </EuiFlexGroup>
        </Panel>
      </EuiFlexItem>

      <EuiFlexItem>
        <Panel title="Mock Model Response" contentWidth="92%">
          <EuiSpacer size="s" />

          <EuiFlexGroup direction="column" gutterSize="m">
            {modelRequestItems.map((item, idx) => (
              <EuiFlexItem key={`simulation-request-${idx}`}>
                <EuiFormRow
                  fullWidth
                  label={
                    <EuiToolTip content={item.helpText}>
                      <span>
                        {item.label}{" "}
                        <EuiIcon type="questionInCircle" color="subdued" />
                      </span>
                    </EuiToolTip>
                  }
                  isInvalid={!!item.error}
                  error={item.error}>
                  <EuiTextArea
                    placeholder={item.helpText}
                    value={item.value}
                    onChange={e => item.onChange(e.target.value)}
                    fullWidth
                  />
                </EuiFormRow>
              </EuiFlexItem>
            ))}
          </EuiFlexGroup>
        </Panel>
      </EuiFlexItem>

      <EuiFlexItem>
        <EuiFlexGroup justifyContent="center">
          <EuiFlexItem grow={false}>
            <EuiButton
              size="s"
              disabled={false}
              fill
              onClick={onSubmit}
              iconType={simulationIcon}>
              Simulate
            </EuiButton>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
