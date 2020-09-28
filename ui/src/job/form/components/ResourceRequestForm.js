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
import {
  EuiFieldNumber,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
  EuiSpacer
} from "@elastic/eui";
import { useMerlinApi } from "../../../hooks/useMerlinApi";
import mocks from "../../../mocks";

export const ResourceRequestForm = ({ resourceRequest, onChange }) => {
  const [defaultResource, setDefaultResource] = useState({
    driver_cpu_request: "",
    driver_memory_request: "",
    executor_cpu_request: "",
    executor_memory_request: "",
    executor_replica: 0
  });

  const [{ data: environments }] = useMerlinApi(
    `/environments`,
    { mock: mocks.environmentList },
    [],
    true
  );

  useEffect(
    () => {
      if (environments) {
        const defaultEnv = environments.find(
          e => e.is_prediction_job_enabled && e.is_default_prediction_job
        );
        if (Object.keys(resourceRequest).length === 0 && defaultEnv) {
          setDefaultResource(
            defaultEnv.default_prediction_job_resource_request
          );

          onChange(
            "driver_cpu_request",
            defaultEnv.default_prediction_job_resource_request
              .driver_cpu_request
          );
          onChange(
            "driver_memory_request",
            defaultEnv.default_prediction_job_resource_request
              .driver_memory_request
          );
          onChange(
            "executor_cpu_request",
            defaultEnv.default_prediction_job_resource_request
              .executor_cpu_request
          );
          onChange(
            "executor_memory_request",
            defaultEnv.default_prediction_job_resource_request
              .executor_memory_request
          );
          onChange(
            "executor_replica",
            defaultEnv.default_prediction_job_resource_request.executor_replica
          );
        }
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [resourceRequest, environments]
  );

  return (
    <EuiForm>
      <EuiFlexGroup>
        <EuiFlexItem>
          <EuiFormRow
            label="Driver CPU Request"
            helpText="Number of cores to use for the driver process.">
            <EuiFieldText
              value={
                resourceRequest.driver_cpu_request
                  ? resourceRequest.driver_cpu_request
                  : defaultResource.driver_cpu_request
              }
              onChange={e => onChange("driver_cpu_request", e.target.value)}
              aria-label="Resource Request Driver CPU"
            />
          </EuiFormRow>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFormRow
            label="Driver Memory Request"
            helpText="Amount of memory to use for the driver process.">
            <EuiFieldText
              value={
                resourceRequest.driver_memory_request
                  ? resourceRequest.driver_memory_request
                  : defaultResource.driver_memory_request
              }
              onChange={e => onChange("driver_memory_request", e.target.value)}
              aria-label="Resource Request Driver Memory"
            />
          </EuiFormRow>
        </EuiFlexItem>
      </EuiFlexGroup>

      <EuiSpacer size="l" />

      <EuiFlexGroup>
        <EuiFlexItem>
          <EuiFormRow
            label="Executor CPU Request"
            helpText="The number of cores to use on each executor.">
            <EuiFieldText
              value={
                resourceRequest.executor_cpu_request
                  ? resourceRequest.executor_cpu_request
                  : defaultResource.executor_cpu_request
              }
              onChange={e => onChange("executor_cpu_request", e.target.value)}
              aria-label="Resource Request Executor CPU"
            />
          </EuiFormRow>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFormRow
            label="Executor Memory Request"
            helpText="Amount of memory to use per executor process.">
            <EuiFieldText
              value={
                resourceRequest.executor_memory_request
                  ? resourceRequest.executor_memory_request
                  : defaultResource.executor_memory_request
              }
              onChange={e =>
                onChange("executor_memory_request", e.target.value)
              }
              aria-label="Resource Request Executor Memory"
            />
          </EuiFormRow>
        </EuiFlexItem>
      </EuiFlexGroup>

      <EuiSpacer size="l" />

      <EuiFlexGroup>
        <EuiFlexItem>
          <EuiFormRow
            label="Executor Replica"
            helpText="The number of executor replica.">
            <EuiFieldNumber
              value={
                resourceRequest.executor_replica
                  ? resourceRequest.executor_replica
                  : defaultResource.executor_replica
              }
              onChange={e => {
                const sanitizedValue = parseInt(e.target.value);
                !isNaN(sanitizedValue) &&
                  onChange("executor_replica", sanitizedValue);
              }}
              aria-label="Resource Request Executor Replica"
              min={1}
            />
          </EuiFormRow>
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiForm>
  );
};
