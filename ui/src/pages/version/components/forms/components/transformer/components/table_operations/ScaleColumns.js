import React from "react";
import { useState } from "react";
import {
  EuiButtonIcon,
  EuiFieldText,
  EuiSuperSelect,
  EuiFormRow
} from "@elastic/eui";
import { InMemoryTableForm, useOnChangeHandler } from "@gojek/mlp-ui";
import { get } from "@gojek/mlp-ui";

export const ScaleColumns = ({ columns, onChangeHandler, errors = {} }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  //An array to keep track of scaler type for each row
  const [scalerArr, setScalerArr] = useState([]);

  //Manage updating of array state
  const updateArr = (idx, v) => {
    let newArr = scalerArr;
    newArr[idx] = v;
    setScalerArr(newArr);
  };

  //Manage deleting of array element of state
  const deleteArrElement = idx => {
    let newArr = scalerArr;
    newArr.splice(idx, 1);
    setScalerArr(newArr);
  };

  const items = [
    ...columns.map((v, idx) => ({ idx, ...v })),
    { idx: columns.length }
  ];

  const onDeleteColumn = idx => () => {
    columns.splice(idx, 1);
    deleteArrElement(idx);
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
    { value: "standardScalerConfig", inputDisplay: "Standard Scaler" },
    { value: "minMaxScalerConfig", inputDisplay: "Min-Max Scaler" }
  ];

  const tableColumns = [
    {
      name: "Column",
      field: "column",
      width: "20%",
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
      width: "20%",
      render: (scaler, item) => {
        return (
          <EuiSuperSelect
            options={scalerOptions}
            valueOfSelected={scalerArr[item.idx] || "Select Scaler"}
            onChange={value => updateArr(item.idx, value)}
            hasDividers
          />
        );
      }
    },
    {
      name: "Param 1",
      field: "param1",
      width: "20%",
      render: (param1, item) => {
        let paramName =
          scalerArr[item.idx] === "standardScalerConfig"
            ? "Mean"
            : scalerArr[item.idx] === "minMaxScalerConfig"
            ? "Min"
            : "";

        return (
          <EuiFormRow
            fullWidth
            label={paramName}
            isInvalid={!!errors}
            error={get(errors, "0")}>
            <EuiFieldText
              placeholder="Param 1"
              onChange={e =>
                onChange(
                  `${item.idx}.${
                    scalerArr[item.idx]
                  }.${paramName.toLowerCase()}`
                )(Number(e.target.value))
              }
            />
          </EuiFormRow>
        );
      }
    },
    {
      name: "Param 2",
      field: "param2",
      width: "20%",
      render: (param2, item) => {
        let paramName =
          scalerArr[item.idx] === "standardScalerConfig"
            ? "Std"
            : scalerArr[item.idx] === "minMaxScalerConfig"
            ? "Max"
            : "";

        return (
          <EuiFormRow
            fullWidth
            label={paramName}
            isInvalid={!!errors}
            error={get(errors, "0")}>
            <EuiFieldText
              placeholder="Param 2"
              onChange={e =>
                onChange(
                  `${item.idx}.${
                    scalerArr[item.idx]
                  }.${paramName.toLowerCase()}`
                )(Number(e.target.value))
              }
            />
          </EuiFormRow>
        );
      }
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
