import {
  EuiAccordion,
  EuiCodeBlock,
  EuiFlexGroup,
  EuiFlexItem,
  EuiIcon,
  EuiInMemoryTable,
  EuiSpacer,
  EuiTabbedContent,
  EuiText,
  EuiTitle
} from "@elastic/eui";
import React, { useEffect, useState } from "react";

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

const Table = ({ data }) => {
  const [columns, setColumns] = useState([]);
  const [items, setItems] = useState([]);
  const [tableName, setTableName] = useState();

  useEffect(() => {
    // Get how many keys the data has so we can determine whether the data is table, variable, or Pandas format (which not supported yet)
    const keys = Object.keys(data);

    // Skip Pandas format fow now.
    if (keys[0] === "instances") {
      return;
    }

    // If the object only has 1 key and the value of this key is array, this is table.
    // For example:
    // { "driver_table": [
    //     {
    //       "driver_id": 1
    //     },
    //     {
    //       "driver_id": 2
    //     },
    //   ]
    // }
    if (keys.length === 1 && Array.isArray(data[keys[0]])) {
      const tableName = keys[0];
      setTableName(tableName);

      let columns = [];
      Object.keys(data[tableName][0]).forEach(key => {
        columns.push({
          field: key,
          name: key,
          truncateText: false
        });
      });
      setColumns(columns);

      setItems(data[tableName]);
      return;
    }

    // This part to handle variable type output
    // For example:
    // {
    //   "variable_1": 1,
    //   "variable_2": 2
    // }
    if (
      keys.length > 0 &&
      typeof data === "object" &&
      !Array.isArray(data) &&
      data !== null
    ) {
      let columns = [];
      keys.forEach(key => {
        columns.push({
          field: key,
          name: key,
          truncateText: false
        });
      });
      setColumns(columns);

      setItems([data]);
      return;
    }
  }, [data]);

  return (
    <EuiFlexGroup gutterSize="none" direction="column">
      <EuiSpacer size="s" />
      {tableName && (
        <EuiFlexItem grow={true}>
          <EuiFlexGroup justifyContent="center">
            <EuiFlexItem grow={false}>
              <EuiText size="xs">
                <strong>{tableName}</strong>
              </EuiText>
            </EuiFlexItem>
          </EuiFlexGroup>
        </EuiFlexItem>
      )}
      {columns.length > 0 ? (
        <EuiFlexItem grow={true}>
          <EuiInMemoryTable columns={columns} items={items} />
        </EuiFlexItem>
      ) : (
        <EuiText size="xs">
          The data format is not supported yet. Please use see the raw data.
        </EuiText>
      )}
    </EuiFlexGroup>
  );
};

const Code = ({ content, language }) => {
  return (
    <EuiCodeBlock
      language={language}
      fontSize="s"
      paddingSize="s"
      overflowHeight={600}>
      {content}
    </EuiCodeBlock>
  );
};

export const PipelineNodeDetails = ({ details }) => {
  const [inputTabs, setInputTabs] = useState([]);
  useEffect(() => {
    if (details.input === null) {
      setInputTabs([]);
      return;
    }

    let tabs = [];
    if (details.operation_type === "table_join_op") {
      let tables = [];
      Object.keys(details.input).forEach(tableName => {
        tables.push(
          <Table
            key={tableName}
            data={{ [tableName]: details.input[tableName] }}
          />
        );
      });
      tabs.push({
        id: "input-table",
        name: "Table",
        content: tables
      });
    } else {
      tabs.push({
        id: "input-table",
        name: "Table",
        content: <Table data={details.input} />
      });
    }

    tabs.push({
      id: "input-json",
      name: "Raw",
      content: (
        <Code
          content={JSON.stringify(details.input, null, 2)}
          language="json"
        />
      )
    });
    setInputTabs(tabs);
  }, [details, details.input, setInputTabs]);

  const [outputTabs, setOutputTabs] = useState([]);
  useEffect(() => {
    if (details.output === null) {
      setOutputTabs([]);
      return;
    }

    let tabs = [];
    if (details.output["instances"] === undefined && details.operation_type !== "upi_preprocess_output_op" && details.operation_type !== "upi_postprocess_output_op") {
      let tables = []
      Object.keys(details.output).forEach(name => {
        tables.push(
          <Table
            key={name}
            data={{ [name]: details.output[name] }}
          />
        );
      });

      tabs.push({
        id: "output-table",
        name: "Table",
        content: tables
      });
    }


    tabs.push({
      id: "output-json",
      name: "Raw",
      content: (
        <Code
          content={JSON.stringify(details.output, null, 2)}
          language="json"
        />
      )
    });
    setOutputTabs(tabs);
  }, [details.output, setOutputTabs, details.operation_type]);

  const items = [
    {
      label: "Spec",
      language: "yaml",
      icon: "tableDensityExpanded",
      initialIsOpen: false,
      tabs: [
        {
          id: "spec-yaml",
          content: (
            <Code
              content={
                details.spec !== null ? yaml.dump(details.spec) : undefined
              }
              language="yaml"
            />
          )
        }
      ]
    },
    {
      label: "Input",
      icon: "logstashInput",
      initialIsOpen: false,
      tabs: inputTabs
    },
    {
      label: "Output",
      icon: "logstashOutput",
      initialIsOpen: true,
      tabs: outputTabs
    }
  ];

  return (
    <>
      {items.map(
        (item, itemIdx) =>
          item.tabs.length > 0 && (
            <EuiAccordion
              id={`node-trace-details-${itemIdx}`}
              key={`node-trace-details-${itemIdx}`}
              className="euiAccordionForm"
              buttonClassName="euiAccordionForm__button"
              buttonContent={AccordionButton(item)}
              initialIsOpen={item.initialIsOpen}>
              {item.tabs.length > 1 ? (
                <EuiTabbedContent
                  tabs={item.tabs}
                  initialSelectedTab={item.tabs[0]}
                  autoFocus="selected"
                  size="s"
                />
              ) : (
                item.tabs[0].content
              )}
            </EuiAccordion>
          )
      )}
    </>
  );
};
