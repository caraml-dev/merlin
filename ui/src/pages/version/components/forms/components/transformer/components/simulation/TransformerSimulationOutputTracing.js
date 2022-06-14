import {
  EuiModal,
  EuiModalBody,
  EuiModalHeader,
  EuiModalHeaderTitle
} from "@elastic/eui";
import dagre from "dagre";
import React, { useCallback, useState } from "react";
import ReactFlow, {
  addEdge,
  Background,
  Controls,
  useEdgesState,
  useNodesState
} from "react-flow-renderer";
import PipelineNode from "./PipelineNode";
import { PipelineNodeDetails } from "./PipelineNodeDetails";

const dagreGraph = new dagre.graphlib.Graph();
dagreGraph.setDefaultEdgeLabel(() => ({}));

const nodeTypes = { pipeline: PipelineNode };

const createNodes = details => {
  let nodes = [];

  if (details.preprocess) {
    for (let i = 0; i < details.preprocess.length; i++) {
      nodes.push({
        id: "preprocess-" + i,
        type: "pipeline",
        position: { x: 0, y: 0 },
        data: {
          label: details.preprocess[i].operation_type,
          ...details.preprocess[i]
        }
      });
    }
  }

  nodes.push({
    id: "prediction",
    position: { x: 0, y: 0 },
    data: { label: "Model Prediction" }
  });

  if (details.postprocess) {
    for (let i = 0; i < details.postprocess.length; i++) {
      nodes.push({
        id: "postprocess-" + i,
        type: "pipeline",
        position: { x: 0, y: 0 },
        data: {
          label: details.postprocess[i].operation_type,
          ...details.postprocess[i]
        }
      });
    }
  }

  return nodes;
};

const createChainEdges = nodes => {
  let edges = [];

  for (let i = 0; i < nodes.length - 1; i++) {
    let sourceNode = nodes[i];
    let targetNode = nodes[i + 1];

    edges.push({
      id: sourceNode.id + "-" + targetNode.id,
      source: sourceNode.id,
      target: targetNode.id
    });
  }

  return edges;
};

const getLayoutedElements = (nodes, edges, direction = "TB") => {
  const isHorizontal = direction === "LR";
  dagreGraph.setGraph({ rankdir: direction });

  nodes.forEach(node => {
    dagreGraph.setNode(node.id, {
      width: 120,
      height: 10
    });
  });

  edges.forEach(edge => {
    dagreGraph.setEdge(edge.source, edge.target);
  });

  dagre.layout(dagreGraph);

  nodes.forEach(node => {
    const nodeWithPosition = dagreGraph.node(node.id);
    node.targetPosition = isHorizontal ? "left" : "top";
    node.sourcePosition = isHorizontal ? "right" : "bottom";

    // We are shifting the dagre node position (anchor=center center) to the top left
    // so it matches the React Flow node anchor point (top left).
    node.position = {
      x: nodeWithPosition.x,
      y: nodeWithPosition.y
    };

    return node;
  });

  edges.forEach(edge => {
    edge.targetHandle = isHorizontal ? "target-left" : "target-top";
    edge.sourceHandle = isHorizontal ? "source-right" : "source-bottom";
    return edge;
  });

  return { nodes, edges };
};

export const TransformerSimulationOutputTracing = ({ tracingDetails }) => {
  const initialNodes = createNodes(tracingDetails);
  const initialEdges = createChainEdges(initialNodes);

  const { nodes: layoutedNodes, edges: layoutedEdges } = getLayoutedElements(
    initialNodes,
    initialEdges,
    "LR"
  );

  const [nodes, , onNodesChange] = useNodesState(layoutedNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(layoutedEdges);

  const onConnect = useCallback(
    params => setEdges(eds => addEdge({ ...params }, eds)),
    [setEdges]
  );

  const [higlightedNode, setHightlightedNode] = useState(undefined);

  const onNodeClicked = (_, node) => {
    if (node.type === "pipeline") {
      setHightlightedNode(node);
    } else {
      setHightlightedNode(undefined);
    }
  };

  const onCloseModal = () => {
    setHightlightedNode(undefined);
  };

  return (
    <div style={{ height: 640 }}>
      <ReactFlow
        nodes={nodes}
        nodeTypes={nodeTypes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        onNodeClick={onNodeClicked}
        fitView>
        <Controls />
        <Background color="#aaa" gap={16} />
      </ReactFlow>

      {!!higlightedNode && (
        <EuiModal onClose={onCloseModal} style={{ width: "800px" }}>
          <EuiModalHeader>
            <EuiModalHeaderTitle>
              <h2>{higlightedNode.data.operation_type}</h2>
            </EuiModalHeaderTitle>
          </EuiModalHeader>

          <EuiModalBody>
            <PipelineNodeDetails details={higlightedNode.data} />
          </EuiModalBody>
        </EuiModal>
      )}
    </div>
  );
};
