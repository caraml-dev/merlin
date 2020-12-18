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
import { EuiHealth, EuiText } from "@elastic/eui";

export const DeploymentStatus = ({ status, size }) => {
  const healthColor = status => {
    switch (status) {
      case "serving":
        return "#fea27f";
      case "running":
        return "success";
      case "pending":
        return "gray";
      case "terminated":
        return "danger";
      case "failed":
        return "danger";
      default:
        return "subdued";
    }
  };

  return (
    <EuiHealth color={healthColor(status)}>
      <EuiText size={size || "s"}>{status}</EuiText>
    </EuiHealth>
  );
};

DeploymentStatus.propTypes = {
  status: PropTypes.string,
  size: PropTypes.oneOf(["xs", "s", "m", "l", "xl"])
};
