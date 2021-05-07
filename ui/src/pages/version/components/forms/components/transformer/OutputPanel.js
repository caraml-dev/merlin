import React, { useCallback, useEffect, useState } from "react";
import {
  EuiDragDropContext,
  euiDragDropReorder,
  EuiDraggable,
  EuiDroppable,
  EuiFlexGroup,
  EuiFlexItem,
  EuiSpacer,
  EuiText
} from "@elastic/eui";
import { useOnChangeHandler } from "@gojek/mlp-ui";
import { Panel } from "../Panel";
import { AddButton } from "./components/AddButton";
import { JsonOutputFieldCard } from "./components/JsonOutputFieldCard";
import { BaseJsonOutputCard } from "./components/BaseJsonOutputCard";
import {
  BaseJson,
  Field,
  FieldFromExpression,
  FieldFromJson,
  FromJson,
  FieldFromTable,
  FromTable,
  JsonOutput,
  Output
} from "../../../../../../services/transformer/TransformerConfig";

export const OutputPanel = ({
  outputs,
  onChangeHandler,
  errors = {} // TODO
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const [baseJson, setBaseJson] = useState(null);

  var fields = [];
  var firstField = new Field();
  firstField.fieldName = "order_id";
  var value = new FieldFromJson();
  value.fromJson.jsonPath = "$.order_id";
  firstField.value = value;
  var secondField = new Field();
  secondField.fieldName = "customer_ids";
  var secondValue = new FieldFromTable();
  secondValue.fromTable.tableName = "table_customer";
  secondValue.fromTable.format = "RECORD";
  secondField.value = secondValue;
  var nestedField = new Field();
  nestedField.fieldName = "nested";
  var nestedFields = [];
  var firstNestedField = new Field();
  firstNestedField.fieldName = "expr";
  var firstNestedFieldVal = new FieldFromExpression();
  firstNestedFieldVal.expression = "table.Row(1)";
  firstNestedField.value = firstNestedFieldVal;
  nestedFields.push(firstNestedFieldVal);
  nestedField.fields = nestedFields;

  fields.push(firstField, secondField, nestedField);
  var output = new Output();
  output.jsonOutput.jsonTemplate.fields = fields;
  console.log("jsonOutput ", JSON.stringify(output.jsonOutput));

  var bJson = new BaseJson();
  bJson.jsonPath = "$.field_1";
  output.jsonOutput.jsonTemplate.baseJson = bJson;
  outputs[0] = output;

  const onAddBaseJson = useCallback((field, input) => {
    // var outJson = output
    // console.log("outJson " + outJson)
    // if (output == undefined) {
    //   outJson = new Output()
    //   console.log("assign " + JSON.stringify(outJson))
    // }
    // var bJson =
    // bJson.jsonPath = input
    // templateJson.baseJson = bJson
    // setBaseJson(bJson)

    // outJson.templateJson = templateJson
    // setOutput(outJson)
    // if (outputs.size == undefined) {
    //   onChangeHandler([new Output()])
    // }

    // setBaseJson(input)
    onChangeHandler([...outputs, { [field]: input }]);
  });

  const onAddOutput = useCallback((field, input) => {
    onChangeHandler([...outputs, { [field]: input }]);
  });

  const onDeleteOutput = idx => () => {
    outputs.splice(idx, 1);
    onChangeHandler([...outputs]);
  };

  const onDragEnd = ({ source, destination }) => {
    if (source && destination) {
      const items = euiDragDropReorder(
        outputs,
        source.index,
        destination.index
      );
      onChangeHandler([...items]);
    }
  };

  const buildJsonFieldConfigurationCard = (
    parentPath,
    index,
    field,
    fieldName,
    result,
    provided
  ) => {
    if (field.fields.size === 0) {
      console.log("fallback");
      result.push(
        <JsonOutputFieldCard
          index={index}
          field={field}
          onChangeHandler={onChange(parentPath)}
          dragHandleProps={provided.dragHandleProps}
        />
      );
      return result;
    }
    for (var i = 0; i < field.fields.size; i++) {
      console.log("nested");
      const fieldInIdx = field.fields[i];
      result = buildJsonFieldConfigurationCard(
        `${parentPath}.fields.${i}`,
        fieldInIdx,
        `${fieldName}.${fieldInIdx.fieldName}`,
        result,
        provided
      );
    }
    return result;
  };

  const idx = 0;

  return (
    <Panel title="Output" contentWidth="75%">
      <EuiDragDropContext onDragEnd={onDragEnd}>
        <EuiFlexGroup direction="column" gutterSize="s">
          <EuiDroppable droppableId="OUTPUTS_DROPPABLE_AREA" spacing="m">
            {outputs.map((output, idx) => (
              <EuiDraggable
                key="baseJson"
                draggableId="baseJson"
                customDragHandle={true}
                disableInteractiveElementBlocking>
                {provided => (
                  <EuiFlexItem key={`baseJson-${idx}`}>
                    {output.jsonOutput.jsonTemplate.baseJson && (
                      <BaseJsonOutputCard
                        baseJson={output.jsonOutput.jsonTemplate.baseJson}
                        onChangeHandler={onChange(
                          `${idx}.jsonOutput.jsonTemplate.baseJson`
                        )}
                        dragHandleProps={provided.dragHandleProps}
                      />
                    )}
                  </EuiFlexItem>
                )}
              </EuiDraggable>
            ))}
          </EuiDroppable>

          <EuiSpacer size="s" />
          <EuiDroppable>
            {outputs.map((output, idx) => (
              <EuiDraggable
                key={`${idx}`}
                index={idx}
                draggableId={`${idx}`}
                customDragHandle={true}
                disableInteractiveElementBlocking>
                {provided =>
                  output.jsonOutput.jsonTemplate.fields.map(
                    (field, fieldIdx) => (
                      <JsonOutputFieldCard
                        index={fieldIdx}
                        field={field}
                        onChangeHandler={onChange(
                          `0.jsonOutput.jsonTemplate.fields.${fieldIdx}`
                        )}
                        dragHandleProps={provided.dragHandleProps}
                      />
                    )
                  )
                }
              </EuiDraggable>
            ))}
          </EuiDroppable>

          <EuiFlexGroup direction="row" gutterSize="s">
            <EuiFlexItem>
              <AddButton
                title="+ Add Base JSON"
                // TODO:
                // description="Use Feast features as input"
                onClick={() => {
                  var jsonOutput = new JsonOutput();
                  if (outputs.size > 0) {
                    jsonOutput = outputs[0];
                  }
                  jsonOutput.jsonTemplate.baseJson = new BaseJson();
                  onAddBaseJson("jsonOutput", jsonOutput);
                }}
              />
            </EuiFlexItem>
            <EuiFlexItem>
              <AddButton
                title="+ Add Field"
                // TODO:
                // description="Use Feast features as input"
                onClick={() => {
                  // var jsonOutput = new JsonOutput()
                  // if (outputs.size > 0) {
                  //   jsonOutput = outputs[0].jsonOutput
                  //   console.log("jsonoutput there " + JSON.stringify(jsonOutput))
                  // }
                  // var fields = jsonOutput.jsonTemplate.fields
                  // if (fields.size === 0) {
                  //   fields = [new Field()]
                  // } else {
                  //   fields = [...fields, new Field()]
                  // }
                  // jsonOutput.jsonTemplate.fields = fields
                  // console.log("outputs " + JSON.stringify(outputs))
                  // outputs[0].jsonOutput = jsonOutput
                  // console.log("outputs after " + JSON.stringify(outputs))
                  // onChangeHandler([...outputs])
                }}
              />
            </EuiFlexItem>
          </EuiFlexGroup>
        </EuiFlexGroup>
      </EuiDragDropContext>
    </Panel>
  );
};
