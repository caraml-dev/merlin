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
  EuiForm,
  EuiFormRow,
  EuiIcon,
  EuiPanel,
  EuiTitle,
  EuiToolTip
} from "@elastic/eui";

export const EndpointResources = ({ resourceRequest, onChange }) => {
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
    <EuiPanel grow={false}>
      <EuiTitle size="xs">
        <h4>Resources</h4>
      </EuiTitle>
      <EuiForm>
        <EuiFormRow
          fullWidth
          label={
            <EuiToolTip content="Specify the total amount of CPU available for model serving. For CPU-intensive models, you can choose more than 1 CPU (e.g. 1.5).">
              <span>
                CPU* <EuiIcon type="questionInCircle" color="subdued" />
              </span>
            </EuiToolTip>
          }
          display="columnCompressed">
          <EuiFieldText
            placeholder="500m"
            value={resourceRequest.cpu_request}
            onChange={setTextValue("cpu_request")}
            name="cpu"
          />
        </EuiFormRow>

        <EuiFormRow
          fullWidth
          label={
            <EuiToolTip content="Specify the total amount of RAM available for model serving.">
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
      </EuiForm>
    </EuiPanel>
  );
};
