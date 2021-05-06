import React, { useEffect } from "react";
import { EuiButtonIcon, EuiFieldText, EuiSpacer } from "@elastic/eui";
import { InMemoryTableForm, useOnChangeHandler } from "@gojek/mlp-ui";

export const VariablesInput = ({ variables, onChangeHandler, errors = {} }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const items = [
    ...variables.map((v, idx) => ({ idx, ...v })),
    { idx: variables.length }
  ];

  const onDeleteVariable = idx => () => {
    variables.splice(idx, 1);
    onChangeHandler(variables);
  };

  const getRowProps = item => {
    const { idx } = item;
    const isInvalid = !!errors[idx];
    return {
      className: isInvalid ? "euiTableRow--isInvalid" : "",
      "data-test-subj": `row-${idx}`
    };
  };

  const columns = [
    {
      name: "Name",
      field: "name",
      width: "30%",
      render: (name, item) => (
        <EuiFieldText
          controlOnly
          className="inlineTableInput"
          placeholder="Name"
          value={name || ""}
          onChange={e => onChange(`${item.idx}.name`)(e.target.value)}
        />
      )
    },
    {
      name: "Type",
      field: "type",
      width: "30%",
      render: (type, item) => (
        <EuiFieldText
          controlOnly
          className="inlineTableInput"
          placeholder="Type"
          value={type || ""}
          onChange={e => onChange(`${item.idx}.type`)(e.target.value)}
        />
      )
    },
    {
      name: "Value",
      field: "value",
      width: "30%",
      render: (value, item) => (
        <EuiFieldText
          controlOnly
          className="inlineTableInput"
          placeholder="Value"
          value={value || ""}
          onChange={e => onChange(`${item.idx}.value`)(e.target.value)}
        />
      )
    },
    {
      width: "10%",
      actions: [
        {
          render: item =>
            item.idx < items.length - 1 ? (
              <EuiButtonIcon
                size="s"
                color="danger"
                iconType="trash"
                onClick={onDeleteVariable(item.idx)}
                aria-label="Remove variable"
              />
            ) : (
              <div />
            )
        }
      ]
    }
  ];

  return (
    <InMemoryTableForm
      columns={columns}
      rowProps={getRowProps}
      items={items}
      hasActions={true}
      errors={errors}
      renderErrorHeader={key => `Row ${parseInt(key) + 1}`}
    />
  );
};
