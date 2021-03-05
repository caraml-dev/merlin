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
  newConfig,
  STANDARD_TRANSFORMER_CONFIG_ENV_NAME
} from "../../../services/transformer/TransformerConfig";
import { feastEndpoints, useFeastApi } from "../../../hooks/useFeastApi";

const yaml = require("js-yaml");

export const StandardTransformerForm = ({ transformer, onChange }) => {
  const [config, setConfig] = useState({});

  const [configInitialized, setConfigInitialized] = useState(false);
  useEffect(
    () => {
      if (configInitialized) {
        return;
      }

      // If transformer already has environment variables, check whether it already has standard transformer config or not
      if (transformer.env_vars && transformer.env_vars.length > 0) {
        const envVar = transformer.env_vars.find(
          e => e.name === STANDARD_TRANSFORMER_CONFIG_ENV_NAME
        );

        // If standard transformer config already exists, parse that value and set it to config
        if (envVar && envVar.value) {
          const envVarJSON = transformToConfig(JSON.parse(envVar.value));
          if (envVarJSON) {
            const tc = Config.from(envVarJSON);
            setConfig(tc);
            setConfigInitialized(true);
          }
        } else {
          // If there's no standard transformer config in environment variables, append the env_vars with empty transformer config.
          // This case is likely because of the user switch from custom to standard transformer
          // Note that we still keep the all previous environment variables if we switching transformer type
          const tc = newConfig();
          setConfig(tc);
          setConfigInitialized(true);
          onChange([
            ...transformer.env_vars,
            {
              name: STANDARD_TRANSFORMER_CONFIG_ENV_NAME,
              value: JSON.stringify(tc)
            }
          ]);
        }
      } else {
        // If there's no environment variables at all, create a new one.
        const tc = newConfig();
        setConfig(tc);
        setConfigInitialized(true);
        onChange([
          {
            name: STANDARD_TRANSFORMER_CONFIG_ENV_NAME,
            value: JSON.stringify(tc)
          }
        ]);
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [configInitialized, transformer.env_vars]
  );

  useEffect(
    () => {
      if (
        config.transformerConfig &&
        config.transformerConfig.feast &&
        transformer.env_vars
      ) {
        const newConfigJSON = JSON.stringify(transformToConfigJSON(config));
        // Find the index of env_var that contains transformer config
        // If it's not exist, create new env var
        // If it's exist, update it
        const idx = transformer.env_vars.findIndex(
          e => e.name === STANDARD_TRANSFORMER_CONFIG_ENV_NAME
        );
        if (idx === -1) {
          const newEnvVars = [
            { name: STANDARD_TRANSFORMER_CONFIG_ENV_NAME, value: newConfigJSON }
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
              + Add Retrieval Table
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
        buttonContent={<EuiText size="xs">See YAML configuration</EuiText>}
        paddingSize="m">
        <EuiCodeBlock language="yaml" fontSize="m" paddingSize="m" isCopyable>
          {yaml.dump(transformToConfigJSON(config))}
        </EuiCodeBlock>
      </EuiAccordion>
    </>
  );
};

const transformToConfigJSON = config => {
  let output = JSON.parse(JSON.stringify(config));
  if (output.transformerConfig) {
    output.transformerConfig.feast.forEach(feastConfig => {
      feastConfig.entities.forEach(entity => {
        if (entity.fieldType === "UDF") {
          entity["udf"] = entity.field;
        } else {
          entity["jsonPath"] = entity.field;
        }
        delete entity["field"];
        delete entity["fieldType"];
      });
    });
  }
  return output;
};

const transformToConfig = config => {
  if (config.transformerConfig) {
    config.transformerConfig.feast.forEach(feastConfig => {
      feastConfig.entities.forEach(entity => {
        if (entity.udf) {
          entity["fieldType"] = "UDF";
          entity["field"] = entity.udf;
        } else {
          entity["fieldType"] = "JSONPath";
          entity["field"] = entity.jsonPath;
        }
      });
    });
  }
  return config;
};

StandardTransformerForm.propTypes = {
  transformer: PropTypes.object,
  onChange: PropTypes.func
};
