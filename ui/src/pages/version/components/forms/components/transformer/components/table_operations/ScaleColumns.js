import React from "react";
import { EuiButtonIcon, EuiFieldText, EuiSuperSelect } from "@elastic/eui";
import { InMemoryTableForm, useOnChangeHandler } from "@gojek/mlp-ui";

export const ScaleColumns = ({ columns, onChangeHandler, errors = {} }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const items = [
    ...columns.map((v, idx) => ({ idx, ...v })),
    { idx: columns.length }
  ];

  const onDeleteColumn = idx => () => {
    columns.splice(idx, 1);
    onChange("scaleColumns")(columns);
  };

  const getRowProps = item => {
    const { idx } = item;
    const isInvalid = !!errors[idx];
    return {
      className: isInvalid ? "euiTableRow--isInvalid" : "",
      "data-test-subj": `row-${idx}`
    };
  };

  const scalerOptions = [
    { value: "STANDARD", inputDisplay: "Standard Scaler" },
    { value: "MIN_MAX", inputDisplay: "Min-Max Scaler" }
  ];

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
      name: "Scaler",
      field: "scaler",
      width: "45%",
      render: (scaler, item) => (
        <EuiSuperSelect
          options={scalerOptions}
          valueOfSelected={scaler || ""}
          onChange={value => onChange(`${item.idx}.scaler`)(value)}
          hasDividers
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
