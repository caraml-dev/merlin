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
import { EuiConfirmModal, EuiOverlayMask, EuiProgress } from "@elastic/eui";
import { useMerlinApi } from "../../hooks/useMerlinApi";

const useUpdateModelEndpoint = (versionEndpoint, model, modelEndpointId) => {
  const [fetchModelEndpointResponse, fetchModelEndpoint] = useMerlinApi(
    `/models/${model.id}/endpoints/${modelEndpointId}`,
    {},
    {},
    false
  );

  const [updateModelEndpointResponse, doUpdateModelEndpoint] = useMerlinApi(
    `/models/${model.id}/endpoints/${modelEndpointId}`,
    { method: "PUT", addToast: true },
    {},
    false
  );

  const [state, setState] = useState(fetchModelEndpointResponse);

  // Fetch model endpoint first
  useEffect(() => {
    if (fetchModelEndpointResponse.isLoaded) {
      if (fetchModelEndpointResponse.error) {
        setState(fetchModelEndpointResponse);
      } else {
        const modelEndpointRequest = {
          ...fetchModelEndpointResponse.data,
          environment_name: versionEndpoint.environment_name,
          rule: {
            destinations: [
              {
                weight: 100,
                version_endpoint_id: versionEndpoint.id
              }
            ]
          }
        };

        doUpdateModelEndpoint({ body: JSON.stringify(modelEndpointRequest) });
      }
    }
  }, [
    versionEndpoint,
    model,
    fetchModelEndpointResponse,
    doUpdateModelEndpoint
  ]);

  useEffect(() => {
    if (updateModelEndpointResponse.isLoaded) {
      setState(updateModelEndpointResponse);
    }
  }, [updateModelEndpointResponse]);

  return [state, () => fetchModelEndpoint()];
};

const ServeVersionEndpointModal = ({
  versionEndpoint,
  version,
  model,
  callback,
  closeModal
}) => {
  const [createEndpointResponse, createModelEndpoint] = useMerlinApi(
    `/models/${model.id}/endpoints`,
    {
      method: "POST",
      addToast: true,
      body: JSON.stringify({
        model_id: model.id,
        environment_name: versionEndpoint.environment_name,
        rule: {
          destinations: [
            {
              weight: 100,
              version_endpoint_id: versionEndpoint.id
            }
          ]
        }
      })
    },
    {},
    false
  );

  const [modelEndpoint] = useState(
    model.endpoints.find(endpoint => {
      return endpoint.environment_name === versionEndpoint.environment_name;
    })
  );

  const [modelEndpointId, setModelEndpointId] = useState("");
  useEffect(() => {
    if (modelEndpoint) {
      setModelEndpointId(modelEndpoint.id);
    }
  }, [modelEndpoint]);

  const [
    updateEndpointResponse,
    doUpdateModelEndpoint
  ] = useUpdateModelEndpoint(versionEndpoint, model, modelEndpointId);

  useEffect(() => {
    if (createEndpointResponse.isLoaded || updateEndpointResponse.isLoaded) {
      closeModal();
      callback();
    }
  }, [createEndpointResponse, updateEndpointResponse, closeModal, callback]);

  const toggleServing = () => {
    model.endpoints.find(endpoint => {
      return endpoint.environment_name === versionEndpoint.environment_name;
    })
      ? doUpdateModelEndpoint()
      : createModelEndpoint();
  };

  return (
    <EuiOverlayMask>
      <EuiConfirmModal
        title="Start model serving"
        onCancel={closeModal}
        onConfirm={toggleServing}
        cancelButtonText="Cancel"
        confirmButtonText="Start Serving">
        <p>
          You're about to set version <b>{version.id}</b> as the serving
          endpoint for model <b>{model.name}</b> in{" "}
          <b>{versionEndpoint.environment_name}</b> environment.
        </p>
        {(createEndpointResponse.isLoading ||
          updateEndpointResponse.isLoading) && (
          <EuiProgress
            size="xs"
            color="accent"
            className="euiProgress-beforePre"
          />
        )}
      </EuiConfirmModal>
    </EuiOverlayMask>
  );
};

export default ServeVersionEndpointModal;
