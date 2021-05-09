import React, { useContext, useEffect, useState } from "react";
import DagreGraph from "dagre-d3-react";
import { EuiSpacer } from "@elastic/eui";
import { FormContext } from "@gojek/mlp-ui";
import { Panel } from "../../Panel";
import "./GraphPanel.scss";

const modelServiceNode = {
  id: "model",
  label: "Model Service",
  class: "modelServiceNode"
};

const placeholderNodes = [
  { id: "preprocess", label: "Preprocess" },
  modelServiceNode,
  { id: "postprocess", label: "Postprocess" }
];

const placeholderLinks = [
  { source: "preprocess", target: modelServiceNode.id },
  { source: modelServiceNode.id, target: "postprocess" }
];

export const GraphPanel = () => {
  const { data } = useContext(FormContext);

  const [graphData, setGraphData] = useState({ nodes: [], links: [] });

  useEffect(() => {
    if (data.transformer && data.transformer.config) {
      const config = data.transformer.config;
      console.log("config", JSON.stringify(config, null, 2));

      let nodes = [],
        links = [];
      let nodeMap = {};

      // Preprocess
      config.preprocess.inputs.forEach((input, idx) => {
        input.feast &&
          input.feast.forEach((feast, feastIdx) => {
            const id = `feast-input-${idx}-${feastIdx}`;
            const tableName = feast.tableName;

            nodes.push({
              id: id,
              label: `#${idx + 1} - Feast Table
${tableName}`
            });
            nodeMap[tableName] = id;
          });

        input.tables &&
          input.tables.forEach((table, tableIdx) => {
            const id = `table-input-${idx}-${tableIdx}`;
            nodes.push({ id: id, label: `#${idx + 1} - Generic Table` });

            const tableName = table.name;
            nodeMap[tableName] = id;
          });
      });

      config.preprocess.transformations.forEach((transformation, idx) => {
        const id = `transformation-${idx}`;
        let label = "";
        let outputTable = "";

        if (transformation.tableTransformation) {
          label = `${idx + 1} - Table Transformation`;
          outputTable = transformation.tableTransformation.outputTable;

          const sourceTable = transformation.tableTransformation.inputTable;
          if (nodeMap.hasOwnProperty(sourceTable)) {
            links.push({ source: nodeMap[sourceTable], target: id });
          }
        } else if (transformation.tableJoin) {
          label = `${idx + 1} - Table Join`;
          outputTable = transformation.tableJoin.outputTable;

          const leftTable = transformation.tableJoin.leftTable;
          if (nodeMap.hasOwnProperty(leftTable)) {
            links.push({ source: nodeMap[leftTable], target: id });
          }

          const rightTable = transformation.tableJoin.rightTable;
          if (nodeMap.hasOwnProperty(rightTable)) {
            links.push({ source: nodeMap[rightTable], target: id });
          }
        }

        nodes.push({ id: id, label: label });
        nodeMap[outputTable] = id;
      });

      // Output
      // nodes.push({ id: "preprocess-output", label: "Preprocess Output" })
      // links.push({ source: "preprocess-output", target: "model"})

      // Model service node
      nodes.push(modelServiceNode);

      // if (nodes.length === 0 && links.length === 0) {
      //   nodes = placeholderNodes;
      //   links = placeholderLinks;
      // }

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
