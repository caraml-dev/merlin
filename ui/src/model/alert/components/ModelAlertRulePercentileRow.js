/**
 * Copyright 2020 The Merlin Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React from "react";
import {
  EuiButtonIcon,
  EuiFieldNumber,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiSelect
} from "@elastic/eui";

export const ModelAlertRulePercentileRow = ({
  ruleRow,
  editRuleRow,
  deleteRuleRow,
  unit,
  isChecked
}) => {
  const setThresholdValue = field => {
    return e => {
      const sanitizedValue = parseInt(e.target.value);
      !isNaN(sanitizedValue) &&
        editRuleRow({ ...ruleRow, [field]: sanitizedValue });
    };
  };

  return (
    <EuiFlexGroup alignItems="center">
      <EuiFlexItem grow={1}>
        <EuiFormRow label="Percentile">
          <EuiSelect
            hasNoInitialSelection
            disabled={!isChecked}
            value={ruleRow.percentile}
            onChange={e =>
              editRuleRow({ ...ruleRow, percentile: e.target.value })
            }
            options={[
              { value: "99", text: "99th percentile" },
              { value: "95", text: "95th percentile" },
              { value: "90", text: "90th percentile" },
              { value: "50", text: "50th percentile" }
            ]}
          />
        </EuiFormRow>
      </EuiFlexItem>
      <EuiFlexItem grow={1}>
        <EuiFormRow label="Warning threshold">
          <EuiFieldNumber
            value={ruleRow.WARNING ? ruleRow.WARNING : ""}
            onChange={setThresholdValue("WARNING")}
            append={unit}
            name="throughputWarning"
            disabled={!isChecked}
          />
        </EuiFormRow>
      </EuiFlexItem>
      <EuiFlexItem grow={1}>
        <EuiFormRow label="Critical threshold">
          <EuiFieldNumber
            value={ruleRow.CRITICAL ? ruleRow.CRITICAL : ""}
            onChange={setThresholdValue("CRITICAL")}
            append={unit}
            name="throughputCritical"
            disabled={!isChecked}
          />
        </EuiFormRow>
      </EuiFlexItem>
      <EuiFlexItem grow={false}>
        <EuiFormRow>
          {deleteRuleRow ? (
            <EuiButtonIcon
              aria-label="Remove alert percentile rule"
              color="danger"
              iconSize="m"
              iconType="cross"
              disabled={!isChecked}
              onClick={deleteRuleRow}
            />
          ) : (
            <EuiButtonIcon
              aria-label="Placeholder"
              iconSize="m"
              iconType="empty"
            />
          )}
        </EuiFormRow>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
