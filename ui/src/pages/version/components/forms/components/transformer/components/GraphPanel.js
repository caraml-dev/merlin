import React, { useContext, useEffect, useState } from "react";
import DagreGraph from "dagre-d3-react";
import { EuiSpacer } from "@elastic/eui";
import { FormContext } from "@gojek/mlp-ui";
import { Panel } from "../../Panel";
import "./GraphPanel.scss";

const modelNodeId = "model";

const modelServiceNode = {
  id: modelNodeId,
  label: "Model Service",
  class: "modelServiceNode"
};

const addFeastInputNodes = (nodes, nodeMap, idx, feast, stage) => {
  feast.forEach(feast => {
    const id = `feast-input-${idx}`;
    const tableName = feast.tableName;
    nodes.push({
      id: id,
      label: `#${idx + 1} - Feast Table\nOutput: ${tableName}`
    });
    nodeMap[tableName] = id;
  });
};

const addGenericTableInputNodes = (nodes, nodeMap, idx, tables, stage) => {
  tables.forEach(table => {
    const id = `table-input-${idx}`;
    const tableName = table.name;
    nodes.push({
      id: id,
      label: `#${idx + 1} - Generic Table\nOutput: ${tableName}`
    });
    nodeMap[tableName] = id;
  });
};

const addVariablesInputNodes = (nodes, nodeMap, idx, variables, stage) => {
  const id = `variables-input-${idx}`;
  let vars = [];
  variables.forEach(variable => {
    if (variable.name !== undefined && variable.name !== "") {
      vars.push(variable.name);
      nodeMap[variable.name] = id;
    }
  });
  nodes.push({
    id: id,
    label: `#${idx + 1} - Variables\nOutput: ${vars.join(", ")}`
  });
};

const addTableTransformationNodesLinks = (
  nodes,
  nodeMap,
  links,
  idx,
  tableTransformation,
  stage
) => {
  const id = `table-transformation-${idx}`;
  const outputTable = tableTransformation.outputTable;
  const label = `#${idx + 1} - Table Transformation\nOutput: ${outputTable}`;

  const sourceTable = tableTransformation.inputTable;
  if (nodeMap.hasOwnProperty(sourceTable)) {
    links.push({ source: nodeMap[sourceTable], target: id });
  }

  // Link to variables
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
  nodeMap,
  links,
  idx,
  tableJoin,
  stage
) => {
  const id = `table-join-${idx}`;
  const outputTable = tableJoin.outputTable ? tableJoin.outputTable : "";
  const label = `#${idx + 1} - Table Join\nOutput: ${outputTable}`;

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

const addFieldLinks = (nodeMap, links, id, field, stage) => {
  if (field.fields && field.fields.length > 0) {
    field.fields.forEach(f => addFieldLinks(id, f));
  }

  if (field.fromTable) {
    if (field.fromTable.tableName && field.fromTable.tableName !== "") {
      let tableName = field.fromTable.tableName;
      if (nodeMap.hasOwnProperty(tableName)) {
        links.push({ source: nodeMap[tableName], target: id });
      }
    }
  }

  if (field.expression && field.expression !== "") {
    let tableName = field.expression;
    if (nodeMap.hasOwnProperty(tableName)) {
      links.push({ source: nodeMap[tableName], target: id });
    }
  }

  links.push({ source: id, target: modelNodeId });
};

const addPipelineNodesLinks = (nodes, nodeMap, links, config, stage) => {
  config[stage].inputs.forEach((input, idx) => {
    if (input.feast) {
      addFeastInputNodes(nodes, nodeMap, idx, input.feast);
    } else if (input.tables) {
      addGenericTableInputNodes(nodes, nodeMap, idx, input.tables);
    } else if (input.variables) {
      addVariablesInputNodes(nodes, nodeMap, idx, input.variables);
    }
  });

  config[stage].transformations.forEach((transformation, idx) => {
    if (transformation.tableTransformation) {
      addTableTransformationNodesLinks(
        nodes,
        nodeMap,
        links,
        idx,
        transformation.tableTransformation
      );
    } else if (transformation.tableJoin) {
      addTableJoinNodesLinks(
        nodes,
        nodeMap,
        links,
        idx,
        transformation.tableJoin
      );
    }
  });

  config[stage].outputs.forEach((output, idx) => {
    if (output.jsonOutput && output.jsonOutput.jsonTemplate) {
      const jsonTemplate = output.jsonOutput.jsonTemplate;
      if (jsonTemplate.baseJson) {
        nodes.push({
          id: "preprocess-output-base-json",
          label: `Output Base JSON: ${jsonTemplate.baseJson.jsonPath}`
        });
        links.push({
          source: "preprocess-output-base-json",
          target: modelNodeId
        });
      }

      if (jsonTemplate.fields) {
        jsonTemplate.fields.forEach((field, idx) => {
          const id = `preprocess-output-field-${idx}`;
          nodes.push({
            id: id,
            label: `Output Field: ${field.fieldName}`
          });
          addFieldLinks(nodeMap, links, id, field);
        });
      }
    }
  });
};

export const GraphPanel = () => {
  const { data } = useContext(FormContext);

  const [graphData, setGraphData] = useState({ nodes: [], links: [] });

  useEffect(() => {
    if (data.transformer && data.transformer.config) {
      const config = data.transformer.config;

      let nodes = [],
        links = [];
      let nodeMap = {};

      // Preprocess
      addPipelineNodesLinks(nodes, nodeMap, links, config, "preprocess");

      // Model service node
      nodes.push(modelServiceNode);

      setGraphData({ nodes, links });
    }
  }, [data, setGraphData]);

  return (
    <Panel title="Transformation Graph" contentWidth="100%">
      <EuiSpacer size="xs" />
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
    </Panel>
  );
};
