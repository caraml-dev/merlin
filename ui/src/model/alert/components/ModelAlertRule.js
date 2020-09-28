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

import React, { Fragment, useState, useEffect } from "react";
import {
  EuiFieldNumber,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiSpacer,
  EuiSwitch,
  EuiText
} from "@elastic/eui";

export const ModelAlertRule = ({
  title,
  name,
  description,
  unit,
  max,
  request,
  setRequest
}) => {
  const getConditionThreshold = (metricType, severity) => {
    if (request && request.alert_conditions) {
      const condition = request.alert_conditions.find(
        c => c.metric_type === metricType && c.severity === severity
      );
      return condition ? condition.target : "";
    } else {
      return "";
    }
  };

  const setConditionThreshold = (metricType, severity) => {
    return e => {
      const sanitizedValue = parseInt(e.target.value);
      if (!isNaN(sanitizedValue)) {
        const conditions = request.alert_conditions;
        const index = conditions
          ? conditions.findIndex(
              c => c.metric_type === metricType && c.severity === severity
            )
          : -1;
        if (index === -1) {
          conditions.push({
            enabled: true,
            metric_type: metricType,
            severity: severity,
            target: sanitizedValue,
            unit: unit ? unit : ""
          });
        } else {
          conditions[index] = {
            ...request.alert_conditions[index],
            target: sanitizedValue
          };
        }

        setRequest({
          ...request,
          alert_conditions: conditions
        });
      }
    };
  };

  const [isChecked, setChecked] = useState(false);

  useEffect(() => {
    if (request && request.alert_conditions) {
      const condition = request.alert_conditions.find(
        c => c.metric_type === name
      );
      condition && condition.enabled && setChecked(true);
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
      <EuiFlexGroup>
        <EuiFlexItem grow={1}>
          <EuiFormRow label={"Warning threshold"}>
            <EuiFieldNumber
              value={getConditionThreshold(name, "WARNING")}
              onChange={setConditionThreshold(name, "WARNING")}
              append={unit}
              name={name + "Warning"}
              disabled={!isChecked}
              max={max}
            />
          </EuiFormRow>
        </EuiFlexItem>
        <EuiFlexItem grow={1}>
          <EuiFormRow label={"Critical threshold"}>
            <EuiFieldNumber
              value={getConditionThreshold(name, "CRITICAL")}
              onChange={setConditionThreshold(name, "CRITICAL")}
              append={unit}
              name={name + "Critical"}
              disabled={!isChecked}
              max={max}
            />
          </EuiFormRow>
        </EuiFlexItem>
      </EuiFlexGroup>
    </Fragment>
  );
};
