import React, { useCallback, useEffect, useState } from "react";
import {
  EuiDragDropContext,
  euiDragDropReorder,
  EuiDraggable,
  EuiDroppable,
  EuiFlexGroup,
  EuiFlexItem,
  EuiSpacer
} from "@elastic/eui";
import { get, useOnChangeHandler } from "@gojek/mlp-ui";
import { Panel } from "../Panel";
import { AddButton } from "./components/AddButton";
import { JsonOutputFieldCard } from "./components/table_outpus/JsonOutputFieldCard";
import { BaseJsonOutputCard } from "./components/table_outpus/BaseJsonOutputCard";
import {
  BaseJson,
  JsonOutput
} from "../../../../../../services/transformer/TransformerConfig";

const expandFields = flattenField => {
  let fields = [];

  flattenField.forEach(f => {
    if (f.fieldName === undefined) {
      return;
    }

    const nameSegments = splitName(f.fieldName);
    let newField = {
      ...f,
      fieldName: nameSegments[nameSegments.length - 1]
    };

    fields = generateOutputFields(fields, nameSegments, newField);
  });

  return fields;
};

const splitName = name => {
  // https://stackoverflow.com/questions/171480/regex-grabbing-values-between-quotation-marks
  return (
    name
      // eslint-disable-next-line no-useless-escape
      .split(/\.(?=(?:[^\"]*\"[^\"]*\")*[^\"]*$)/g)
      // eslint-disable-next-line no-useless-escape
      .map(n => n.replace(/\"/g, ""))
  );
};

const generateOutputFields = (fields, nameSegment, fieldValue) => {
  if (nameSegment.length <= 1) {
    fields.push({ ...fieldValue, fields: [] });
    return fields;
  }

  let field = searchField(fields, nameSegment[0]);
  if (field === undefined) {
    field = {
      fieldName: nameSegment[0],
      fields: []
    };
    fields.push(field);
  }

  generateOutputFields(field.fields, nameSegment.slice(1), fieldValue);
  return fields;
};

const searchField = (fields, fieldName) => {
  return fields && fields.find(f => f.fieldName === fieldName);
};

const flattenField = fields => {
  let all = [];
  fields.forEach(f => {
    const flattenedFields = flatten(f, [f.fieldName]);
    all = all.concat(flattenedFields);
  });
  return all;
};

const flatten = (field, path) => {
  if (field.fields === undefined || field.fields.length === 0) {
    return [
      {
        ...field,
        fieldName: mergePath(path)
      }
    ];
  }

  let fields = [];
  field.fields.forEach(f => {
    path.push(f.fieldName);
    fields = fields.concat(flatten(f, path));
    path.pop();
  });
  return fields;
};

const mergePath = path => {
  if (path.length === 1) {
    return path[0];
  }

  return '"' + path.join('"."') + '"';
};

export const OutputPanel = ({ outputs = [], onChangeHandler, errors = {} }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const [flattenedFields, setFlattenedFields] = useState([]);
  useEffect(
    () => {
      const newFlattenedFields = flattenField(
        outputs.length > 0
          ? Array.isArray(outputs[0].jsonOutput.jsonTemplate.fields)
            ? outputs[0].jsonOutput.jsonTemplate.fields
            : []
          : []
      );
      if (
        JSON.stringify(flattenedFields) !== JSON.stringify(newFlattenedFields)
      ) {
        setFlattenedFields(newFlattenedFields);
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [outputs]
  );

  useEffect(
    () => {
      if (flattenedFields.length > 0) {
        let fields = expandFields(flattenedFields);
        onChange(`0.jsonOutput.jsonTemplate.fields`)(fields);
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [flattenedFields]
  );

  const onFieldChange = (idx, fieldObj) => {
    flattenedFields[idx] = fieldObj;
    setFlattenedFields([...flattenedFields]);
  };

  const onAddBaseJson = useCallback(
    (field, input) => {
      onChangeHandler([{ [field]: input }]);
    },
    [onChangeHandler]
  );

  const onDeleteJsonOutputField = idx => () => {
    flattenedFields.splice(idx, 1);
    setFlattenedFields([...flattenedFields]);
  };

  const onDragEnd = ({ source, destination }) => {
    if (source && destination) {
      const items = euiDragDropReorder(
        flattenedFields,
        source.index,
        destination.index
      );
      setFlattenedFields(items);
    }
  };

  return (
    <Panel title="Output" contentWidth="80%">
      <EuiSpacer size="xs" />

      {outputs.findIndex(
        output =>
          output.jsonOutput &&
          output.jsonOutput.jsonTemplate &&
          output.jsonOutput.jsonTemplate.baseJson
      ) !== -1 && (
        <EuiFlexGroup direction="column" gutterSize="s">
          <EuiFlexItem>
            <BaseJsonOutputCard
              baseJson={outputs[0].jsonOutput.jsonTemplate.baseJson}
              onChangeHandler={onChange(`0.jsonOutput.jsonTemplate.baseJson`)}
              errors={get(errors, `0.jsonOutput.jsonTemplate.baseJson`)}
              onDelete={() => {
                onChange(`0.jsonOutput.jsonTemplate.baseJson`)(undefined);
              }}
            />
            <EuiSpacer size="s" />
          </EuiFlexItem>
        </EuiFlexGroup>
      )}

      <EuiDragDropContext onDragEnd={onDragEnd}>
        <EuiFlexGroup direction="column" gutterSize="s">
          <EuiDroppable droppableId="OUTPUTS_DROPPABLE_AREA" spacing="m">
            {flattenedFields.map((field, fieldIdx) => (
              <EuiDraggable
                key={`fields-${fieldIdx}`}
                index={fieldIdx}
                draggableId={`fields-${fieldIdx}`}
                customDragHandle={true}
                disableInteractiveElementBlocking>
                {provided => (
                  <EuiFlexItem key={`fields-${fieldIdx}`}>
                    <JsonOutputFieldCard
                      index={fieldIdx}
                      field={field}
                      onChange={onFieldChange}
                      onDelete={
                        flattenedFields.length > 1
                          ? onDeleteJsonOutputField(fieldIdx)
                          : undefined
                      }
                      dragHandleProps={provided.dragHandleProps}
                    />
                    <EuiSpacer size="s" />
                  </EuiFlexItem>
                )}
              </EuiDraggable>
            ))}
          </EuiDroppable>

          <EuiSpacer size="s" />

          <EuiFlexGroup direction="row" gutterSize="s">
            {outputs.findIndex(
              output =>
                output.jsonOutput &&
                output.jsonOutput.jsonTemplate &&
                output.jsonOutput.jsonTemplate.baseJson
            ) === -1 && (
              <EuiFlexItem>
                <AddButton
                  title="+ Add Base JSON"
                  description="Copy the structure and value from another JSON object or array using JSONPath."
                  onClick={() => {
                    var jsonOutput = new JsonOutput();
                    if (outputs.length > 0) {
                      jsonOutput = outputs[0].jsonOutput;
                    }
                    jsonOutput = {
                      ...jsonOutput,
                      jsonTemplate: {
                        ...jsonOutput.jsonTemplate,
                        baseJson: new BaseJson()
                      }
                    };
                    onAddBaseJson("jsonOutput", jsonOutput);
                  }}
                />
              </EuiFlexItem>
            )}

            <EuiFlexItem>
              <AddButton
                title="+ Add Field"
                description="Create a field from JSONPath, table, or expression."
                onClick={() => {
                  setFlattenedFields([...flattenedFields, {}]);
                }}
              />
            </EuiFlexItem>
          </EuiFlexGroup>
        </EuiFlexGroup>
      </EuiDragDropContext>
    </Panel>
  );
};
