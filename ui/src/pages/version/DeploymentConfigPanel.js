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
import PropTypes from "prop-types";
import { EuiDescriptionList } from "@elastic/eui";
import { ConfigSection, ConfigSectionPanel } from "../../components/section";

/**
 * DeploymentConfigPanel UI Panel containing deployment configuration of an endpoint
 * @param {*} endpoint The endpoint object
 */
export const DeploymentConfigPanel = ({ endpoint }) => {
  const autoscalingMetricsDesc = {
    concurrency: "Concurrency",
    cpu_utilization: "CPU Utilization",
    memory_utilization: "Memory Utilization",
    rps: "Request Per Second",
  };

  const deploymentModeDesc = {
    raw_deployment: "Raw Deployment",
    serverless: "Serverless",
  };

  const protocolDesc = {
    UPI_V1: "Universal Prediction Interface (gRPC)",
    HTTP_JSON: "JSON over HTTP",
  };

  const deploymentConfig = [
    {
      title: "Protocol",
      description: endpoint.protocol
        ? protocolDesc[endpoint.protocol]
        : protocolDesc["HTTP_JSON"],
    },
    {
      title: "Deployment Mode",
      description: endpoint.deployment_mode
        ? deploymentModeDesc[endpoint.deployment_mode]
        : deploymentModeDesc["serverless"],
    },
    {
      title: "Autoscaling Metrics",
      description: endpoint.autoscaling_policy
        ? autoscalingMetricsDesc[endpoint.autoscaling_policy.metrics_type]
        : autoscalingMetricsDesc["concurrency"],
    },
    {
      title: "Autoscaling Target",
      description: endpoint.autoscaling_policy
        ? endpoint.autoscaling_policy.target_value
        : 1,
    },
  ];

  return (
    <ConfigSection title="Deployment Config">
      <ConfigSectionPanel>
        <EuiDescriptionList
          compressed
          type="responsiveColumn"
          listItems={deploymentConfig}
        ></EuiDescriptionList>
      </ConfigSectionPanel>
    </ConfigSection>
  );
};

DeploymentConfigPanel.propTypes = {
  endpoint: PropTypes.object.isRequired,
};
