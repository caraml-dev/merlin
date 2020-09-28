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

const querystring = require("querystring");

const stackdriverAPI = "https://console.cloud.google.com/logs/viewer";

const stackdriverFilter = container => {
  return `resource.type:"k8s_container"
resource.labels.project_id:${container.gcp_project}
resource.labels.cluster_name:${container.cluster}
resource.labels.namespace_name:${container.namespace}
resource.labels.pod_name:${container.pod_name}
resource.labels.container_name:${container.name}`;
};

export const createStackdriverUrl = container => {
  const advanceFilter = stackdriverFilter(container);
  const url = {
    interval: "NO_LIMIT",
    project: container.gcp_project,
    minLogLevel: 0,
    expandAll: false,
    advancedFilter: advanceFilter
  };

  const stackdriverParams = querystring.stringify(url);
  return stackdriverAPI + "?" + stackdriverParams;
};
