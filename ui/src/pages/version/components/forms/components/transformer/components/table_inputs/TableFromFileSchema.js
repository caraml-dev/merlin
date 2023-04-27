import React from "react";
import { EuiButtonIcon, EuiFieldText, EuiSuperSelect } from "@elastic/eui";
import { get, InMemoryTableForm, useOnChangeHandler } from "@caraml-dev/ui-lib";
import "../../RowCell.scss";

export const TableFromFileSchema = ({
  columns,
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const items = [
    ...columns.map((v, idx) => ({ idx, ...v })),
    { idx: columns.length }
  ];

  const onDeleteVariable = idx => () => {
    columns.splice(idx, 1);
    onChange(`baseTable.fromFile.schema`)(columns);
  };

  const typeOptions = [
    { value: "STRING", inputDisplay: "String" },
    { value: "INT", inputDisplay: "Integer" },
    { value: "FLOAT", inputDisplay: "Float" },
    { value: "BOOL", inputDisplay: "Boolean" }
  ];

  const onVariableChange = (idx, field, value) => {
    let newItem = { ...items[idx], [field]: value };

    if (newItem.name === undefined) {
      newItem.name = "";
    }
    if (newItem.idx !== undefined) {
      delete newItem.idx;
    }

    switch (newItem.type) {
      case "STRING":
        newItem = { ...newItem, type: "STRING" };
        break;
      case "INT":
        newItem = { ...newItem, type: "INT" };
        break;
      case "FLOAT":
        newItem = { ...newItem, type: "FLOAT" };
        break;
      case "BOOL":
        newItem = { ...newItem, type: "BOOL" };
        break;
      default:
        break;
    }

    onChange(`baseTable.fromFile.schema.${idx}`)(newItem);
  };

  const cols = [
    {
      name: "Name",
      field: "name",
      width: "50%",
      render: (name, item) => (
        <EuiFieldText
          placeholder="Name"
          value={name || ""}
          onChange={e =>
            onChange(`baseTable.fromFile.schema.${item.idx}.name`)(
              e.target.value
            )
          }
          isInvalid={
            !!get(errors, `baseTable.fromFile.schema.${item.idx}.name`)
          }
        />
      )
    },
    {
      name: "Type",
      field: "type",
      width: "40%",
      render: (type, item) => (
        <EuiSuperSelect
          options={typeOptions}
          valueOfSelected={type || ""}
          onChange={value => onVariableChange(item.idx, "type", value)}
          isInvalid={
            !!get(errors, `baseTable.fromFile.schema.${item.idx}.type`)
          }
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

  const getRowProps = item => {
    const { idx } = item;
    const isInvalid = !!errors[idx];
    return {
      className: isInvalid ? "euiTableRow--isInvalid" : "",
      "data-test-subj": `row-${idx}`
    };
  };

  return (
    <InMemoryTableForm
      columns={cols}
      rowProps={getRowProps}
      items={items}
      hasActions={true}
      errors={errors}
      renderErrorHeader={key => `Row ${parseInt(key) + 1}`}
    />
  );
};
