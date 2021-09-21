import React, { useEffect, useState } from "react";
import { EuiButtonIcon, EuiFieldText, EuiFormRow } from "@elastic/eui";
import { InMemoryTableForm, useOnChangeHandler } from "@gojek/mlp-ui";

export const EncodeColumns = ({ columns, onChangeHandler, errors = {} }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const [items, setItems] = useState([
    ...Object.keys(columns).map((v, idx) => ({
      idx,
      column: v,
      encoder: columns[v]
    })),
    { idx: Object.keys(columns).length }
  ]);

  useEffect(
    () => {
      let newColumns = {};
      items
        .filter(item => item.column && item.column !== "")
        .forEach(item => {
          newColumns[item.column] = item.encoder;
        });
      onChange("encodeColumns")(newColumns);
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [items]
  );

  const onDeleteItem = idx => () => {
    items.splice(idx, 1);
    setItems([...items.map((v, idx) => ({ ...v, idx }))]);
  };

  const getRowProps = item => {
    const { idx } = item;
    const isInvalid = !!errors[idx];
    return {
      className: isInvalid ? "euiTableRow--isInvalid" : "",
      "data-test-subj": `row-${idx}`
    };
  };

  const onChangeRow = (idx, field) => {
    return e => {
      items[idx] = { ...items[idx], [field]: e.target.value };

      setItems(_ =>
        field === "column" &&
        items[items.length - 1].column &&
        items[items.length - 1].column.trim()
          ? [...items, { idx: items.length }]
          : [...items]
      );
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
          onChange={onChangeRow(item.idx, "column")}
        />
      )
    },
    {
      name: "Encoder",
      field: "encoder",
      width: "45%",
      render: (encoder, item) => (
        <EuiFieldText
          placeholder="Encoder name"
          value={encoder || ""}
          onChange={onChangeRow(item.idx, "encoder")}
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
                onClick={onDeleteItem(item.idx)}
                aria-label="Remove column"
              />
            ) : (
              <div />
            )
        }
      ]
    }
  ];

  return (
    <EuiFormRow
      fullWidth
      label="Columns to be encoded"
      helpText="This operation will map a column to an encoder.">
      <InMemoryTableForm
        columns={tableColumns}
        rowProps={getRowProps}
        items={items}
        hasActions={true}
        errors={errors}
        renderErrorHeader={key => `Row ${parseInt(key) + 1}`}
      />
    </EuiFormRow>
  );
};
