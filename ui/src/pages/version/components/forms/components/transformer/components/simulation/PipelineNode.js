import { EuiButton, EuiFlexGroup, EuiFlexItem, EuiText } from "@elastic/eui";
import React from "react";
import { Handle } from "react-flow-renderer";
import { Panel } from "../../../Panel";

function PipelineNode({ data }) {
  return (
    <>
      <Handle type="target" position="left" id="target-left" />
      <Handle type="target" position="right" id="target-right" />
      <Handle type="target" position="top" id="target-top" />
      <Panel>
        <EuiFlexGroup
          direction="column"
          gutterSize="m"
          justifyContent="center"
          alignItems="center">
          <EuiFlexItem>
            <EuiText>{data.operation_type}</EuiText>
          </EuiFlexItem>
          <EuiFlexItem>
            <EuiButton fill={false} size="s" color="primary">
              <EuiText size="xs">Show Details</EuiText>
            </EuiButton>
          </EuiFlexItem>
        </EuiFlexGroup>
      </Panel>
      <Handle type="source" position="right" id="source-right" />
      <Handle type="source" position="left" id="source-left" />
      <Handle type="source" position="bottom" id="source-bottom" />
    </>
  );
}

export default PipelineNode;
