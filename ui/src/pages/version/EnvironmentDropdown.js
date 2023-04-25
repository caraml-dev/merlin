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
import PropTypes from "prop-types";
import { useNavigate } from "react-router-dom";
import {
  EuiSuperSelect,
  EuiText,
  EuiTextColor,
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

export const EnvironmentDropdown = ({ version, selected, environments }) => {
  const navigate = useNavigate();
  const [environmentOptions, setEnvironmentOptions] = useState([]);

  useEffect(() => {
    if (version) {
      const options = environments
        .sort((a, b) => (a.name > b.name ? 1 : -1))
        .map(environment => {
          const versionEndpoint = version.endpoints.find(ve => {
            return ve.environment_name === environment.name;
          });

          return {
            value: versionEndpoint.id,
            inputDisplay: environment.name,
            dropdownDisplay: (
              <EnvironmentDropdownOption
                environment={environment}
                endpoint={versionEndpoint}
              />
            )
          };
        });

      setEnvironmentOptions(options);
    }
  }, [version, environments, setEnvironmentOptions]);

  return (
    <EuiSuperSelect
      compressed={true}
      options={environmentOptions}
      valueOfSelected={selected || ""}
      onChange={value =>
        navigate(
          `/merlin/projects/${version.model.project_id}/models/${version.model.id}/versions/${version.id}/endpoints/${value}/details`
        )
      }
      itemLayoutAlign="top"
      hasDividers
    />
  );
};

EnvironmentDropdown.propTypes = {
  version: PropTypes.object.isRequired,
  selected: PropTypes.string.isRequired,
  environments: PropTypes.array.isRequired
};
