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
import { EuiIcon, EuiTitle } from "@elastic/eui";

import { appConfig } from "../config";

export const PageTitle = ({ title, size, iconSize }) => (
  <EuiTitle size={size || "l"}>
    <span>
      <EuiIcon type={appConfig.appIcon} size={iconSize || "xl"} />
      &nbsp;{title}
    </span>
  </EuiTitle>
);

PageTitle.propTypes = {
  title: PropTypes.object.isRequired,
  size: PropTypes.oneOf(["xs", "s", "m", "l", "xl"]),
  iconSize: PropTypes.oneOf(["xs", "s", "m", "l", "xl"])
};
