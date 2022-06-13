import { EuiPanel, EuiTabbedContent } from "@elastic/eui";
import { FormContext } from "@gojek/mlp-ui";
import React, { useContext, useEffect, useState } from "react";
import { TransformationSpec } from "./components/TransformationSpec";

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
        // TODO: Rework Transformation Graph to match the operation trace and use React Flow
        // setTabs([
        //   tabs[0],
        //   {
        //     id: "graph-panel",
        //     name: "Transformation Graph",
        //     content: <TransformationGraph />
        //   }
        // ]);
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
