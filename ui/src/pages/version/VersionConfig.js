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
import { Link } from "@reach/router";
import {
  EuiButton,
  EuiButtonEmpty,
  EuiEmptyPrompt,
  EuiFlexGroup,
  EuiFlexItem,
  EuiLoadingSpinner,
  EuiSpacer,
  EuiText
} from "@elastic/eui";
import config from "../../config";
import { ModelServicePanel } from "./ModelServicePanel";
import { TransformerServicePanel } from "./TransformerServicePanel";

export const VersionConfig = ({ model, version, endpoint }) => {
  return endpoint &&
    (endpoint.status === "serving" || endpoint.status === "running") ? (
    <Fragment>
      <EuiFlexGroup direction="column">
        <EuiFlexItem>
          <ModelServicePanel endpoint={endpoint} />
        </EuiFlexItem>

        {endpoint.transformer && endpoint.transformer.enabled && (
          <EuiFlexItem>
            <TransformerServicePanel endpoint={endpoint} />
          </EuiFlexItem>
        )}
      </EuiFlexGroup>
    </Fragment>
  ) : (
    <EuiEmptyPrompt
      title={
        <h2>
          {endpoint.status === "pending"
            ? "Deploying model version"
            : "Model version is not deployed"}
        </h2>
      }
      body={
        endpoint.status === "pending" ? (
          <Fragment>
            <EuiLoadingSpinner size="xl" />
            <EuiSpacer size="s" />
            <EuiButtonEmpty size="xs" onClick={() => window.location.reload()}>
              <EuiText size="xs">Reload page</EuiText>
            </EuiButtonEmpty>
          </Fragment>
        ) : (
          <Fragment>
            <p>
              Deploy it first and wait for deployment to complete before you can
              see the configuration details here.
            </p>
            <Link
              to={`${config.HOMEPAGE}/projects/${model.project_id}/models/${model.id}/versions/${version.id}/endpoints/${endpoint.id}/redeploy`}
              state={{ model: model, version: version }}>
              <EuiButton iconType="importAction" size="xs">
                <EuiText size="xs">
                  {model.type !== "pyfunc_v2"
                    ? "Redeploy"
                    : "Redeploy Endpoint"}
                </EuiText>
              </EuiButton>
            </Link>
          </Fragment>
        )
      }
    />
  );
};

VersionConfig.propTypes = {
  model: PropTypes.object.isRequired,
  version: PropTypes.object.isRequired,
  endpoint: PropTypes.object.isRequired
};
