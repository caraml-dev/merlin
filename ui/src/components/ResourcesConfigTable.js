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
    liveness_probe_initial_delay_seconds,
    liveness_probe_period_seconds,
    liveness_probe_timeout_seconds,
    liveness_probe_success_threshold,
    liveness_probe_failure_threshold,
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

  // Add liveness probe configuration if any value is set
  if (liveness_probe_initial_delay_seconds !== undefined && liveness_probe_initial_delay_seconds !== null) {
    items.push({
      title: "Liveness Initial Delay",
      description: `${liveness_probe_initial_delay_seconds}s`,
    });
  }

  if (liveness_probe_period_seconds !== undefined && liveness_probe_period_seconds !== null) {
    items.push({
      title: "Liveness Period",
      description: `${liveness_probe_period_seconds}s`,
    });
  }

  if (liveness_probe_timeout_seconds !== undefined && liveness_probe_timeout_seconds !== null) {
    items.push({
      title: "Liveness Timeout",
      description: `${liveness_probe_timeout_seconds}s`,
    });
  }

  if (liveness_probe_success_threshold !== undefined && liveness_probe_success_threshold !== null) {
    items.push({
      title: "Liveness Success Threshold",
      description: liveness_probe_success_threshold,
    });
  }

  if (liveness_probe_failure_threshold !== undefined && liveness_probe_failure_threshold !== null) {
    items.push({
      title: "Liveness Failure Threshold",
      description: liveness_probe_failure_threshold,
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
