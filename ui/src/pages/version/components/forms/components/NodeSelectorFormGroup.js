import React, { Fragment, useState } from "react";
import {
  EuiButtonEmpty,
  EuiButtonIcon,
  EuiDescribedFormGroup,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiSpacer,
  EuiText,
} from "@elastic/eui";

const toRows = (obj) => Object.entries(obj || {}).map(([key, value]) => ({ key, value }));

const toObject = (rows) => {
  const obj = {};
  rows.forEach((r) => {
    const key = (r.key || "").trim();
    if (key !== "") obj[key] = r.value || "";
  });
  return obj;
};

/**
 * NodeSelectorFormGroup - edit the pod nodeSelector as a list of label key/value pairs.
 *
 * Props
 *   nodeSelector    – object of label key -> value (may be undefined / {})
 *   onChangeHandler – called with the rebuilt object whenever the list changes
 */
export const NodeSelectorFormGroup = ({ nodeSelector = {}, onChangeHandler }) => {
  const [rows, setRows] = useState(() => toRows(nodeSelector));

  const push = (next) => {
    setRows(next);
    onChangeHandler(toObject(next));
  };

  const onChangeCell = (idx, field) => (e) => {
    const value = e.target.value;
    push(rows.map((r, i) => (i === idx ? { ...r, [field]: value } : r)));
  };

  const onAddRow = () => push([...rows, { key: "", value: "" }]);

  const onDeleteRow = (idx) => () => push(rows.filter((_, i) => i !== idx));

  return (
    <EuiDescribedFormGroup
      title={<p>Node Selector</p>}
      description={
        <Fragment>
          <EuiText size="s">
            Pin the pods onto nodes whose labels match every entry. Combine with
            a matching toleration to run on dedicated / tainted node pools.
          </EuiText>
        </Fragment>
      }
      fullWidth
    >
      <EuiSpacer size="s" />
      {rows.map((row, idx) => (
        <EuiFlexGroup key={idx} gutterSize="s" alignItems="center">
          <EuiFlexItem>
            <EuiFieldText
              compressed
              placeholder="label key (e.g. pool)"
              value={row.key || ""}
              onChange={onChangeCell(idx, "key")}
            />
          </EuiFlexItem>
          <EuiFlexItem>
            <EuiFieldText
              compressed
              placeholder="value (e.g. workload-optimized)"
              value={row.value || ""}
              onChange={onChangeCell(idx, "value")}
            />
          </EuiFlexItem>
          <EuiFlexItem grow={false}>
            <EuiButtonIcon
              size="s"
              color="danger"
              iconType="trash"
              onClick={onDeleteRow(idx)}
              aria-label="Remove node selector"
            />
          </EuiFlexItem>
        </EuiFlexGroup>
      ))}
      <EuiSpacer size="s" />
      <EuiFormRow>
        <EuiButtonEmpty size="s" iconType="plusInCircle" onClick={onAddRow}>
          Add node selector
        </EuiButtonEmpty>
      </EuiFormRow>
    </EuiDescribedFormGroup>
  );
};
