import React, { useEffect, useState } from "react";
import { EuiButtonIcon, EuiFieldText, EuiFormRow } from "@elastic/eui";
import { InMemoryTableForm, useOnChangeHandler } from "@gojek/mlp-ui";

export const OrdinalEncoderMapper = ({
  mappings,
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const [items, setItems] = useState([
    ...Object.keys(mappings).map((v, idx) => ({
      idx,
      original: v,
      transformed: mappings[v]
    })),
    { idx: Object.keys(mappings).length }
  ]);

  useEffect(
    () => {
      let newMappings = {};
      items
        .filter(item => item.original && item.original !== "")
        .forEach(item => {
          newMappings[item.original] = item.transformed;
        });
      onChange("mapping")(newMappings);
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
        field === "original" &&
        items[items.length - 1].original &&
        items[items.length - 1].original.trim()
          ? [...items, { idx: items.length }]
          : [...items]
      );
    };
  };

  const tableColumns = [
    {
      name: "Original Value",
      field: "original",
      width: "45%",
      render: (original, item) => (
        <EuiFieldText
          placeholder="Value to encode"
          value={original || ""}
          onChange={onChangeRow(item.idx, "original")}
        />
      )
    },
    {
      name: "Transformed Value",
      field: "transformed",
      width: "45%",
      render: (transformed, item) => (
        <EuiFieldText
          placeholder="Value to map to"
          value={transformed || ""}
          onChange={onChangeRow(item.idx, "transformed")}
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
                aria-label="Remove mapping"
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
