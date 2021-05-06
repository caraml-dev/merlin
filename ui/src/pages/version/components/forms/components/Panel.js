import React from "react";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiPanel,
  EuiSpacer,
  EuiTitle
} from "@elastic/eui";

const PanelContent = ({ contentWidth = "70%", children }) => (
  <EuiFlexGroup direction="row" justifyContent="center">
    <EuiFlexItem grow={false} style={{ width: contentWidth }}>
      {children}
    </EuiFlexItem>
  </EuiFlexGroup>
);

export const Panel = ({ title, contentWidth, color = "plain", children }) => {
  return (
    <EuiPanel grow={false} color={color}>
      <EuiTitle size="xs">
        <h4>{title}</h4>
      </EuiTitle>

      <PanelContent contentWidth={contentWidth}>
        {children}
        <EuiSpacer size="s" />
      </PanelContent>
    </EuiPanel>
  );
};
