import {
  EuiButton,
  EuiButtonIcon,
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
import {appConfig} from "../../../../../../../../config"

export const TransformerSimulationInput = ({
  simulationPayload,
  protocol,
  onChange,
  onSubmit,
  errors
}) => {
  const populateSampleRequest = () => {
    const defaultPayload = {"target_name":"probability","prediction_table":{"name":"prediction_table_sample","columns":[{"name":"col1","type":"TYPE_INTEGER"},{"name":"col2","type":"TYPE_DOUBLE"}],"rows":[{"rowId":"row1","values":[{"integer_value":1},{"doubleValue":4.4}]},{"rowId":"row2","values":[{"integer_value":2},{"doubleValue":8.8}]}]},"transformer_input":{"tables":[{"name":"sample_table","columns":[{"name":"col1","type":"TYPE_INTEGER"},{"name":"col2","type":"TYPE_DOUBLE"}],"rows":[{"rowId":"row1","values":[{"integer_value":1},{"double_value":4.4}]},{"rowId":"row2","values":[{"integer_value":2},{"double_value":8.8}]}]}],"variables":[{"name":"var1","type":"TYPE_INTEGER","integer_value":1},{"name":"var2","type":"TYPE_DOUBLE","double_value":1.1}]},"prediction_context":[{"name":"ctx1","type":"TYPE_INTEGER","integer_value":1},{"name":"ctx2","type":"TYPE_STRING","double_value":"SAMPLE"}]}
    onChange("payload", JSON.stringify(defaultPayload, null, 2))
  }

  const populateSampleModelPredictorResponse = () => {
    const defaultPayload = {"prediction_result_table":{"name":"model_prediction_table","columns":[{"name":"probability","type":"TYPE_DOUBLE"}],"rows":[{"rowId":"row1","values":[{"double_value":0.2}]},{"rowId":"row2","values":[{"double_value":0.3}]},{"rowId":"row3","values":[{"double_value":0.4}]},{"rowId":"row4","values":[{"double_value":0.5}]},{"rowId":"row5","values":[{"double_value":0.6}]}]}}
    onChange("mock_response_body", JSON.stringify(defaultPayload, null, 2))
  }
  const clientRequestItems = [
    {
      label: "Client Request Body",
      helpText: protocol === PROTOCOL.HTTP_JSON ?
        "Specify the request body for your standard transformer, must be in JSON object format.": "Specify the request body for your standard transformer, must be in UPI PredictionValuesRequest type",
      value: !!simulationPayload.payload ? simulationPayload.payload : "",
      labelAppend: (
        <span>
          <EuiToolTip content="Go to UPI interface documentation">
            <EuiButtonIcon 
                iconType="documentation"
                href={appConfig.upiDocumentationUrl}
            />
          </EuiToolTip>
          
            
          <EuiButton
              size="s"
              color="primary"
              onClick={populateSampleRequest}
          >
            Populate sample payload
          </EuiButton>
        </span>  
      ),
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
      labelAppend: (
        <span>
          <EuiToolTip content="Go to UPI interface documentation">
            <EuiButtonIcon 
                iconType="documentation"
                href={appConfig.upiDocumentationUrl}
            />
          </EuiToolTip>
            
          <EuiButton
              size="s"
              color="primary"
              onClick={populateSampleModelPredictorResponse}
          >
            Populate sample model predictor payload
          </EuiButton>
        </span>  
      ),
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
                  labelAppend={item.labelAppend}
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
                  labelAppend={item.labelAppend}
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
