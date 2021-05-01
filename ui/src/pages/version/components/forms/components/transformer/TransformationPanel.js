import React from "react";
import { EuiButton, EuiFlexGroup, EuiFlexItem, EuiSpacer } from "@elastic/eui";
import { Panel } from "../Panel";
import { TableTransformationCard } from "./TableTransformationCard";
import { TableJoinCard } from "./TableJoinCard";

export const TransformationPanel = ({ onChangeHandler, errors = {} }) => {
  return (
    <Panel title="Transformation">
      <EuiFlexGroup direction="column" gutterSize="s">
        <EuiFlexItem>
          <TableTransformationCard />
          <EuiSpacer size="s" />
        </EuiFlexItem>
        <EuiFlexItem>
          <TableJoinCard />
          <EuiSpacer size="s" />
        </EuiFlexItem>
        <EuiFlexItem>
          <EuiButton onClick={() => {}}>+ Add Transformation</EuiButton>
        </EuiFlexItem>
      </EuiFlexGroup>
    </Panel>
  );
};
