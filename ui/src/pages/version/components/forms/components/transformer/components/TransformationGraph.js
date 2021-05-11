import React, { useContext, useEffect, useState } from "react";
import DagreGraph from "dagre-d3-react";
import { FormContext, get } from "@gojek/mlp-ui";
import "./TransformationGraph.scss";

const MODEL_RESPONSE_PREFIX = "$.model_response";

const modelServiceNode = {
  id: "model-node",
  label: "Model Service",
  class: "modelServiceNode"
};

const addFeastInputNodesLinks = (nodes, links, nodeMap, idx, feast, stage) => {
  feast.forEach((feast, feastIdx) => {
    const id = `${stage}-input-${idx}-feast-${feastIdx}`;
    const tableName = feast.tableName;

    nodes.push({
      id: id,
      label: `Feast Table\nOutput: ${tableName}`
    });
    nodeMap[tableName] = id;

    if (stage === "postprocess") {
      feast.entities.forEach(entity => {
        if (
          entity.fieldType === "JSONPath" &&
          entity.field &&
          entity.field.startsWith(MODEL_RESPONSE_PREFIX)
        ) {
          links.push({ source: modelServiceNode.id, target: id });
        }
      });
    }
  });
};

const addGenericTableInputNodesLinks = (
  nodes,
  links,
  nodeMap,
  idx,
  tables,
  stage
) => {
  tables.forEach((table, tableIdx) => {
    const id = `${stage}-input-${idx}-table-${tableIdx}`;
    const tableName = table.name;

    nodes.push({
      id: id,
      label: `Generic Table\nOutput: ${tableName}`
    });
    nodeMap[tableName] = id;

    if (stage === "postprocess") {
      const jsonPath = get(table, "baseTable.fromJson.jsonPath");
      if (
        jsonPath !== undefined &&
        jsonPath.startsWith(MODEL_RESPONSE_PREFIX)
      ) {
        links.push({ source: modelServiceNode.id, target: id });
      }

      get(table, "baseTable.fromTable") &&
        table.columns &&
        table.columns.forEach(column => {
          if (
            column.jsonPath !== undefined &&
            column.jsonPath.startsWith(MODEL_RESPONSE_PREFIX)
          ) {
            links.push({ source: modelServiceNode.id, target: id });
          }
        });
    }
  });
};

const addVariablesInputNodesLinks = (
  nodes,
  links,
  nodeMap,
  idx,
  variables,
  stage
) => {
  const id = `${stage}-variables-input-${idx}`;
  let vars = [];

  variables.forEach(variable => {
    if (variable.name !== undefined && variable.name !== "") {
      vars.push(variable.name);
      nodeMap[variable.name] = id;
    }

    if (
      stage === "postprocess" &&
      variable.jsonPath !== undefined &&
      variable.jsonPath.startsWith(MODEL_RESPONSE_PREFIX)
    ) {
      links.push({ source: modelServiceNode.id, target: id });
    }
  });

  nodes.push({
    id: id,
    label: `Variables\nOutput:\n- ${vars.join("\n- ")}`
  });
};

const addTableTransformationNodesLinks = (
  nodes,
  links,
  nodeMap,
  idx,
  tableTransformation,
  stage
) => {
  const id = `${stage}-table-transformation-${idx}`;
  const outputTable = tableTransformation.outputTable;
  const label = `Table Transformation\nOutput: ${outputTable}`;

  const sourceTable = tableTransformation.inputTable;
  if (nodeMap.hasOwnProperty(sourceTable)) {
    links.push({ source: nodeMap[sourceTable], target: id });
  }

  tableTransformation.steps &&
    tableTransformation.steps.forEach(step => {
      step.updateColumns &&
        step.updateColumns.forEach(updateColumn => {
          if (
            updateColumn.expression !== undefined &&
            updateColumn.expression !== ""
          ) {
            if (nodeMap.hasOwnProperty(updateColumn.expression)) {
              links.push({
                source: nodeMap[updateColumn.expression],
                target: id
              });
            }
          }
        });
    });

  nodes.push({ id: id, label: label });
  nodeMap[outputTable] = id;
};

const addTableJoinNodesLinks = (
  nodes,
  links,
  nodeMap,
  idx,
  tableJoin,
  stage
) => {
  const id = `${stage}-table-join-${idx}`;
  const outputTable = tableJoin.outputTable ? tableJoin.outputTable : "";
  const label = `Table Join\nOutput: ${outputTable}`;

  const leftTable = tableJoin.leftTable;
  if (nodeMap.hasOwnProperty(leftTable)) {
    links.push({ source: nodeMap[leftTable], target: id });
  }

  const rightTable = tableJoin.rightTable;
  if (nodeMap.hasOwnProperty(rightTable)) {
    links.push({ source: nodeMap[rightTable], target: id });
  }

  nodes.push({ id: id, label: label });
  nodeMap[outputTable] = id;
};

const addFieldLinks = (links, nodeMap, id, field, stage) => {
  if (field.fields && field.fields.length > 0) {
    field.fields.forEach(f => {
      addFieldLinks(links, nodeMap, id, f, stage);
    });
  }

  if (
    field.fromTable &&
    field.fromTable.tableName &&
    field.fromTable.tableName !== ""
  ) {
    let tableName = field.fromTable.tableName;
    if (nodeMap.hasOwnProperty(tableName)) {
      links.push({ source: nodeMap[tableName], target: id });
    }
  }

  if (field.expression && field.expression !== "") {
    let tableName = field.expression;
    if (nodeMap.hasOwnProperty(tableName)) {
      links.push({ source: nodeMap[tableName], target: id });
    }
  }

  if (stage === "preprocess") {
    links.push({ source: id, target: modelServiceNode.id });
  } else if (stage === "postprocess") {
    const jsonPath = get(field, "fromJson.jsonPath");
    if (jsonPath !== undefined && jsonPath.startsWith(MODEL_RESPONSE_PREFIX)) {
      links.push({ source: modelServiceNode.id, target: id });
    }
  }
};

const addPipelineNodesLinks = (nodes, links, nodeMap, config, stage) => {
  if (!config[stage]) {
    return;
  }

  config[stage].inputs &&
    config[stage].inputs.forEach((input, idx) => {
      if (input.feast) {
        addFeastInputNodesLinks(nodes, links, nodeMap, idx, input.feast, stage);
      } else if (input.tables) {
        addGenericTableInputNodesLinks(
          nodes,
          links,
          nodeMap,
          idx,
          input.tables,
          stage
        );
      } else if (input.variables) {
        addVariablesInputNodesLinks(
          nodes,
          links,
          nodeMap,
          idx,
          input.variables,
          stage
        );
      }
    });

  config[stage].transformations &&
    config[stage].transformations.forEach((transformation, idx) => {
      if (transformation.tableTransformation) {
        addTableTransformationNodesLinks(
          nodes,
          links,
          nodeMap,
          idx,
          transformation.tableTransformation,
          stage
        );
      } else if (transformation.tableJoin) {
        addTableJoinNodesLinks(
          nodes,
          links,
          nodeMap,
          idx,
          transformation.tableJoin,
          stage
        );
      }
    });

  config[stage].outputs &&
    config[stage].outputs.forEach(output => {
      if (output.jsonOutput && output.jsonOutput.jsonTemplate) {
        let outputNodeExists = false;
        let id = `${stage}-output-base-json`;
        let labels = ["Output"];
        let fields = [];

        const jsonTemplate = output.jsonOutput.jsonTemplate;

        if (jsonTemplate.baseJson) {
          outputNodeExists = true;

          nodes.push({
            id: id,
            label: `Base JSON:\n${jsonTemplate.baseJson.jsonPath}`
          });
          labels.push(`Base JSON: ${jsonTemplate.baseJson.jsonPath}`);

          if (stage === "preprocess") {
            links.push({
              source: id,
              target: modelServiceNode.id
            });
          } else if (stage === "postprocess") {
            if (
              jsonTemplate.baseJson.jsonPath !== undefined &&
              jsonTemplate.baseJson.jsonPath.startsWith(MODEL_RESPONSE_PREFIX)
            ) {
              links.push({ source: modelServiceNode.id, target: id });
            }
          }
        }

        if (jsonTemplate.fields) {
          jsonTemplate.fields.forEach((field, idx) => {
            if (field && field.fieldName) {
              outputNodeExists = true;

              nodes.push({
                id: id,
                label: `Output Field: ${field.fieldName}`
              });

              fields.push(`- ${field.fieldName}`);

              addFieldLinks(links, nodeMap, id, field, stage);
            }
          });
        }

        if (outputNodeExists) {
          if (fields.length > 0) {
            labels.push("Fields:");
            labels = labels.concat(fields);
          }

          // Overwrite with new label
          nodes.push({
            id: id,
            label: labels.join("\n")
          });
        }
      }
    });
};

export const TransformationGraph = () => {
  const { data } = useContext(FormContext);

  const [graphData, setGraphData] = useState({ nodes: [], links: [] });

  useEffect(() => {
    if (
      data.transformer &&
      data.transformer.config &&
      data.transformer.config.transformerConfig
    ) {
      const config = data.transformer.config.transformerConfig;

      let nodes = [],
        links = [];
      let nodeMap = {};

      addPipelineNodesLinks(nodes, links, nodeMap, config, "preprocess");
      nodes.push(modelServiceNode);
      addPipelineNodesLinks(nodes, links, nodeMap, config, "postprocess");

      setGraphData({ nodes, links });
    }
  }, [data, setGraphData]);

  return (
    <DagreGraph
      nodes={graphData.nodes}
      links={graphData.links}
      config={{
        rankdir: "TB"
      }}
      width="100%"
      height="640px"
      className="transformationGraph"
      zoomable
    />
  );
};
