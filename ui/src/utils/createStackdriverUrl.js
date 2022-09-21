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

const stackdriverAPI = "https://console.cloud.google.com/logs/viewer";

const stackdriverFilter = query => {
  return `resource.type:"k8s_query" OR "k8s_container" OR "k8s_pod"
resource.labels.project_id:${query.gcp_project}
resource.labels.cluster_name:${query.cluster}
resource.labels.namespace_name:${query.namespace}
resource.labels.pod_name:${query.pod_name}`;
};

export const createStackdriverUrl = query => {
  const advanceFilter = stackdriverFilter(query);
  const url = {
    interval: "PT1H",
    project: query.gcp_project,
    minLogLevel: 0,
    expandAll: false,
    advancedFilter: advanceFilter
  };

  const stackdriverParams = new URLSearchParams(url).toString();
  return stackdriverAPI + "?" + stackdriverParams;
};
