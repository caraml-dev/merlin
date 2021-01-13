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

import React, { useEffect } from "react";
import { EuiFlexGroup, EuiFlexItem, EuiPanel, EuiSpacer } from "@elastic/eui";
import { LoggerForm } from "./LoggerForm";
import PropTypes from "prop-types";

export const Logger = ({ logger, transformer, onChange }) => {
  const setValue = (field, value) =>
    onChange({
      ...logger,
      [field]: value
    });

  useEffect(() => {
    if (!logger.transformer) {
      return;
    }
    if (!transformer) {
      logger.transformer.enabled = false;
    }
  }, [transformer, logger.transformer]);

  const modelConfig = logger.model || { enabled: false, mode: "all" };
  const transformerConfig = logger.transformer || {
    enabled: false,
    mode: "all"
  };
  return (
    <EuiFlexGroup direction="column">
      <EuiFlexItem grow={false}>
        <EuiPanel grow={false}>
          <EuiSpacer size="s" />
          <LoggerForm
            name="Model Logger"
            configuration={modelConfig}
            onChange={value => setValue("model", value)}
          />
        </EuiPanel>
      </EuiFlexItem>

      {transformer && (
        <EuiFlexItem grow={false}>
          <EuiPanel grow={false}>
            <EuiSpacer size="s" />
            <LoggerForm
              name="Transformer Logger"
              configuration={transformerConfig}
              onChange={value => setValue("transformer", value)}
            />
          </EuiPanel>
        </EuiFlexItem>
      )}
    </EuiFlexGroup>
  );
};

Logger.propTypes = {
  logger: PropTypes.object,
  transformer: PropTypes.bool,
  onChange: PropTypes.func
};
