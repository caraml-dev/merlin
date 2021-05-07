import React, { useEffect, useMemo } from "react";
import { EuiButtonIcon, EuiFieldText, EuiFormRow, EuiText } from "@elastic/eui";
import { InMemoryTableForm, useOnChangeHandler } from "@gojek/mlp-ui";

export const RenameColumns = ({ columns, onChangeHandler, errors = {} }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const items = [
    ...Object.keys(columns).map((v, idx) => ({
      idx,
      before: v,
      after: columns[v]
    })),
    { idx: Object.keys(columns).length }
  ];

  const onDeleteColumn = idx => () => {
    columns.splice(idx, 1);
    onChangeHandler(columns);
  };

  const getRowProps = item => {
    const { idx } = item;
    const isInvalid = !!errors[idx];
    return {
      className: isInvalid ? "euiTableRow--isInvalid" : "",
      "data-test-subj": `row-${idx}`
    };
  };

  // useEffect(() => {
  //   console.log("COL", columns);
  // }, [columns]);

  // useEffect(() => {
  //   console.log("ITEMS", items);
  // }, [items]);

  const tableColumns = [
    {
      name: "Before",
      field: "before",
      width: "45%",
      render: (before, item) => (
        <EuiFieldText
          placeholder="Column Name Before"
          value={before || ""}
          onChange={e => {
            // if (!columns.hasOwnProperty(e.target.value)) {
            let newColumns = columns;
            newColumns[e.target.value] = item.after;
            delete newColumns[before];
            onChange(`renameColumns`)(newColumns);
            // }
          }}
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
          onChange={e => {
            onChange(`renameColumns`)({
              ...columns,
              [item.before ? item.before : ""]: e.target.value
            });
          }}
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
