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
import {
  EuiEmptyPrompt,
  EuiFlexGroup,
  EuiFlexItem,
  EuiText
} from "@elastic/eui";
import {
  ConfigSection,
  ConfigSectionPanel,
  ConfigSectionPanelTitle
} from "../../components/section";
import { ContainerConfigTable } from "../../components/ContainerConfigTable";
import { EnvVarsConfigTable } from "../../components/EnvVarsConfigTable";
import { ResourcesConfigTable } from "../../components/ResourcesConfigTable";

export const VersionConfig = ({ model, version, endpoint }) => {
  return endpoint &&
    (endpoint.status === "serving" || endpoint.status === "running") ? (
    <Fragment>
      <EuiFlexGroup direction="column">
        <EuiFlexItem>
          <ConfigSection title="Model Service">
            <EuiFlexGroup direction="row" wrap>
              <EuiFlexItem grow={3}>
                <ConfigSectionPanel>
                  <EuiFlexGroup direction="column" gutterSize="m">
                    <EuiFlexItem>
                      <ConfigSectionPanelTitle title="Environment Variables" />
                      {endpoint.env_vars ? (
                        <EnvVarsConfigTable variables={endpoint.env_vars} />
                      ) : (
                        <EuiText>Not available</EuiText>
                      )}
                    </EuiFlexItem>
                  </EuiFlexGroup>
                </ConfigSectionPanel>
              </EuiFlexItem>

              <EuiFlexItem grow={1} className="euiFlexItem--smallPanel">
                <ConfigSectionPanel title="Model Resources">
                  {endpoint.resource_request ? (
                    <ResourcesConfigTable
                      resourceRequest={endpoint.resource_request}
                    />
                  ) : (
                    <EuiText>Not available</EuiText>
                  )}
                </ConfigSectionPanel>
              </EuiFlexItem>
            </EuiFlexGroup>
          </ConfigSection>
        </EuiFlexItem>

        {endpoint.transformer && endpoint.transformer.enabled && (
          <EuiFlexItem>
            <ConfigSection title="Transformer Service">
              <EuiFlexGroup direction="row" wrap>
                <EuiFlexItem grow={3}>
                  <ConfigSectionPanel>
                    <EuiFlexGroup direction="column" gutterSize="m">
                      <EuiFlexItem>
                        <ConfigSectionPanelTitle title="Container" />
                        <ContainerConfigTable config={endpoint.transformer} />
                      </EuiFlexItem>

                      {endpoint.transformer.env_vars ? (
                        <EuiFlexItem>
                          <ConfigSectionPanelTitle title="Environment Variables" />
                          <EnvVarsConfigTable
                            variables={endpoint.transformer.env_vars}
                          />
                        </EuiFlexItem>
                      ) : (
                        <EuiText>Not available</EuiText>
                      )}
                    </EuiFlexGroup>
                  </ConfigSectionPanel>
                </EuiFlexItem>

                <EuiFlexItem grow={1} className="euiFlexItem--smallPanel">
                  <ConfigSectionPanel title="Transformer Resources">
                    {endpoint.resource_request ? (
                      <ResourcesConfigTable
                        resourceRequest={endpoint.transformer.resource_request}
                      />
                    ) : (
                      <EuiText>Not available</EuiText>
                    )}
                  </ConfigSectionPanel>
                </EuiFlexItem>
              </EuiFlexGroup>
            </ConfigSection>
          </EuiFlexItem>
        )}
      </EuiFlexGroup>
    </Fragment>
  ) : (
    <EuiEmptyPrompt
      title={<h2>Model version is not deployed</h2>}
      body={
        <p>
          Deploy it first and wait for deployment to complete before you can see
          the configuration details here.
        </p>
      }
    />
  );
};

VersionConfig.propTypes = {
  model: PropTypes.object.isRequired,
  version: PropTypes.object.isRequired,
  endpoint: PropTypes.object.isRequired
};
