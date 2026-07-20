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

import { EuiDescriptionList } from "@elastic/eui";
import PropTypes from "prop-types";
import React from "react";

export const ResourcesConfigTable = ({
  resourceRequest: {
    cpu_request,
    cpu_limit,
    memory_request,
    min_replica,
    max_replica,
    gpu_name,
    gpu_request,
    liveness_probe,
    readiness_probe,
    tolerations,
    node_selector,
  },
}) => {
  const items = [
    {
      title: "CPU Request",
      description: cpu_request,
    },
    ...(cpu_limit !== undefined && cpu_limit !== "0" && cpu_limit !== "") ? [
      {
        title: "CPU Limit",
        description: cpu_limit,
      }
    ] : [],
    {
      title: "Memory Request",
      description: memory_request,
    },
    {
      title: "Min Replicas",
      description: min_replica,
    },
    {
      title: "Max Replicas",
      description: max_replica,
    },
  ];

  if (gpu_name !== undefined && gpu_name !== "") {
    items.push({
      title: "GPU Name",
      description: gpu_name,
    });
  }

  if (gpu_request !== undefined && gpu_request !== "0") {
    items.push({
      title: "GPU Request",
      description: gpu_request,
    });
  }

  // Add liveness probe info if configured
  if (liveness_probe && Object.keys(liveness_probe).some(k => liveness_probe[k])) {
    const probeDetails = [];
    if (liveness_probe.initial_delay_seconds) probeDetails.push(`delay: ${liveness_probe.initial_delay_seconds}s`);
    if (liveness_probe.timeout_seconds) probeDetails.push(`timeout: ${liveness_probe.timeout_seconds}s`);
    if (liveness_probe.period_seconds) probeDetails.push(`period: ${liveness_probe.period_seconds}s`);
    if (liveness_probe.failure_threshold) probeDetails.push(`failures: ${liveness_probe.failure_threshold}`);
    if (liveness_probe.success_threshold) probeDetails.push(`successes: ${liveness_probe.success_threshold}`);
    if (liveness_probe.path) probeDetails.push(`path: ${liveness_probe.path}`);
    if (probeDetails.length > 0) {
      items.push({
        title: "Liveness Probe",
        description: probeDetails.join(", "),
      });
    }
  }

  // Add readiness probe info if configured
  if (readiness_probe && Object.keys(readiness_probe).some(k => readiness_probe[k])) {
    const probeDetails = [];
    if (readiness_probe.initial_delay_seconds) probeDetails.push(`delay: ${readiness_probe.initial_delay_seconds}s`);
    if (readiness_probe.timeout_seconds) probeDetails.push(`timeout: ${readiness_probe.timeout_seconds}s`);
    if (readiness_probe.period_seconds) probeDetails.push(`period: ${readiness_probe.period_seconds}s`);
    if (readiness_probe.failure_threshold) probeDetails.push(`failures: ${readiness_probe.failure_threshold}`);
    if (readiness_probe.success_threshold) probeDetails.push(`successes: ${readiness_probe.success_threshold}`);
    if (readiness_probe.path) probeDetails.push(`path: ${readiness_probe.path}`);
    if (probeDetails.length > 0) {
      items.push({
        title: "Readiness Probe",
        description: probeDetails.join(", "),
      });
    }
  }

  // Add tolerations if configured
  if (tolerations && tolerations.length > 0) {
    tolerations.forEach((t, idx) => {
      const parts = [];
      if (t.key) parts.push(`key: ${t.key}`);
      if (t.operator) parts.push(`op: ${t.operator}`);
      if (t.value) parts.push(`value: ${t.value}`);
      if (t.effect) parts.push(`effect: ${t.effect}`);
      items.push({
        title: idx === 0 ? "Tolerations" : "",
        description: parts.join(", ") || "—",
      });
    });
  }

  // Add node selectors if configured
  if (node_selector && Object.keys(node_selector).length > 0) {
    Object.entries(node_selector).forEach(([key, value], idx) => {
      items.push({
        title: idx === 0 ? "Node Selector" : "",
        description: `${key}: ${value}`,
      });
    });
  }

  return (
    <EuiDescriptionList
      compressed
      type="responsiveColumn"
      listItems={items}
      columnWidths={[1, 1]}
    />
  );
};

ResourcesConfigTable.propTypes = {
  resourceRequest: PropTypes.object.isRequired,
};
