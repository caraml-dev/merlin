import React from "react";
import { EuiButtonIcon, EuiFieldText } from "@elastic/eui";
import { InMemoryTableForm, useOnChangeHandler } from "@gojek/mlp-ui";
import { ColumnsComboBox } from "./ColumnsComboBox";

export const EncodeColumns = ({ columns, onChangeHandler, errors = {} }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const items = [
    ...columns.map((v, idx) => ({ idx, ...v })),
    { idx: columns.length }
  ];

  const onDeleteColumn = idx => () => {
    columns.splice(idx, 1);
    onChange("encodeColumns")(columns);
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
      name: 'Columns (Press "return" or "comma" for new entry)',
      field: "columns",
      width: "50%",
      render: (columns, item) => (
        <ColumnsComboBox
          columns={columns || []}
          onChange={onChange(`${item.idx}.columns`)}
        />
      )
    },
    {
      name: "Encoder",
      field: "encoder",
      width: "40%",
      render: (encoder, item) => (
        <EuiFieldText
          placeholder="Encoder Name"
          value={encoder || ""}
          onChange={e => onChange(`${item.idx}.encoder`)(e.target.value)}
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
