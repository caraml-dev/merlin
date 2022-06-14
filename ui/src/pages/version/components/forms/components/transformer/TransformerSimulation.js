import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { FormContext } from "@gojek/mlp-ui";
import React, { useContext, useState } from "react";
import { useMerlinApi } from "../../../../../../hooks/useMerlinApi";
import { TransformerSimulationInput } from "./components/simulation/TransformerSimulationInput";
import { TransformerSimulationOutput } from "./components/simulation/TransformerSimulationOutput";

class SimulationPayload {
  constructor() {
    this.payload = undefined;
    this.headers = undefined;
    this.config = undefined;
    this.model_prediction_config = undefined;
  }
}

const convertToJson = (val) => {
  if (val === null || val === undefined || val === "") {
    return undefined;
  }

  try {
    return JSON.parse(val);
  } catch (e) {
    throw new Error(`Unable to parse JSON object: ${e.message}`);
  }
};

export const TransformerSimulation = () => {
  const [simulationPayload, setSimulationPayload] = useState(
    new SimulationPayload()
  );
  const [errors, setErrors] = useState({});

    const [simulationResponse, submitForm] = useMerlinApi(
    `/standard_transformer/simulate`,
    { method: "POST" },
    {},
    false
  );

  const {
    data: {
      transformer: { config },
    },
  } = useContext(FormContext);

  const onSubmit = () => {
    let errors = {};

    Object.keys(simulationPayload).forEach((name) => {
      try {
        convertToJson(simulationPayload[name]);
      } catch (e) {
        errors[name] = e.message;
      }
    });

    setErrors(errors);

    if (Object.keys(errors).length > 0) {
      return;
    }

    submitForm({
      body: JSON.stringify({
        config: config,
        payload: convertToJson(simulationPayload.payload),
        headers: convertToJson(simulationPayload.headers),
        model_prediction_config: {
          mock_response: {
            body: convertToJson(simulationPayload.mock_response_body),
            headers: convertToJson(simulationPayload.mock_response_headers),
          },
        },
      }),
    });
  };

  const onChange = (field, value) => {
    setSimulationPayload({ ...simulationPayload, [field]: value });
  };

  return (
    <EuiFlexGroup gutterSize="m" direction="column">
      <EuiFlexItem>
        <TransformerSimulationInput
          simulationPayload={simulationPayload}
          onChange={onChange}
          onSubmit={onSubmit}
          errors={errors}
        />
      </EuiFlexItem>

      <EuiFlexItem>
        <TransformerSimulationOutput simulationResponse={simulationResponse} />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
