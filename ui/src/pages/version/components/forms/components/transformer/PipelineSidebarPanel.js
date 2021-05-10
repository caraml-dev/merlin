import React from "react";
import { EuiPanel, EuiTabbedContent } from "@elastic/eui";
import { TransformationGraph } from "./components/TransformationGraph";
import { TransformationSpec } from "./components/TransformationSpec";

export const PipelineSidebarPanel = () => {
  const tabs = [
    {
      id: "spec-panel",
      name: "YAML Specification",
      content: <TransformationSpec />
    },
    {
      id: "graph-panel",
      name: "Transformation Graph",
      content: <TransformationGraph />
    }
  ];

  return (
    <EuiPanel grow={false}>
      <EuiTabbedContent
        tabs={tabs}
        initialSelectedTab={tabs[0]}
        autoFocus="selected"
        size="s"
      />
    </EuiPanel>
  );
};
