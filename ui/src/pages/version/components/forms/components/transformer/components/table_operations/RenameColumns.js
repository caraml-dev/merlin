import React, { useEffect, useState } from "react";
import { EuiButtonIcon, EuiFieldText, EuiFormRow } from "@elastic/eui";
import { InMemoryTableForm, useOnChangeHandler } from "@caraml-dev/ui-lib";

export const RenameColumns = ({ columns, onChangeHandler, errors = {} }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const [items, setItems] = useState([
    ...Object.keys(columns).map((v, idx) => ({
      idx,
      before: v,
      after: columns[v]
    })),
    { idx: Object.keys(columns).length }
  ]);

  useEffect(
    () => {
      let newColumns = {};
      items
        .filter(item => item.before && item.before !== "")
        .forEach(item => {
          newColumns[item.before] = item.after;
        });
      onChange("renameColumns")(newColumns);
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
        field === "before" &&
        items[items.length - 1].before &&
        items[items.length - 1].before.trim()
          ? [...items, { idx: items.length }]
          : [...items]
      );
    };
  };

  const tableColumns = [
    {
      name: "Before",
      field: "before",
      width: "45%",
      render: (before, item) => (
        <EuiFieldText
          placeholder="Column Name Before"
          value={before || ""}
          onChange={onChangeRow(item.idx, "before")}
        />
      )
    },
    {
      name: "After",
      field: "after",
      width: "45%",
      render: (after, item) => (
        <EuiFieldText
          placeholder="Column Name After"
          value={after || ""}
          onChange={onChangeRow(item.idx, "after")}
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
      label="Columns to be renamed"
      helpText="This operations will rename one column to another.">
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
