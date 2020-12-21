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

import React, { Fragment, useEffect, useState } from "react";
import {
  EuiButtonEmpty,
  EuiFlexGroup,
  EuiFlexItem,
  EuiSpacer,
  EuiSwitch,
  EuiText
} from "@elastic/eui";
import { ModelAlertRulePercentileRow } from "./ModelAlertRulePercentileRow";
import PropTypes from "prop-types";

export const ModelAlertRulePercentile = ({
  title,
  name,
  description,
  unit,
  request,
  setRequest
}) => {
  const [isAddAlertDisabled, setAddAlertDisabled] = useState(true);

  const [ruleRowsInitialized, setRuleRowsInitialized] = useState(false);
  const [ruleRows, setRuleRows] = useState([{}]);
  const maxRuleRows = 4;

  useEffect(() => {
    ruleRows.length === maxRuleRows
      ? setAddAlertDisabled(false)
      : setAddAlertDisabled(true);

    const severities = ["WARNING", "CRITICAL"];
    const conditions = request.alert_conditions;

    ruleRows
      .filter(ruleRow => ruleRow.percentile)
      .forEach(ruleRow => {
        severities.forEach(severity => {
          if (ruleRow[severity]) {
            const index = conditions
              ? conditions.findIndex(
                  c =>
                    c.metric_type === name &&
                    c.percentile === parseFloat(ruleRow.percentile) &&
                    c.severity === severity
                )
              : -1;

            if (index === -1) {
              conditions.push({
                enabled: true,
                metric_type: name,
                severity: severity,
                target: ruleRow[severity],
                unit: unit ? unit : "",
                percentile: parseFloat(ruleRow.percentile)
              });
            } else {
              conditions[index] = {
                ...conditions[index],
                target: ruleRow[severity]
              };
            }
          }
        });

        setRequest(r => ({
          ...r,
          alert_conditions: conditions
        }));
      });
  }, [ruleRows, name, unit, request.alert_conditions, setRequest]);

  useEffect(() => {
    if (request.alert_conditions && !ruleRowsInitialized) {
      const conditions = request.alert_conditions.filter(
        c => c.metric_type === name
      );
      if (conditions.length > 0) {
        const ruleRowMappings = {};
        conditions.forEach(condition => {
          if (condition.percentile in ruleRowMappings) {
            ruleRowMappings[condition.percentile] = {
              ...ruleRowMappings[condition.percentile],
              [condition.severity]: condition.target
            };
          } else {
            ruleRowMappings[condition.percentile] = {
              percentile: condition.percentile.toString(),
              [condition.severity]: condition.target
            };
          }
        });

        const ruleRows = [];
        for (var percentile in ruleRowMappings) {
          ruleRows.push(ruleRowMappings[percentile]);
        }
        setRuleRows(
          ruleRows.sort((a, b) => (a.percentile < b.percentile ? 1 : -1))
        );
        setRuleRowsInitialized(true);
      }
    }
  }, [request, name, ruleRowsInitialized, setRuleRowsInitialized, setRuleRows]);

  const addRuleRow = () => {
    if (ruleRows.length < maxRuleRows) {
      setRuleRows([...ruleRows, {}]);
    }
  };

  const editRuleRow = idx => {
    return ruleRow => {
      ruleRows[idx] = ruleRow;
      setRuleRows([...ruleRows]);
    };
  };

  const deleteRuleRow = idx => {
    ruleRows.splice(idx, 1);
    setRuleRows([...ruleRows]);
  };

  const [isChecked, setChecked] = useState(false);

  useEffect(() => {
    if (request && request.alert_conditions) {
      const condition = request.alert_conditions.find(
        c => c.metric_type === name
      );
      condition &&
        condition.enabled &&
        setChecked(condition.enabled) &&
        setAddAlertDisabled(!condition.enabled);
    }
  }, [request, name, setChecked]);

  const setConditionEnabled = (metricType, enabled) => {
    setChecked(enabled);

    const conditions = request.alert_conditions;
    conditions.forEach((condition, index) => {
      if (conditions[index].metric_type === metricType) {
        conditions[index] = {
          ...conditions[index],
          enabled: enabled
        };
      }
    });
    setRequest({
      ...request,
      alert_conditions: conditions
    });
  };

  const onSwitchChange = (metricType, enabled) => {
    setConditionEnabled(metricType, enabled);
  };

  return (
    <Fragment>
      <EuiSpacer size="l" />
      <EuiFlexGroup alignItems="center" justifyContent="spaceBetween">
        <EuiFlexItem grow={false}>
          <EuiText size="s">
            <h4>{title}</h4>
          </EuiText>
          <EuiText size="xs">{description}</EuiText>
        </EuiFlexItem>
        <EuiFlexItem grow={false}>
          <EuiSwitch
            label=""
            checked={isChecked}
            onChange={e => onSwitchChange(name, e.target.checked)}
          />
        </EuiFlexItem>
      </EuiFlexGroup>

      <Fragment>
        {ruleRows.map((row, idx) => (
          <ModelAlertRulePercentileRow
            key={`model-alert-rule-pecentile-${idx}`}
            ruleRow={row}
            editRuleRow={editRuleRow(idx)}
            deleteRuleRow={() => deleteRuleRow(idx)}
            unit={unit}
            isChecked={isChecked}
          />
        ))}

        <EuiFlexGroup>
          <EuiFlexItem grow={false}>
            <EuiButtonEmpty
              size="xs"
              iconType="plusInCircle"
              disabled={!isChecked || !isAddAlertDisabled}
              onClick={addRuleRow}>
              Add more latency alert rule
            </EuiButtonEmpty>
          </EuiFlexItem>
        </EuiFlexGroup>
      </Fragment>
    </Fragment>
  );
};

ModelAlertRulePercentile.propTypes = {
  title: PropTypes.string,
  name: PropTypes.string,
  description: PropTypes.string,
  unit: PropTypes.string,
  request: PropTypes.object,
  setRequest: PropTypes.func
};
