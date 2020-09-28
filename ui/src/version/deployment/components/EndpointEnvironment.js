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

import React, { Fragment, useEffect, useState } from "react";
import {
  EuiForm,
  EuiFormRow,
  EuiIcon,
  EuiPanel,
  EuiSuperSelect,
  EuiText,
  EuiTextColor,
  EuiTitle,
  EuiToolTip
} from "@elastic/eui";

const EnvironmentDropdownOption = ({ environment, endpoint, disabled }) => {
  const option = (
    <Fragment>
      <strong>{environment.name}</strong>
      <EuiText size="s">
        <EuiTextColor color="subdued">{environment.cluster}</EuiTextColor>
      </EuiText>
      {endpoint && (
        <EuiText size="s">
          <EuiTextColor color="subdued">{endpoint.status}</EuiTextColor>
        </EuiText>
      )}
    </Fragment>
  );

  return disabled ? (
    <EuiToolTip
      position="left"
      content={`${environment.name} already has a ${endpoint.status} endpoint`}>
      {option}
    </EuiToolTip>
  ) : (
    option
  );
};

const isEnvironmentDisabled = endpoint => {
  return (
    endpoint &&
    (endpoint.status === "serving" ||
      endpoint.status === "running" ||
      endpoint.status === "pending")
  );
};

export const EndpointEnvironment = ({
  version,
  selected,
  environments,
  onChange,
  disabled
}) => {
  const [environmentOptions, setEnvironmentOptions] = useState([]);

  useEffect(() => {
    if (version) {
      const options = environments
        .sort((a, b) => (a.name > b.name ? 1 : -1))
        .map(environment => {
          const versionEndpoint = version.endpoints.find(ve => {
            return ve.environment_name === environment.name;
          });

          const isDisabled = isEnvironmentDisabled(versionEndpoint);

          return {
            value: environment.name,
            inputDisplay: environment.name,
            disabled: isDisabled,
            dropdownDisplay: (
              <EnvironmentDropdownOption
                environment={environment}
                endpoint={versionEndpoint}
                disabled={isDisabled}
              />
            )
          };
        });

      setEnvironmentOptions(options);
    }
  }, [version, environments, setEnvironmentOptions]);

  return (
    <EuiPanel grow={false}>
      <EuiTitle size="xs">
        <h4>Environment</h4>
      </EuiTitle>

      <EuiForm>
        <EuiFormRow
          fullWidth
          label={
            <EuiToolTip content="Specify the target environment your model version will be deployed to.">
              <span>
                Environment* <EuiIcon type="questionInCircle" color="subdued" />
              </span>
            </EuiToolTip>
          }
          display="columnCompressed">
          <EuiSuperSelect
            options={environmentOptions}
            valueOfSelected={selected || ""}
            onChange={onChange}
            itemLayoutAlign="top"
            hasDividers
            disabled={disabled}
          />
        </EuiFormRow>
      </EuiForm>
    </EuiPanel>
  );
};
