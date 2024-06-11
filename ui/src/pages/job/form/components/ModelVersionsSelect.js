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
import { EuiFormRow, EuiSuperSelect, EuiText } from "@elastic/eui";
import PropTypes from "prop-types";

const moment = require("moment");

export const ModelVersionSelect = ({
  isDisabled,
  selected,
  versions,
  onChange,
}) => {
  const [options, setOptions] = useState([]);

  useEffect(() => {
    if (versions) {
      const options = [];
      versions
        .sort((a, b) => (a.id > b.id ? -1 : 1))
        .forEach((version) => {
          options.push({
            value: version.id,
            inputDisplay: (
              <Fragment>
                Model version <strong>{version.id}</strong>
              </Fragment>
            ),
            dropdownDisplay: (
              <Fragment>
                Model version <strong>{version.id}</strong>
                <EuiText size="xs">
                  <p className="euiTextColor--subdued">
                    Created at{" "}
                    {moment(
                      version.created_at,
                      "YYYY-MM-DDTHH:mm.SSZ",
                    ).fromNow()}
                  </p>
                </EuiText>
              </Fragment>
            ),
          });
        });
      setOptions(options);
    }
  }, [versions]);

  return (
    <EuiFormRow
      label="Select Model Version *"
      helpText="Choose model version to be used for batch prediction job."
    >
      <EuiSuperSelect
        disabled={isDisabled}
        options={options}
        valueOfSelected={selected ? selected : ""}
        onChange={onChange}
        hasDividers
      />
    </EuiFormRow>
  );
};

ModelVersionSelect.propTypes = {
  isDisabled: PropTypes.bool,
  selected: PropTypes.string,
  versions: PropTypes.array,
  onChange: PropTypes.func,
};
