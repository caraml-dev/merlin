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

import React, { useEffect, useState } from "react";
import PropTypes from "prop-types";
import { EuiCodeBlock, EuiFlexGroup, EuiFlexItem, EuiText } from "@elastic/eui";
import {
  ConfigSection,
  ConfigSectionPanel,
  ConfigSectionPanelTitle
} from "../../components/section";
import { ContainerConfigTable } from "../../components/ContainerConfigTable";
import { EnvVarsConfigTable } from "../../components/EnvVarsConfigTable";
import { ResourcesConfigTable } from "../../components/ResourcesConfigTable";
import { Config } from "../../services/transformer/TransformerConfig";

const yaml = require("js-yaml");

const isCustomTransformer = transformer => {
  return (
    transformer.transformer_type === undefined ||
    transformer.transformer_type === "" ||
    transformer.transformer_type === "custom"
  );
};

export const TransformerServicePanel = ({ endpoint }) => {
  const [standardTransformerConfig, setStandardTransformerConfig] = useState();
  useEffect(() => {
    if (
      endpoint.transformer.env_vars &&
      endpoint.transformer.env_vars.length > 0
    ) {
      const envVar = endpoint.transformer.env_vars.find(
        e => e.name === "TRANSFORMER_CONFIG"
      );
      if (envVar && envVar.value) {
        const envVarJSON = JSON.parse(envVar.value);
        if (envVarJSON) {
          const tc = Config.from(envVarJSON);
          setStandardTransformerConfig(tc);
          return;
        }
      }
    }
  }, [endpoint.transformer]);

  return (
    <ConfigSection title="Transformer Service">
      <EuiFlexGroup direction="row" wrap>
        <EuiFlexItem grow={3}>
          <ConfigSectionPanel>
            <EuiFlexGroup direction="column" gutterSize="m">
              <EuiFlexItem>
                <ConfigSectionPanelTitle title="Container" />
                <ContainerConfigTable config={endpoint.transformer} />
              </EuiFlexItem>

              {!isCustomTransformer(endpoint.transformer) &&
                standardTransformerConfig && (
                  <EuiFlexItem>
                    <ConfigSectionPanelTitle title="Standard Transformer Configuration" />
                    <EuiCodeBlock
                      language="yaml"
                      fontSize="m"
                      paddingSize="m"
                      isCopyable>
                      {yaml.dump(standardTransformerConfig)}
                    </EuiCodeBlock>
                  </EuiFlexItem>
                )}

              {isCustomTransformer(endpoint.transformer) &&
                endpoint.transformer.env_vars && (
                  <EuiFlexItem>
                    <ConfigSectionPanelTitle title="Environment Variables" />
                    <EnvVarsConfigTable
                      variables={endpoint.transformer.env_vars}
                    />
                  </EuiFlexItem>
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
  );
};

TransformerServicePanel.propTypes = {
  endpoint: PropTypes.object.isRequired
};
