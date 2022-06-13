import {
  EuiAccordion,
  EuiCodeBlock,
  EuiFlexGroup,
  EuiFlexItem,
  EuiIcon,
  EuiTitle
} from "@elastic/eui";
import React from "react";

const yaml = require("js-yaml");

const AccordionButton = item => {
  return (
    <EuiFlexGroup gutterSize="s" alignItems="center" responsive={false}>
      <EuiFlexItem grow={false}>
        <EuiIcon type={item.icon} size="m" />
      </EuiFlexItem>

      <EuiFlexItem>
        <EuiTitle size="xs">
          <h3>{item.label}</h3>
        </EuiTitle>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};

export const PipelineNodeDetails = ({ details }) => {
  const items = [
    {
      label: "Spec",
      language: "yaml",
      icon: "tableDensityExpanded",
      data: details.spec !== null ? yaml.dump(details.spec) : undefined,
      initialIsOpen: false
    },
    {
      label: "Input",
      language: "json",
      icon: "logstashInput",
      data:
        details.input !== null
          ? JSON.stringify(details.input, null, 2)
          : undefined,
      initialIsOpen: false
    },
    {
      label: "Output",
      language: "json",
      icon: "logstashOutput",
      data:
        details.output !== null
          ? JSON.stringify(details.output, null, 2)
          : undefined,
      initialIsOpen: true
    }
  ];

  return (
    <>
      {items.map(
        (item, itemIdx) =>
          item.data && (
            <EuiAccordion
              id={`node-trace-details-${itemIdx}`}
              key={`node-trace-details-${itemIdx}`}
              className="euiAccordionForm"
              buttonClassName="euiAccordionForm__button"
              buttonContent={AccordionButton(item)}
              initialIsOpen={item.initialIsOpen}>
              <EuiCodeBlock
                language={item.language}
                fontSize="s"
                paddingSize="s"
                overflowHeight={600}>
                {item.data}
              </EuiCodeBlock>
            </EuiAccordion>
          )
      )}
    </>
  );
};
