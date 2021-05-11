import React, { useContext, useEffect, useState } from "react";
import { EuiPanel, EuiTabbedContent } from "@elastic/eui";
import { TransformationGraph } from "./components/TransformationGraph";
import { TransformationSpec } from "./components/TransformationSpec";
import { FormContext } from "@gojek/mlp-ui";

export const PipelineSidebarPanel = ({ importEnabled = true }) => {
  const {
    data: { transformer }
  } = useContext(FormContext);

  const [tabs, setTabs] = useState([
    {
      id: "spec-panel",
      name: "YAML Specification",
      content: <TransformationSpec importEnabled={importEnabled} />
    }
  ]);

  useEffect(
    () => {
      if (transformer.type_on_ui === "standard") {
        setTabs([
          tabs[0],
          {
            id: "graph-panel",
            name: "Transformation Graph",
            content: <TransformationGraph />
          }
        ]);
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [transformer]
  );

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
