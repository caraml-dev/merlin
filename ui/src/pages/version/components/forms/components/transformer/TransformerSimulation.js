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

const convertToJson = val => {
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

  const [submitForm] = useMerlinApi(
    // const [simulationResponse, submitForm] = useMerlinApi(
    `/standard_transformer/simulate`,
    { method: "POST" },
    {},
    false
  );

  const simulationResponse = JSON.parse(
    '{"data":{"response":{"instances":{"columns":["rank","driver_id","customer_id","merlin_test_driver_features:test_int32","merlin_test_driver_features:test_float"],"data":[[0,"1234",1111,-1,0],[1,"5678",1111,-1,0]]}},"operation_tracing":{"preprocess":[{"input":null,"output":{"customer_id":1111},"spec":{"name":"customer_id","jsonPathConfig":{"jsonPath":"$.customer.id"}},"operation_type":"variable_op"},{"input":null,"output":{"driver_table":[{"id":"1234","name":"driver-1","row_number":0},{"id":"5678","name":"driver-2","row_number":1}]},"spec":{"name":"driver_table","baseTable":{"fromJson":{"jsonPath":"$.drivers[*]","addRowNumber":true}}},"operation_type":"create_table_op"},{"input":null,"output":{"driver_feature_table":[{"merlin_test_driver_features:test_float":0,"merlin_test_driver_features:test_int32":-1,"merlin_test_driver_id":"1234"},{"merlin_test_driver_features:test_float":0,"merlin_test_driver_features:test_int32":-1,"merlin_test_driver_id":"5678"}]},"spec":{"project":"merlin","entities":[{"name":"merlin_test_driver_id","valueType":"STRING","jsonPathConfig":{"jsonPath":"$.drivers[*].id"}}],"features":[{"name":"merlin_test_driver_features:test_int32","valueType":"INT32","defaultValue":"-1"},{"name":"merlin_test_driver_features:test_float","valueType":"FLOAT","defaultValue":"0"}],"tableName":"driver_feature_table","source":"REDIS"},"operation_type":"feast_op"},{"input":{"driver_table":[{"id":"1234","name":"driver-1","row_number":0},{"id":"5678","name":"driver-2","row_number":1}]},"output":{"driver_table":[{"customer_id":1111,"merlin_test_driver_id":"5678","rank":1},{"customer_id":1111,"merlin_test_driver_id":"1234","rank":0}]},"spec":{"inputTable":"driver_table","outputTable":"driver_table","steps":[{"sort":[{"column":"row_number","order":"DESC"}]},{"renameColumns":{"id":"merlin_test_driver_id","row_number":"rank"}},{"updateColumns":[{"column":"customer_id","expression":"customer_id"}]},{"selectColumns":["customer_id","merlin_test_driver_id","rank"]}]},"operation_type":"table_transform_op"},{"input":{"driver_feature_table":[{"merlin_test_driver_features:test_float":0,"merlin_test_driver_features:test_int32":-1,"merlin_test_driver_id":"1234"},{"merlin_test_driver_features:test_float":0,"merlin_test_driver_features:test_int32":-1,"merlin_test_driver_id":"5678"}],"driver_table":[{"customer_id":1111,"merlin_test_driver_id":"5678","rank":1},{"customer_id":1111,"merlin_test_driver_id":"1234","rank":0}]},"output":{"result_table":[{"customer_id":1111,"merlin_test_driver_features:test_float":0,"merlin_test_driver_features:test_int32":-1,"merlin_test_driver_id":"5678","rank":1},{"customer_id":1111,"merlin_test_driver_features:test_float":0,"merlin_test_driver_features:test_int32":-1,"merlin_test_driver_id":"1234","rank":0}]},"spec":{"leftTable":"driver_table","rightTable":"driver_feature_table","outputTable":"result_table","how":"LEFT","onColumns":["merlin_test_driver_id"]},"operation_type":"table_join_op"},{"input":{"result_table":[{"customer_id":1111,"merlin_test_driver_features:test_float":0,"merlin_test_driver_features:test_int32":-1,"merlin_test_driver_id":"5678","rank":1},{"customer_id":1111,"merlin_test_driver_features:test_float":0,"merlin_test_driver_features:test_int32":-1,"merlin_test_driver_id":"1234","rank":0}]},"output":{"result_table":[{"customer_id":1111,"driver_id":"1234","merlin_test_driver_features:test_float":0,"merlin_test_driver_features:test_int32":-1,"rank":0},{"customer_id":1111,"driver_id":"5678","merlin_test_driver_features:test_float":0,"merlin_test_driver_features:test_int32":-1,"rank":1}]},"spec":{"inputTable":"result_table","outputTable":"result_table","steps":[{"sort":[{"column":"rank"}]},{"renameColumns":{"merlin_test_driver_id":"driver_id"}},{"selectColumns":["rank","driver_id","customer_id","merlin_test_driver_features:test_int32","merlin_test_driver_features:test_float"]}]},"operation_type":"table_transform_op"},{"input":null,"output":{"instances":{"columns":["rank","driver_id","customer_id","merlin_test_driver_features:test_int32","merlin_test_driver_features:test_float"],"data":[[0,"1234",1111,-1,0],[1,"5678",1111,-1,0]]}},"spec":{"jsonTemplate":{"fields":[{"fieldName":"instances","fromTable":{"tableName":"result_table","format":"SPLIT"}}]}},"operation_type":"json_output_op"}],"postprocess":[]}},"isLoading":false,"isLoaded":true,"error":null,"headers":{"content-type":"application/json; charset=UTF-8"}}'
  );

  const {
    data: {
      transformer: { config }
    }
  } = useContext(FormContext);

  const onSubmit = () => {
    let errors = {};

    Object.keys(simulationPayload).forEach(name => {
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
