import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
  EuiTextArea,
  EuiText,
  EuiButton,
  EuiToolTip,
  EuiSpacer,
  EuiIcon
} from "@elastic/eui";
import React, { useEffect, useState, useContext } from "react";
import { Panel } from "../components/Panel";
import { useMerlinApi } from "../../../../../hooks/useMerlinApi";
import mocks from "../../../../../mocks";
import { FormContext } from "@gojek/mlp-ui";
import { TransformerSimulationOutputTracing } from "./TransformerSimulationOutputTracing";

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

export const TransformerSimulation = ({ errors = {} }) => {
  const [simulationPayload, setSimulationPayload] = useState(
    new SimulationPayload()
  );
  const [submissionResponse, submitForm] = useMerlinApi(
    `/standard_transformer/simulate`,
    { method: "POST" },
    {},
    false
  );

  const tempResponse = {
    isLoaded: true,
    data: {
      response: {
        instances: {
          columns: [
            "rank",
            "driver_id",
            "customer_id",
            "merlin_test_redis_driver_features:completion_rate",
            "merlin_test_redis_driver_features:cancellation_rate",
            "merlin_test_bt_driver_features:rating"
          ],
          data: [
            [0, "driver_1", 1111, 0, 0, 4.2],
            [1, "driver_2", 1111, 0, 0, 4.2]
          ]
        }
      },
      operation_tracing: {
        preprocess: [
          {
            input: null,
            output: { customer_id: 1111 },
            spec: { name: "customer_id", jsonPath: "$.customer.id" },
            operation_type: "variable_op"
          },
          {
            input: null,
            output: {
              driver_table: [
                { id: "driver_1", name: "driver-1", row_number: 0 },
                { id: "driver_2", name: "driver-2", row_number: 1 }
              ]
            },
            spec: {
              name: "driver_table",
              baseTable: {
                fromJson: { jsonPath: "$.drivers[*]", addRowNumber: true }
              }
            },
            operation_type: "create_table_op"
          },
          {
            input: null,
            output: {
              driver_redis_feature_table: [
                {
                  merlin_test_driver_id: "driver_1",
                  "merlin_test_redis_driver_features:cancellation_rate": 0,
                  "merlin_test_redis_driver_features:completion_rate": 0
                },
                {
                  merlin_test_driver_id: "driver_2",
                  "merlin_test_redis_driver_features:cancellation_rate": 0,
                  "merlin_test_redis_driver_features:completion_rate": 0
                }
              ]
            },
            spec: {
              project: "merlin",
              entities: [
                {
                  name: "merlin_test_driver_id",
                  valueType: "STRING",
                  jsonPath: "$.drivers[*].id"
                }
              ],
              features: [
                {
                  name: "merlin_test_redis_driver_features:completion_rate",
                  valueType: "DOUBLE",
                  defaultValue: "0"
                },
                {
                  name: "merlin_test_redis_driver_features:cancellation_rate",
                  valueType: "DOUBLE",
                  defaultValue: "0"
                }
              ],
              tableName: "driver_redis_feature_table",
              source: "REDIS"
            },
            operation_type: "feast_op"
          },
          {
            input: null,
            output: {
              driver_bigtable_feature_table: [
                {
                  "merlin_test_bt_driver_features:rating": 4.2,
                  merlin_test_driver_id: "driver_1"
                },
                {
                  "merlin_test_bt_driver_features:rating": 4.2,
                  merlin_test_driver_id: "driver_2"
                }
              ]
            },
            spec: {
              project: "merlin",
              entities: [
                {
                  name: "merlin_test_driver_id",
                  valueType: "STRING",
                  jsonPath: "$.drivers[*].id"
                }
              ],
              features: [
                {
                  name: "merlin_test_bt_driver_features:rating",
                  valueType: "DOUBLE",
                  defaultValue: "0"
                }
              ],
              tableName: "driver_bigtable_feature_table",
              source: "BIGTABLE"
            },
            operation_type: "feast_op"
          },
          {
            input: {
              driver_table: [
                { id: "driver_1", name: "driver-1", row_number: 0 },
                { id: "driver_2", name: "driver-2", row_number: 1 }
              ]
            },
            output: {
              driver_table: [
                {
                  customer_id: 1111,
                  merlin_test_driver_id: "driver_2",
                  rank: 1
                },
                {
                  customer_id: 1111,
                  merlin_test_driver_id: "driver_1",
                  rank: 0
                }
              ]
            },
            spec: {
              inputTable: "driver_table",
              outputTable: "driver_table",
              steps: [
                { sort: [{ column: "row_number", order: "DESC" }] },
                {
                  renameColumns: {
                    id: "merlin_test_driver_id",
                    row_number: "rank"
                  }
                },
                {
                  updateColumns: [
                    { column: "customer_id", expression: "customer_id" }
                  ]
                },
                {
                  selectColumns: [
                    "customer_id",
                    "merlin_test_driver_id",
                    "rank"
                  ]
                }
              ]
            },
            operation_type: "table_transform_op"
          },
          {
            input: {
              driver_redis_feature_table: [
                {
                  merlin_test_driver_id: "driver_1",
                  "merlin_test_redis_driver_features:cancellation_rate": 0,
                  "merlin_test_redis_driver_features:completion_rate": 0
                },
                {
                  merlin_test_driver_id: "driver_2",
                  "merlin_test_redis_driver_features:cancellation_rate": 0,
                  "merlin_test_redis_driver_features:completion_rate": 0
                }
              ],
              driver_table: [
                {
                  customer_id: 1111,
                  merlin_test_driver_id: "driver_2",
                  rank: 1
                },
                {
                  customer_id: 1111,
                  merlin_test_driver_id: "driver_1",
                  rank: 0
                }
              ]
            },
            output: {
              result_table: [
                {
                  customer_id: 1111,
                  merlin_test_driver_id: "driver_2",
                  "merlin_test_redis_driver_features:cancellation_rate": 0,
                  "merlin_test_redis_driver_features:completion_rate": 0,
                  rank: 1
                },
                {
                  customer_id: 1111,
                  merlin_test_driver_id: "driver_1",
                  "merlin_test_redis_driver_features:cancellation_rate": 0,
                  "merlin_test_redis_driver_features:completion_rate": 0,
                  rank: 0
                }
              ]
            },
            spec: {
              leftTable: "driver_table",
              rightTable: "driver_redis_feature_table",
              outputTable: "result_table",
              how: "LEFT",
              onColumn: "merlin_test_driver_id"
            },
            operation_type: "table_join_op"
          },
          {
            input: {
              driver_bigtable_feature_table: [
                {
                  "merlin_test_bt_driver_features:rating": 4.2,
                  merlin_test_driver_id: "driver_1"
                },
                {
                  "merlin_test_bt_driver_features:rating": 4.2,
                  merlin_test_driver_id: "driver_2"
                }
              ],
              result_table: [
                {
                  customer_id: 1111,
                  merlin_test_driver_id: "driver_2",
                  "merlin_test_redis_driver_features:cancellation_rate": 0,
                  "merlin_test_redis_driver_features:completion_rate": 0,
                  rank: 1
                },
                {
                  customer_id: 1111,
                  merlin_test_driver_id: "driver_1",
                  "merlin_test_redis_driver_features:cancellation_rate": 0,
                  "merlin_test_redis_driver_features:completion_rate": 0,
                  rank: 0
                }
              ]
            },
            output: {
              result_table: [
                {
                  customer_id: 1111,
                  "merlin_test_bt_driver_features:rating": 4.2,
                  merlin_test_driver_id: "driver_2",
                  "merlin_test_redis_driver_features:cancellation_rate": 0,
                  "merlin_test_redis_driver_features:completion_rate": 0,
                  rank: 1
                },
                {
                  customer_id: 1111,
                  "merlin_test_bt_driver_features:rating": 4.2,
                  merlin_test_driver_id: "driver_1",
                  "merlin_test_redis_driver_features:cancellation_rate": 0,
                  "merlin_test_redis_driver_features:completion_rate": 0,
                  rank: 0
                }
              ]
            },
            spec: {
              leftTable: "result_table",
              rightTable: "driver_bigtable_feature_table",
              outputTable: "result_table",
              how: "LEFT",
              onColumn: "merlin_test_driver_id"
            },
            operation_type: "table_join_op"
          },
          {
            input: {
              result_table: [
                {
                  customer_id: 1111,
                  "merlin_test_bt_driver_features:rating": 4.2,
                  merlin_test_driver_id: "driver_2",
                  "merlin_test_redis_driver_features:cancellation_rate": 0,
                  "merlin_test_redis_driver_features:completion_rate": 0,
                  rank: 1
                },
                {
                  customer_id: 1111,
                  "merlin_test_bt_driver_features:rating": 4.2,
                  merlin_test_driver_id: "driver_1",
                  "merlin_test_redis_driver_features:cancellation_rate": 0,
                  "merlin_test_redis_driver_features:completion_rate": 0,
                  rank: 0
                }
              ]
            },
            output: {
              result_table: [
                {
                  customer_id: 1111,
                  driver_id: "driver_1",
                  "merlin_test_bt_driver_features:rating": 4.2,
                  "merlin_test_redis_driver_features:cancellation_rate": 0,
                  "merlin_test_redis_driver_features:completion_rate": 0,
                  rank: 0
                },
                {
                  customer_id: 1111,
                  driver_id: "driver_2",
                  "merlin_test_bt_driver_features:rating": 4.2,
                  "merlin_test_redis_driver_features:cancellation_rate": 0,
                  "merlin_test_redis_driver_features:completion_rate": 0,
                  rank: 1
                }
              ]
            },
            spec: {
              inputTable: "result_table",
              outputTable: "result_table",
              steps: [
                { sort: [{ column: "rank" }] },
                { renameColumns: { merlin_test_driver_id: "driver_id" } },
                {
                  selectColumns: [
                    "rank",
                    "driver_id",
                    "customer_id",
                    "merlin_test_redis_driver_features:completion_rate",
                    "merlin_test_redis_driver_features:cancellation_rate",
                    "merlin_test_bt_driver_features:rating"
                  ]
                }
              ]
            },
            operation_type: "table_transform_op"
          },
          {
            input: null,
            output: {
              instances: {
                columns: [
                  "rank",
                  "driver_id",
                  "customer_id",
                  "merlin_test_redis_driver_features:completion_rate",
                  "merlin_test_redis_driver_features:cancellation_rate",
                  "merlin_test_bt_driver_features:rating"
                ],
                data: [
                  [0, "driver_1", 1111, 0, 0, 4.2],
                  [1, "driver_2", 1111, 0, 0, 4.2]
                ]
              }
            },
            spec: {
              jsonTemplate: {
                fields: [
                  {
                    fieldName: "instances",
                    fromTable: { tableName: "result_table", format: "SPLIT" }
                  }
                ]
              }
            },
            operation_type: "json_output_op"
          }
        ],
        postprocess: []
      }
    }
  };

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
                        !!simulationPayload.payload
                          ? simulationPayload.payload
                          : ""
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
                        !!simulationPayload.headers
                          ? simulationPayload.headers
                          : ""
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
                      onChange={e =>
                        onChange("mock_response_body", e.target.value)
                      }
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
                      onChange={e =>
                        onChange("mock_response_headers", e.target.value)
                      }
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
          </div>
        </EuiFlexItem>
        {tempResponse.isLoaded && (
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
                      <EuiTextArea
                        value={JSON.stringify(tempResponse.data.response)}
                        readOnly={true}
                      />
                    </EuiFormRow>
                  </EuiFlexItem>
                  <EuiFlexItem>
                    <EuiFormRow
                      fullWidth
                      label="Operation Tracing"
                      display="row">
                      <TransformerSimulationOutputTracing
                        tracingDetails={tempResponse.data.operation_tracing}
                      />
                    </EuiFormRow>
                  </EuiFlexItem>
                </EuiFlexGroup>
              </Panel>
            </div>
          </EuiFlexItem>
        )}
      </EuiFlexGroup>
    </>
  );
};
