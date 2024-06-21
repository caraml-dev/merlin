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
