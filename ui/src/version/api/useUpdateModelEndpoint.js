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

import { useEffect, useState } from "react";
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

export default useUpdateModelEndpoint;
