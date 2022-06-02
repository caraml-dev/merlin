import React from "react";
import ReactFlow, {
  addEdge,
  MiniMap,
  Controls,
  Background,
  useNodesState,
  useEdgesState
} from "react-flow-renderer";

import { MarkerType } from "react-flow-renderer";

const createChainEdges = nodes => {
  var edges = [];
  for (let i = 0; i < nodes.length - 1; i++) {
    var sourceNode = nodes[i];
    var targetNode = nodes[i + 1];
    edges.push({
      id: sourceNode.id + "-" + targetNode.id,
      source: sourceNode.id,
      target: targetNode.id
    });
  }
};

// const nodeTypes = { pipeline: PipelineNode };

const onInit = reactFlowInstance =>
  console.log("flow loaded:", reactFlowInstance);

export const TransformerSimulationOutputTracing = tracingDetails => {
  const createNodes = tracingDetails => {
    const nodeHeight = 100;
    const nodeWidth = 70;
    const marginHorizontal = 20;
    const marginVertical = 30;
    var nodes = [];
    const initialXPosition = 0;
    const initialYPosition = 20;
    let lastXPosition = initialXPosition;
    let lastYPosition = initialYPosition;
    for (let i = 0; i < tracingDetails.preprocess.length; i++) {
      if (i != 0) {
        lastXPosition = lastXPosition + marginHorizontal + nodeWidth;
      }
      nodes.push({
        id: "preprocess-" + i,
        type: "pipeline",
        position: { x: lastXPosition, y: lastYPosition },
        data: { value: tracingDetails.preprocess[i] }
      });
    }

    lastYPosition = lastYPosition + marginVertical + nodeHeight;
    nodes.push({
      id: "prediction",
      type: "prediction",
      position: { x: lastXPosition, y: lastYPosition },
      data: { label: "Prediction" }
    });

    var directionToRight = false;
    for (let i = 0; i < tracingDetails.postprocess.length; i++) {
      let tempXPosition = lastXPosition - (marginHorizontal + nodeWidth);
      if (tempXPosition < 0) {
        directionToRight = true;
        lastYPosition = lastYPosition + marginVertical + nodeHeight;
      }
      lastXPosition = directionToRight
        ? lastXPosition + marginHorizontal + nodeWidth
        : tempXPosition;

      nodes.push({
        id: "preprocess-" + i,
        type: "pipeline",
        position: { x: lastXPosition, y: lastYPosition },
        data: { value: tracingDetails.postprocess[i] }
      });
    }
  };
  const initialNodes = createNodes();
  const initialEdges = createChainEdges(initialEdges);
  const [nodes, setNodes, onNodesChange] = useNodesState(createNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const onConnect = params => setEdges(eds => addEdge(params, eds));

  return (
    <div style={{ height: 800 }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        onInit={onInit}
        fitView
        attributionPosition="top-right">
        <MiniMap
          nodeStrokeColor={n => {
            if (n.style?.background) return n.style.background;
            if (n.type === "input") return "#0041d0";
            if (n.type === "output") return "#ff0072";
            if (n.type === "default") return "#1a192b";

            return "#eee";
          }}
          nodeColor={n => {
            if (n.style?.background) return n.style.background;

            return "#fff";
          }}
          nodeBorderRadius={2}
        />
        <Controls />
        <Background color="#aaa" gap={16} />
      </ReactFlow>
    </div>
  );
};
