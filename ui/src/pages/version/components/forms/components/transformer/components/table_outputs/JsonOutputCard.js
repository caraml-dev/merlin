import React, { useCallback, useEffect, useState } from "react";
import {
  EuiDragDropContext,
  euiDragDropReorder,
  EuiDraggable,
  EuiDroppable,
  EuiFlexGroup,
  EuiFlexItem,
  EuiPanel,
  EuiSpacer,
  EuiToolTip,
  EuiText
} from "@elastic/eui";
import { get, useOnChangeHandler } from "@caraml-dev/ui-lib";
import { AddButton } from "../../components/AddButton";
import { JsonOutputFieldCard } from "./JsonOutputFieldCard";
import { BaseJsonOutputCard } from "./BaseJsonOutputCard";
import {
  BaseJson,
} from "../../../../../../../../services/transformer/TransformerConfig";
import { DraggableHeader } from "../../../../components/DraggableHeader";


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

export const JsonOutputCard = ({ 
  jsonOutput, 
  onDelete,
  onChangeHandler, 
  errors = {},
  ...props
}) => {
    const { onChange } = useOnChangeHandler(onChangeHandler);
  
    const [flattenedFields, setFlattenedFields] = useState([]);
    useEffect(
      () => {
        const fields = Array.isArray(
            jsonOutput.jsonTemplate.fields
        ) ? jsonOutput.jsonTemplate.fields
        : []

        const newFlattenedFields = flattenField(fields);
        if (
          JSON.stringify(flattenedFields) !== JSON.stringify(newFlattenedFields)
        ) {
          setFlattenedFields(newFlattenedFields);
        }
      },
      // eslint-disable-next-line react-hooks/exhaustive-deps
      [jsonOutput]
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
    <EuiPanel>
        
        <DraggableHeader
          onDelete={onDelete}
          dragHandleProps={props.dragHandleProps}
        />
        <EuiSpacer size="s" />
        <EuiFlexGroup direction="column" gutterSize="s">
          <EuiFlexItem>
            <EuiToolTip content="JSON Output will create output in JSON format">
              <span>
              <EuiText size="s">
                <h4>JSON Output</h4>
              </EuiText>
                
              </span>
            </EuiToolTip>
          </EuiFlexItem>
        </EuiFlexGroup>
        {
            jsonOutput &&
            jsonOutput.jsonTemplate &&
            jsonOutput.jsonTemplate.baseJson && (
            <EuiFlexGroup direction="column" gutterSize="s">
              <EuiFlexItem>
                <BaseJsonOutputCard
                  baseJson={jsonOutput.jsonTemplate.baseJson}
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
                        onDelete={onDeleteJsonOutputField(fieldIdx)}
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
              { jsonOutput &&
                jsonOutput.jsonTemplate &&
                !jsonOutput.jsonTemplate.baseJson
               && (
                <EuiFlexItem>
                  <AddButton
                    title="+ Add Base JSON"
                    description="Copy the structure and value from another JSON object or array using JSONPath."
                    onClick={() => {
                      var output = {
                        ...jsonOutput,
                        jsonTemplate: {
                          ...jsonOutput.jsonTemplate,
                          baseJson: new BaseJson()
                        }
                      };
                      onAddBaseJson("jsonOutput", output);
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
      </EuiPanel>
    );
};
  