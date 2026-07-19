import React, { Fragment, useState } from "react";
import { EuiSuperSelect, EuiText } from "@elastic/eui";
import { useMerlinApi } from "../../../../../hooks/useMerlinApi";

const DEFAULT_VALUE = "__default__";

const poolId = (pool) =>
  `${pool.taint.key}=${pool.taint.value || ""}:${pool.taint.effect || ""}`;

/**
 * NodePoolSelect - a single dropdown that pins a model to a discovered node pool.
 *
 * Selecting a pool sets the node selector (to pin) and adds the pool's matching
 * toleration (to permit), so the user never types raw labels or taints.
 *
 * Props
 *   environment          – environment name (used to fetch pools for its cluster)
 *   nodeSelector         – current node_selector object
 *   tolerations          – current tolerations array
 *   onSelect(sel, tols)  – called with the new node_selector + tolerations
 */
export const NodePoolSelect = ({
  environment,
  nodeSelector = {},
  tolerations = [],
  onSelect,
}) => {
  const [selected, setSelected] = useState(DEFAULT_VALUE);

  const [{ data: pools, isLoaded }] = useMerlinApi(
    `/environments/${environment}/node-pools`,
    {},
    [],
    !!environment
  );

  const options = [
    {
      value: DEFAULT_VALUE,
      inputDisplay: "Default — any available node",
    },
    ...(pools || []).map((pool) => ({
      value: poolId(pool),
      inputDisplay: `${pool.taint.key}=${pool.taint.value || "\"\""}`,
      dropdownDisplay: (
        <Fragment>
          <strong>{`${pool.taint.key}=${pool.taint.value || "\"\""}`}</strong>
          <EuiText size="s" color="subdued">
            <p>{`taint ${poolId(pool)} · selector ${JSON.stringify(pool.node_selector || {})}`}</p>
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
    const pool = (pools || []).find((p) => poolId(p) === value);
    if (!pool) return;
    const toleration = {
      key: pool.taint.key,
      operator: "Equal",
      ...(pool.taint.value ? { value: pool.taint.value } : {}),
      ...(pool.taint.effect ? { effect: pool.taint.effect } : {}),
    };
    onSelect({ ...(pool.node_selector || {}) }, [toleration]);
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
