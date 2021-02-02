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
  listProjects: "/feast.core.CoreService/ListProjects",
  listEntities: "/feast.core.CoreService/ListEntities",
  listFeatureTables: "/feast.core.CoreService/ListFeatureTables"
};

export const useFeastApi = (
  endpoint,
  options,
  result,
  callImmediately = true
) => {
  const authCtx = undefined;

  return useApi(
    endpoint,
    {
      baseApiUrl: config.FEAST_API,
      timeout: config.TIMEOUT,
      useMockData: config.USE_MOCK_DATA,
      ...options
    },
    authCtx,
    result,
    callImmediately
  );
};
