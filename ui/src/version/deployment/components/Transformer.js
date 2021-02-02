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
import {
  EuiForm,
  EuiFormRow,
  EuiPanel,
  EuiSpacer,
  EuiSuperSelect,
  EuiTitle
} from "@elastic/eui";
import { CustomTransformerForm } from "./CustomTransformerForm";
import { EnvironmentVariables } from "./EnvironmentVariables";
import { LoggerForm } from "./LoggerForm";
import { StandardTransformerForm } from "./StandardTransformerForm";

const extractTransformerType = transformer => {
  if (!transformer.enabled) {
    return "disabled";
  }
  if (
    transformer.transformer_type === undefined ||
    transformer.transformer_type === "" ||
    transformer.transformer_type === "custom"
  ) {
    return "custom";
  }
  return "standard";
};

const transformerTypes = [
  {
    value: "disabled",
    inputDisplay: "No Transformer",
    dropdownDisplay: "No Transformer"
  },
  {
    value: "standard",
    inputDisplay: "Standard Transformer",
    dropdownDisplay: "Standard Transformer"
  },
  {
    value: "custom",
    inputDisplay: "Custom Transformer",
    dropdownDisplay: "Custom Transformer"
  }
];

export const Transformer = ({
  transformer,
  onChange,
  defaultResourceRequest,
  logger,
  onLoggerChange
}) => {
  const setValue = (field, value) =>
    onChange({
      ...transformer,
      [field]: value
    });

  const transformerType = extractTransformerType(transformer);

  const onTransformerTypeChange = value => {
    if (value === "disabled") {
      setValue("enabled", false);
      return;
    }

    onChange({
      ...transformer,
      enabled: true,
      transformer_type: value
    });
  };

  const onVariablesChange = value => {
    if (JSON.stringify(value) !== JSON.stringify(transformer.env_vars)) {
      setValue("env_vars", value);
    }
  };

  return (
    <EuiPanel grow={false}>
      <EuiTitle size="xs">
        <h4>Transformer Configuration</h4>
      </EuiTitle>

      <EuiSpacer size="m" />

      <EuiForm>
        <EuiFormRow
          fullWidth
          label="Transformer Type"
          display="columnCompressed">
          <EuiSuperSelect
            fullWidth
            options={transformerTypes}
            valueOfSelected={transformerType}
            onChange={value => onTransformerTypeChange(value)}
            itemLayoutAlign="top"
            hasDividers
          />
        </EuiFormRow>

        {transformer.enabled && (
          <>
            {(transformer.transformer_type === undefined ||
              transformer.transformer_type === "" ||
              transformer.transformer_type === "custom") && (
              <CustomTransformerForm
                transformer={transformer}
                defaultResourceRequest={
                  transformer.resource_request || defaultResourceRequest
                }
                onTransformerChange={onChange}
              />
            )}

            {transformer.transformer_type === "standard" && (
              <>
                <EuiSpacer size="l" />
                <StandardTransformerForm
                  transformer={transformer}
                  onChange={onVariablesChange}
                />
              </>
            )}

            <EuiSpacer size="l" />
            <LoggerForm
              logger={logger}
              config_type="transformer"
              onChange={onLoggerChange}
            />

            {(transformer.transformer_type === undefined ||
              transformer.transformer_type === "" ||
              transformer.transformer_type === "custom") && (
              <>
                <EuiSpacer size="l" />
                <EuiFormRow fullWidth label="Environment Variables">
                  <EnvironmentVariables
                    variables={transformer.env_vars || []}
                    onChange={onVariablesChange}
                  />
                </EuiFormRow>
              </>
            )}
          </>
        )}
      </EuiForm>
    </EuiPanel>
  );
};

Transformer.propTypes = {
  transformer: PropTypes.object,
  onChange: PropTypes.func,
  defaultResourceRequest: PropTypes.object
};
