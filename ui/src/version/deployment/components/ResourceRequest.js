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
import {
  EuiFieldText,
  EuiFieldNumber,
  EuiFormRow,
  EuiIcon,
  EuiToolTip
} from "@elastic/eui";
import PropTypes from "prop-types";

export const ResourceRequest = ({ resourceRequest, onChange }) => {
  const setValue = (field, value) =>
    onChange({
      ...resourceRequest,
      [field]: value
    });

  const setTextValue = field => {
    return e => setValue(field, e.target.value);
  };

  const setNumericValue = field => {
    return e => {
      const sanitizedValue = parseInt(e.target.value);
      !isNaN(sanitizedValue) && setValue(field, sanitizedValue);
    };
  };

  return (
    <>
      <EuiFormRow
        fullWidth
        label={
          <EuiToolTip content="Specify the minimum amount of CPU required for each replica. You can choose more than 1 CPU (e.g. 1.5 or 2) for CPU-intensive workloads.">
            <span>
              CPU* <EuiIcon type="questionInCircle" color="subdued" />
            </span>
          </EuiToolTip>
        }
        display="columnCompressed">
        <EuiFieldText
          placeholder="1"
          value={resourceRequest.cpu_request}
          onChange={setTextValue("cpu_request")}
          name="cpu"
        />
      </EuiFormRow>

      <EuiFormRow
        fullWidth
        label={
          <EuiToolTip content="Specify the minimum amount of memory required for each replica.">
            <span>
              Memory* <EuiIcon type="questionInCircle" color="subdued" />
            </span>
          </EuiToolTip>
        }
        display="columnCompressed">
        <EuiFieldText
          placeholder="500Mi"
          value={resourceRequest.memory_request}
          onChange={setTextValue("memory_request")}
          name="memory"
        />
      </EuiFormRow>

      <EuiFormRow fullWidth label="Min replicas*" display="columnCompressed">
        <EuiFieldNumber
          placeholder="0"
          min={0}
          value={resourceRequest.min_replica}
          onChange={setNumericValue("min_replica")}
          name="minReplicas"
        />
      </EuiFormRow>

      <EuiFormRow fullWidth label="Max replicas*" display="columnCompressed">
        <EuiFieldNumber
          placeholder="2"
          min={0}
          value={resourceRequest.max_replica}
          onChange={setNumericValue("max_replica")}
          name="maxReplicas"
        />
      </EuiFormRow>
    </>
  );
};

ResourceRequest.propTypes = {
  resourceRequest: PropTypes.object,
  onChange: PropTypes.func
};
