import {
  EuiModal,
  EuiModalBody,
  EuiModalHeader,
  EuiModalHeaderTitle
} from "@elastic/eui";
import dagre from "dagre";
import React, { useState } from "react";
import ReactFlow, {
  addEdge,
  Background,
  Controls,
  MarkerType,
  useEdgesState,
  useNodesState
} from "react-flow-renderer";
import PipelineNode from "./PipelineNode";
import { PipelineNodeDetails } from "./PipelineNodeDetails";

const nodeTypes = { pipeline: PipelineNode };

export const TransformerSimulationOutputTracing = ({ tracingDetails }) => {
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));

  const leftDirection = "left";
  const rightDirection = "right";
  const topDirection = "top";
  const bottomDirection = "bottom";

  const reverseDirection = position => {
    if (position === leftDirection) {
      return rightDirection;
    }
    if (position === rightDirection) {
      return leftDirection;
    }
    if (position === topDirection) {
      return bottomDirection;
    }
    if (position === bottomDirection) {
      return topDirection;
    }
  };

  const createChainEdges = nodes => {
    var edges = [];
    for (let i = 0; i < nodes.length - 1; i++) {
      var sourceNode = nodes[i];
      var targetNode = nodes[i + 1];
      let sourcePosition = targetNode.sourcePosition;
      if (sourcePosition === undefined) {
        sourcePosition = rightDirection;
      }
      edges.push({
        id: sourceNode.id + "-" + targetNode.id,
        source: sourceNode.id,
        target: targetNode.id,
        sourceHandle: "source-" + sourcePosition,
        targetHandle: "target-" + reverseDirection(sourcePosition),
        markerEnd: {
          type: MarkerType.ArrowClosed
        }
      });
    }

    return edges;
  };

  // this function is to create position like reverted 's'
  const calculatePosition = (
    positionX,
    positionY,
    currentDirection,
    directionInfo
  ) => {
    const selectedDirectionInfo = directionInfo[currentDirection];
    let newPositionY = positionY;
    let newPositionX = positionX;
    let newDirection = currentDirection;
    if (
      selectedDirectionInfo.isViolateBounderyFn(
        positionX + selectedDirectionInfo.horizontalDistance
      )
    ) {
      newPositionY = positionY + selectedDirectionInfo.verticalDistance;
      newDirection = reverseDirection(newDirection);
    } else {
      newPositionX = positionX + selectedDirectionInfo.horizontalDistance;
    }

    let sourcePosition = currentDirection;
    if (newDirection !== currentDirection) {
      sourcePosition = bottomDirection;
    }

    return {
      x: newPositionX,
      y: newPositionY,
      direction: newDirection,
      sourcePosition: sourcePosition
    };
  };

  const createNodes = details => {
    const nodeHeight = 100;
    const nodeWidth = 200;
    const marginHorizontal = 50;
    const marginVertical = 50;
    const maximumWidth = 800;
    const initialXPosition = 0;
    const initialYPosition = 0;
    let lastXPosition = initialXPosition;
    let lastYPosition = initialYPosition;

    var nodes = [];

    var direction = rightDirection;
    const directionInfo = {
      right: {
        isViolateBounderyFn: position => {
          return position > maximumWidth;
        },
        horizontalDistance: marginHorizontal + nodeWidth,
        verticalDistance: marginVertical + nodeHeight
      },
      left: {
        isViolateBounderyFn: position => {
          return position < 0;
        },
        horizontalDistance: -1 * (marginHorizontal + nodeWidth),
        verticalDistance: marginVertical + nodeHeight
      }
    };

    if (details.preprocess) {
      for (let i = 0; i < details.preprocess.length; i++) {
        if (i !== 0) {
          var generatedPosition = calculatePosition(
            lastXPosition,
            lastYPosition,
            direction,
            directionInfo
          );
          lastXPosition = generatedPosition.x;
          lastYPosition = generatedPosition.y;
          direction = generatedPosition.direction;
          nodes.push({
            id: "preprocess-" + i,
            type: "pipeline",
            width: nodeWidth,
            height: nodeHeight,
            position: { x: lastXPosition, y: lastYPosition },
            data: details.preprocess[i],
            style: { width: nodeWidth },
            sourcePosition: generatedPosition.sourcePosition
          });
        } else {
          nodes.push({
            id: "preprocess-" + i,
            type: "pipeline",
            width: nodeWidth,
            height: nodeHeight,
            position: { x: lastXPosition, y: lastYPosition },
            data: details.preprocess[i],
            style: { width: nodeWidth }
          });
        }
      }
    }

    generatedPosition = calculatePosition(
      lastXPosition,
      lastYPosition,
      direction,
      directionInfo
    );
    lastXPosition = generatedPosition.x;
    lastYPosition = generatedPosition.y;
    direction = generatedPosition.direction;

    nodes.push({
      id: "prediction",
      width: nodeWidth,
      height: nodeHeight,
      position: { x: lastXPosition, y: lastYPosition },
      data: { label: "Model Prediction" },
      sourcePosition: "right",
      targetPosition: "left"
    });

    if (details.postprocess) {
      for (let i = 0; i < details.postprocess.length; i++) {
        generatedPosition = calculatePosition(
          lastXPosition,
          lastYPosition,
          direction,
          directionInfo
        );
        lastXPosition = generatedPosition.x;
        lastYPosition = generatedPosition.y;
        direction = generatedPosition.direction;
        nodes.push({
          id: "postprocess-" + i,
          type: "pipeline",
          width: { nodeWidth },
          height: { nodeHeight },
          position: { x: lastXPosition, y: lastYPosition },
          data: details.postprocess[i]
        });
      }
    }
    return nodes;
  };
  const initialNodes = createNodes(tracingDetails);
  const initialEdges = createChainEdges(initialNodes);

  const [nodes, , onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const onConnect = params => setEdges(eds => addEdge(params, eds));

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
    <div style={{ height: 500 }}>
      <ReactFlow
        nodes={nodes}
        nodeTypes={nodeTypes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        onNodeClick={onNodeClicked}
        fitView
        attributionPosition="top-right">
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
