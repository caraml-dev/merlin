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
import { EuiFlexGroup, EuiFlexItem, EuiText } from "@elastic/eui";
import {
  ConfigSection,
  ConfigSectionPanel,
  ConfigSectionPanelTitle
} from "../../components/section";
import { EnvVarsConfigTable } from "../../components/EnvVarsConfigTable";
import { SecretsConfigTable } from "../../components/SecretsConfigTable";
import { ResourcesConfigTable } from "../../components/ResourcesConfigTable";
import { ContainerConfigTable } from "../../components/ContainerConfigTable";

export const ModelServicePanel = ({ endpoint, version }) => {
  return (
    <ConfigSection title="Model Service">
      <EuiFlexGroup direction="row" wrap>
        <EuiFlexItem grow={3}>
          <ConfigSectionPanel>
            <EuiFlexGroup direction="column" gutterSize="m">
              {version.custom_predictor && (
                <EuiFlexItem>
                  <ConfigSectionPanelTitle title="Container" />
                  <ContainerConfigTable config={version.custom_predictor} />
                </EuiFlexItem>
              )}
              <EuiFlexItem>
                <ConfigSectionPanelTitle title="Environment Variables" />
                {endpoint.env_vars ? (
                  <EnvVarsConfigTable variables={endpoint.env_vars} />
                ) : (
                  <EuiText>Not available</EuiText>
                )}
              </EuiFlexItem>
              <EuiFlexItem>
                <ConfigSectionPanelTitle title="Secrets" />
                {endpoint.secrets ? (
                  <SecretsConfigTable variables={endpoint.secrets} />
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
  );
};

ModelServicePanel.propTypes = {
  endpoint: PropTypes.object.isRequired
};
