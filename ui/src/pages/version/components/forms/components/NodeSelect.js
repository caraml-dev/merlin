import React, { Fragment, useState } from "react";
import { EuiSuperSelect, EuiText } from "@elastic/eui";
import { useMerlinApi } from "../../../../../hooks/useMerlinApi";

const DEFAULT_VALUE = "__default__";
const HOSTNAME_LABEL = "kubernetes.io/hostname";

const taintSummary = (taints) =>
  taints && taints.length > 0
    ? taints
        .map((t) => `${t.key}=${t.value || ""}:${t.effect || ""}`)
        .join(", ")
    : "no taints";

/**
 * NodeSelect - a dropdown to pin a model onto a specific node.
 *
 * Selecting a node sets the node selector to that node's hostname (pin) and adds
 * a toleration for each of the node's taints (permit), so the pod can land there.
 *
 * Props
 *   environment          – environment name (used to fetch that cluster's nodes)
 *   onSelect(sel, tols)  – called with the new node_selector + tolerations
 */
export const NodeSelect = ({ environment, onSelect }) => {
  const [selected, setSelected] = useState(DEFAULT_VALUE);

  const [{ data: nodes, isLoaded }] = useMerlinApi(
    `/environments/${environment}/nodes`,
    {},
    [],
    !!environment
  );

  const options = [
    {
      value: DEFAULT_VALUE,
      inputDisplay: "Default — any available node",
    },
    ...(nodes || []).map((node) => ({
      value: node.name,
      inputDisplay: node.name,
      disabled: !node.ready,
      dropdownDisplay: (
        <Fragment>
          <strong>
            {node.name}
            {!node.ready ? " (not ready)" : ""}
          </strong>
          <EuiText size="s" color="subdued">
            <p>{taintSummary(node.taints)}</p>
          </EuiText>
        </Fragment>
      ),
    })),
  ];

  const onChange = (value) => {
    setSelected(value);
    if (value === DEFAULT_VALUE) {
      onSelect({}, []);
      return;
    }
    const node = (nodes || []).find((n) => n.name === value);
    if (!node) return;

    const nodeSelector = { [HOSTNAME_LABEL]: node.name };
    const tolerations = (node.taints || []).map((t) => ({
      key: t.key,
      operator: t.value ? "Equal" : "Exists",
      ...(t.value ? { value: t.value } : {}),
      ...(t.effect ? { effect: t.effect } : {}),
    }));
    onSelect(nodeSelector, tolerations);
  };

  return (
    <EuiSuperSelect
      fullWidth
      options={options}
      valueOfSelected={selected}
      onChange={onChange}
      isLoading={!!environment && !isLoaded}
      hasDividers
    />
  );
};
