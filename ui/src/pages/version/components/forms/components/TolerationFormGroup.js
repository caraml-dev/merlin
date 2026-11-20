import React, { Fragment } from "react";
import {
  EuiButtonIcon,
  EuiDescribedFormGroup,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiSelect,
  EuiSpacer,
  EuiText,
} from "@elastic/eui";
import { InMemoryTableForm, useOnChangeHandler } from "@caraml-dev/ui-lib";

const OPERATOR_OPTIONS = [
  { value: "", text: "—" },
  { value: "Equal", text: "Equal" },
  { value: "Exists", text: "Exists" },
];

const EFFECT_OPTIONS = [
  { value: "", text: "—  (Any)" },
  { value: "NoSchedule", text: "NoSchedule" },
  { value: "PreferNoSchedule", text: "PreferNoSchedule" },
  { value: "NoExecute", text: "NoExecute" },
];

/**
 * TolerationFormGroup - renders an inline table for adding/removing Kubernetes tolerations.
 *
 * Props
 *   tolerations    – array of toleration objects (may be undefined / [])
 *   onChangeHandler – function to call when the list changes
 *   errors          – validation errors (optional)
 */
export const TolerationFormGroup = ({
  tolerations = [],
  onChangeHandler,
  errors = {},
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const items = [
    ...tolerations.map((t, idx) => ({ idx, ...t })),
    { idx: tolerations.length }, // empty "add" row
  ];

  const onDeleteToleration = (idx) => () => {
    const updated = [...tolerations];
    updated.splice(idx, 1);
    onChangeHandler(updated);
  };

  const getRowProps = (item) => {
    const { idx } = item;
    const isInvalid = !!errors[idx];
    return {
      className: isInvalid ? "euiTableRow--isInvalid" : "",
      "data-test-subj": `toleration-row-${idx}`,
    };
  };

  const columns = [
    {
      name: "Key",
      field: "key",
      width: "22%",
      render: (key, item) => (
        <EuiFieldText
          controlOnly
          className="inlineTableInput"
          placeholder="dedicated"
          value={key || ""}
          onChange={(e) => onChange(`${item.idx}.key`)(e.target.value)}
        />
      ),
    },
    {
      name: "Operator",
      field: "operator",
      width: "16%",
      render: (operator, item) => (
        <EuiSelect
          compressed
          options={OPERATOR_OPTIONS}
          value={operator || ""}
          onChange={(e) => {
            const op = e.target.value;
            // When switching to Exists, clear the value field
            if (op === "Exists") {
              onChange(`${item.idx}.value`)("");
            }
            onChange(`${item.idx}.operator`)(op);
          }}
        />
      ),
    },
    {
      name: "Value",
      field: "value",
      width: "22%",
      render: (value, item) => {
        const isExists = item.operator === "Exists";
        return (
          <EuiFieldText
            controlOnly
            className="inlineTableInput"
            placeholder="ml-team"
            value={isExists ? "" : (value || "")}
            disabled={isExists}
            onChange={(e) => onChange(`${item.idx}.value`)(e.target.value)}
          />
        );
      },
    },
    {
      name: "Effect",
      field: "effect",
      width: "26%",
      render: (effect, item) => (
        <EuiSelect
          compressed
          options={EFFECT_OPTIONS}
          value={effect || ""}
          onChange={(e) => onChange(`${item.idx}.effect`)(e.target.value)}
        />
      ),
    },
    {
      width: "14%",
      actions: [
        {
          render: (item) =>
            item.idx < items.length - 1 ? (
              <EuiButtonIcon
                size="s"
                color="danger"
                iconType="trash"
                onClick={onDeleteToleration(item.idx)}
                aria-label="Remove toleration"
              />
            ) : (
              <div />
            ),
        },
      ],
    },
  ];

  return (
    <EuiDescribedFormGroup
      title={<p>Node Tolerations</p>}
      description={
        <Fragment>
          <EuiText size="s">
            Allow pods to be scheduled on nodes with matching taints. Leave
            the last blank row empty to add a new entry.
          </EuiText>
        </Fragment>
      }
      fullWidth
    >
      <EuiSpacer size="s" />
      <EuiFlexGroup direction="column" gutterSize="none">
        <EuiFlexItem>
          <InMemoryTableForm
            columns={columns}
            rowProps={getRowProps}
            items={items}
            errors={errors}
            renderErrorHeader={(key) => `Row ${parseInt(key) + 1}`}
          />
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiDescribedFormGroup>
  );
};

