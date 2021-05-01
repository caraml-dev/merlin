import React from "react";
import { EuiButton, EuiFlexGroup, EuiFlexItem, EuiSpacer } from "@elastic/eui";
import { useOnChangeHandler } from "@gojek/mlp-ui";
import { Panel } from "../Panel";
import { VariableInputCard } from "./VariableInputCard";
import { TableCreationCard } from "./TableCreationCard";

export const InputPanel = ({ inputs, errors = {} }) => {
  return (
    <Panel title="Input">
      <EuiFlexGroup direction="column" gutterSize="s">
        <EuiFlexItem>
          <TableCreationCard />
          <EuiSpacer size="s" />
        </EuiFlexItem>
        <EuiFlexItem>
          <VariableInputCard />
          <EuiSpacer size="s" />
        </EuiFlexItem>
      </EuiFlexGroup>
    </Panel>
  );
};
