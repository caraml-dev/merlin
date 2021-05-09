import React from "react";
import { EuiButtonIcon, EuiFieldText } from "@elastic/eui";
import { InMemoryTableForm, useOnChangeHandler } from "@gojek/mlp-ui";

export const UpdateColumns = ({ columns, onChangeHandler, errors = {} }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const items = [
    ...columns.map((v, idx) => ({ idx, ...v })),
    { idx: columns.length }
  ];

  const onDeleteColumn = idx => () => {
    columns.splice(idx, 1);
    onChange("update")(columns);
  };

  const getRowProps = item => {
    const { idx } = item;
    const isInvalid = !!errors[idx];
    return {
      className: isInvalid ? "euiTableRow--isInvalid" : "",
      "data-test-subj": `row-${idx}`
    };
  };

  const tableColumns = [
    {
      name: "Column",
      field: "column",
      width: "45%",
      render: (column, item) => (
        <EuiFieldText
          placeholder="Column Name"
          value={column || ""}
          onChange={e => onChange(`${item.idx}.column`)(e.target.value)}
        />
      )
    },
    {
      name: "Expression",
      field: "expression",
      width: "45%",
      render: (expression, item) => (
        <EuiFieldText
          placeholder="Expression"
          value={expression || ""}
          onChange={e => onChange(`${item.idx}.expression`)(e.target.value)}
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
                onClick={onDeleteColumn(item.idx)}
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
      columns={tableColumns}
      rowProps={getRowProps}
      items={items}
      hasActions={true}
      errors={errors}
      renderErrorHeader={key => `Row ${parseInt(key) + 1}`}
    />
  );
};
