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

import { useApi } from "@gojek/mlp-ui";
import config from "../config";

export const feastEndpoints = {
  listProjects: "/feast/api/projects",
  listEntities: "/feast/api/entities",
  listFeatureTables: "/feast/api/tables",
};

export const useFeastApi = (
  endpoint,
  options,
  result,
  callImmediately = true
) => {
  /* Use undefined for authCtx so that the authorization header passed in the options
   * will be used instead of being overwritten. Ref: https://github.com/gojek/mlp/blob/main/ui/packages/lib/src/utils/fetchJson.js#L39*/

  return useApi(
    endpoint,
    {
      baseApiUrl: config.FEAST_CORE_API,
      timeout: config.TIMEOUT,
      useMockData: config.USE_MOCK_DATA,
      headers: {
        "Content-Type": "application/json",
      },

      ...options,
    },
    undefined,
    result,
    callImmediately
  );
};
