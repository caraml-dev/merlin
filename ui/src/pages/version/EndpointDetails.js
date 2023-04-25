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

import React, { Fragment } from "react";
import PropTypes from "prop-types";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { ModelServicePanel } from "./ModelServicePanel";
import { TransformerServicePanel } from "./TransformerServicePanel";
import { DeploymentConfigPanel } from "./DeploymentConfigPanel";

/**
 * EndpointDetails UI component containing detailed information of an endpoint
 * @param {*} endpoint The endpoint
 * @param {*} version The model version from which the endpoint is created
 */
export const EndpointDetails = ({ endpoint, version }) => {
  return (
    endpoint && (
      <Fragment>
        <EuiFlexGroup direction="column">
          <EuiFlexItem>
            <DeploymentConfigPanel endpoint={endpoint} />
          </EuiFlexItem>

          <EuiFlexItem>
            <ModelServicePanel endpoint={endpoint} version={version} />
          </EuiFlexItem>

          {endpoint.transformer && endpoint.transformer.enabled && (
            <EuiFlexItem>
              <TransformerServicePanel endpoint={endpoint} />
            </EuiFlexItem>
          )}
        </EuiFlexGroup>
      </Fragment>
    )
  );
};

EndpointDetails.propTypes = {
  version: PropTypes.object.isRequired,
  endpoint: PropTypes.object.isRequired
};
