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
import {
  EuiAccordion,
  EuiButton,
  EuiCodeBlock,
  EuiFlexGroup,
  EuiFlexItem,
  EuiText,
  EuiSpacer
} from "@elastic/eui";
import { FeastTransformationPanel } from "./FeastTransformationPanel";
import {
  Config,
  FeastConfig,
  newConfig
} from "../../../services/transformer/TransformerConfig";
import { feastEndpoints, useFeastApi } from "../../../hooks/useFeastApi";

const yaml = require("js-yaml");

export const StandardTransformerForm = ({ transformer, onChange }) => {
  const [config, setConfig] = useState({});

  const [configInitialized, setConfigInitialized] = useState(false);
  useEffect(() => {
    if (configInitialized) {
      return;
    }

    if (transformer.env_vars && transformer.env_vars.length > 0) {
      const envVar = transformer.env_vars.find(
        e => e.name === "TRANSFORMER_CONFIG"
      );
      if (envVar && envVar.value) {
        const envVarJSON = JSON.parse(envVar.value);
        if (envVarJSON) {
          const tc = Config.from(envVarJSON);
          setConfig(tc);
          setConfigInitialized(true);
          return;
        }
      }
    }

    const tc = newConfig();
    setConfig(tc);
    setConfigInitialized(true);
    onChange([{ name: "TRANSFORMER_CONFIG", value: JSON.stringify(tc) }]);
  }, [configInitialized, transformer.env_vars]);

  useEffect(
    () => {
      if (
        config.transformerConfig &&
        config.transformerConfig.feast &&
        transformer.env_vars
      ) {
        const newConfigJSON = JSON.stringify(config);
        // Find the index of env_var that contains transformer config
        // If it's not exist, create new env var
        // If it's exist, update it
        const idx = transformer.env_vars.findIndex(
          e => e.name === "TRANSFORMER_CONFIG"
        );
        if (idx === -1) {
          const newEnvVars = [
            { name: "TRANSFORMER_CONFIG", value: newConfigJSON }
          ];
          onChange(newEnvVars);
        } else {
          if (transformer.env_vars[idx].value !== newConfigJSON) {
            let newEnvVars = [...transformer.env_vars];
            newEnvVars[idx] = {
              ...newEnvVars[idx],
              value: newConfigJSON
            };
            onChange(newEnvVars);
          }
        }
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [config]
  );

  const onFeastChange = index => value => {
    if (config.transformerConfig) {
      let items = [...config.transformerConfig.feast];
      items[index] = value;
      setConfig({
        ...config,
        transformerConfig: {
          ...config.transformerConfig,
          feast: items
        }
      });
    }
  };

  const onAddFeastConfig = () => {
    if (config.transformerConfig) {
      const items = [
        ...config.transformerConfig.feast,
        new FeastConfig("", [], [])
      ];
      setConfig({
        ...config,
        transformerConfig: {
          ...config.transformerConfig,
          feast: items
        }
      });
    }
  };

  const onDeleteFeastConfig = idx => () => {
    if (config.transformerConfig) {
      const items = config.transformerConfig.feast.filter((value, index) => {
        return index !== idx;
      });
      setConfig({
        ...config,
        transformerConfig: {
          ...config.transformerConfig,
          feast: items
        }
      });
    }
  };

  const [{ data: feastProjects }] = useFeastApi(
    feastEndpoints.listProjects,
    { method: "POST", muteError: true },
    {},
    true
  );

  const [{ data: feastEntities }] = useFeastApi(
    feastEndpoints.listEntities,
    { method: "POST", muteError: true },
    {},
    true
  );

  const [{ data: feastFeatureTables }] = useFeastApi(
    feastEndpoints.listFeatureTables,
    { method: "POST", muteError: true },
    {},
    true
  );

  return (
    <>
      <EuiFlexGroup direction="column" gutterSize="s">
        {config.transformerConfig &&
          config.transformerConfig.feast.map((feastConfig, idx) => (
            <EuiFlexItem key={`feast-transform-${idx}`}>
              <>
                <FeastTransformationPanel
                  index={idx}
                  feastConfig={feastConfig}
                  feastProjects={feastProjects}
                  feastEntities={feastEntities}
                  feastFeatureTables={feastFeatureTables}
                  onChange={onFeastChange(idx)}
                  onDelete={onDeleteFeastConfig(idx)}
                />
                <EuiSpacer size="s" />
              </>
            </EuiFlexItem>
          ))}

        <EuiFlexItem>
          <EuiButton size="s">
            <EuiText size="s" onClick={onAddFeastConfig}>
              + Add Request Enrichment
            </EuiText>
          </EuiButton>
        </EuiFlexItem>
      </EuiFlexGroup>
      <EuiSpacer size="l" />

      {/* <EuiAccordion id="model-request-payload" buttonContent="See Request Payload to Model">
      </EuiAccordion>
      <EuiSpacer size="l" /> */}

      <EuiAccordion
        id="transformer-config-yaml"
        buttonContent="See YAML configuration">
        <EuiCodeBlock language="yaml" fontSize="m" paddingSize="m" isCopyable>
          {yaml.dump(config)}
        </EuiCodeBlock>
      </EuiAccordion>
    </>
  );
};

StandardTransformerForm.propTypes = {
  transformer: PropTypes.object,
  onChange: PropTypes.func
};
