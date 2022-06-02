import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import React, { useState, useContext } from "react";
import { useMerlinApi } from "../../../../../hooks/useMerlinApi";
import { FormContext } from "@gojek/mlp-ui";
import { TransformerSimulationRequest } from "./TransformerSimulationRequest";
import { TransformerSimulationOutput } from "./TransformerSimulationOutput";

class SimulationPayload {
  constructor() {
    this.payload = undefined;
    this.headers = undefined;
    this.config = undefined;
    this.model_prediction_config = undefined;
  }
}

const convertToJson = val => {
  try {
    return JSON.parse(val);
  } catch (error) {
    return undefined;
  }
};

export const TransformerSimulation = () => {
  const [simulationPayload, setSimulationPayload] = useState(
    new SimulationPayload()
  );
  const [submissionResponse, submitForm] = useMerlinApi(
    `/standard_transformer/simulate`,
    { method: "POST" },
    {},
    false
  );

  const {
    data: {
      transformer: { config }
    }
  } = useContext(FormContext);

  const onSubmit = () => {
    submitForm({
      body: JSON.stringify({
        config: config,
        payload: convertToJson(simulationPayload.payload),
        headers: convertToJson(simulationPayload.headers),
        model_prediction_config: {
          mock: {
            body: convertToJson(simulationPayload.mock_response_body),
            headers: convertToJson(simulationPayload.mock_response_headers)
          }
        }
      })
    });
  };

  const onChange = (field, value) => {
    setSimulationPayload({ ...simulationPayload, [field]: value });
  };

  return (
    <>
      <EuiFlexGroup gutterSize="m" direction="column">
        <EuiFlexItem grow={7}>
          <div id="simulation-request">
            <TransformerSimulationRequest
              simulationPayload={simulationPayload}
              onChange={onChange}
              onSubmit={onSubmit}
            />
          </div>
        </EuiFlexItem>
        <TransformerSimulationOutput simulationResponse={submissionResponse} />
      </EuiFlexGroup>
    </>
  );
};
