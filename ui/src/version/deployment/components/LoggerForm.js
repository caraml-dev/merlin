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
import {
  EuiForm,
  EuiFormRow,
  EuiIcon,
  EuiSpacer,
  EuiText,
  EuiSuperSelect,
  EuiSwitch,
  EuiToolTip
} from "@elastic/eui";
import PropTypes from "prop-types";

export const DEFAULT_LOGGER_CONFIG = { enabled: false, mode: "all" };

export const LoggerForm = ({ logger, config_type, onChange }) => {
  const setValue = (field, value) => {
    const configuration = getConfiguration(config_type);
    const updatedConfig = { ...configuration, [field]: value };
    onChange({
      ...logger,
      [config_type]: updatedConfig
    });
  };

  const getConfiguration = config_type =>
    logger[config_type] || DEFAULT_LOGGER_CONFIG;

  const configuration = getConfiguration(config_type);
  const enabled = configuration.enabled || false;
  const selectedMode = configuration.mode || "all";
  const loggerModes = [
    {
      name: "All",
      value: "all",
      desc: "Log request and response"
    },
    {
      name: "Request",
      value: "request",
      desc: "Only log request"
    },
    {
      name: "Response",
      value: "response",
      desc: "Only log response"
    }
  ];

  const loggerModeOptions = loggerModes.map(mode => {
    return {
      value: mode.value,
      inputDisplay: mode.name,
      dropdownDisplay: (
        <Fragment>
          <strong>{mode.name}</strong>
          <EuiText size="s" color="subdued">
            <p className="euiTextColor--subdued">{mode.desc}</p>
          </EuiText>
        </Fragment>
      )
    };
  });

  return (
    <>
      <EuiForm>
        <EuiFormRow fullWidth label="Logger enabled" display="columnCompressed">
          <>
            <EuiSpacer size="s" />
            <EuiSwitch
              label=""
              checked={enabled}
              onChange={e => setValue("enabled", e.target.checked)}></EuiSwitch>
          </>
        </EuiFormRow>

        {enabled && (
          <EuiFormRow
            fullWidth
            label={
              <EuiToolTip content="Specify logger mode">
                <span>
                  Logger mode{" "}
                  <EuiIcon type="questionInCircle" color="subdued" />
                </span>
              </EuiToolTip>
            }
            display="columnCompressed">
            <EuiSuperSelect
              options={loggerModeOptions}
              valueOfSelected={selectedMode}
              onChange={value => setValue("mode", value)}
              itemLayoutAlign="top"
              hasDividers
            />
          </EuiFormRow>
        )}
      </EuiForm>
    </>
  );
};

LoggerForm.propTypes = {
  logger: PropTypes.object,
  config_type: PropTypes.string,
  onChange: PropTypes.func
};
