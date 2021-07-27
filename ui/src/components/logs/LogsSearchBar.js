import React, { useMemo } from "react";
import { EuiSearchBar, EuiToolTip } from "@elastic/eui";

export const LogsSearchBar = ({ componentTypes, params, setParams }) => {
  const filters = useMemo(() => {
    return [
      {
        type: "field_value_toggle_group",
        field: "component_type",
        items: [
          {
            value: "image_builder",
            name: (
              <EuiToolTip position="top" content="TODO">
                <p>Image Builder</p>
              </EuiToolTip>
            )
          },
          {
            value: "model",
            name: (
              <EuiToolTip position="top" content="TODO">
                <p>Model</p>
              </EuiToolTip>
            )
          },
          {
            value: "transformer",
            name: (
              <EuiToolTip position="top" content="TODO">
                <p>Transformer</p>
              </EuiToolTip>
            )
          },
          {
            value: "batch_job_driver",
            name: (
              <EuiToolTip position="top" content="TODO">
                <p>Batch Job Driver</p>
              </EuiToolTip>
            )
          },
          {
            value: "batch_job_executor",
            name: (
              <EuiToolTip position="top" content="TODO">
                <p>Batch Job Executor</p>
              </EuiToolTip>
            )
          }
        ].filter(option => componentTypes.includes(option.value))
      },
      {
        type: "field_value_selection",
        field: "tail_lines",
        name: "Log Tail",
        multiSelect: false,
        options: [
          {
            value: "100",
            name: "Last 100 records"
          },
          {
            value: "1000",
            name: "Last 1000 records"
          },
          {
            value: "",
            name: "From the container start"
          }
        ]
      },
      {
        type: "field_value_selection",
        field: "prefix",
        name: "Prefix",
        multiSelect: false,
        options: [
          {
            value: "",
            name: "No prefix"
          },
          {
            value: "container",
            name: "Container name"
          },
          {
            value: "pod",
            name: "Pod name"
          },
          {
            value: "pod_and_container",
            name: "Pod + container name"
          }
        ]
      }
    ];
  }, [componentTypes]);

  const queryString = useMemo(() => {
    return Object.entries(params)
      .map(([k, v]) => `${k}:"${v}"`)
      .join(" ");
  }, [params]);

  const onChange = ({ query, error }) => {
    if (!error) {
      const newParams = {
        ...params,
        ...query.ast.clauses.reduce((acc, { field, value }) => {
          acc[field] = value;
          return acc;
        }, {})
      };

      if (JSON.stringify(newParams) !== JSON.stringify(params)) {
        setParams(newParams);
      }
    }
  };

  return (
    <EuiSearchBar
      query={queryString}
      box={{ readOnly: true }}
      filters={filters}
      onChange={onChange}
    />
  );
};
