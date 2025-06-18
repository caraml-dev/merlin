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

import objectAssignDeep from "object-assign-deep";
import React from "react";

const getEnv = (env) => {
  return window.env && env in window.env ? window.env[env] : process.env[env];
};

export const sentryConfig = {
  dsn: getEnv("REACT_APP_SENTRY_DSN"),
  environment: getEnv("REACT_APP_ENVIRONMENT"),
};

export const appConfig = {
  appIcon: "machineLearningApp",
  docsUrl: getEnv("REACT_APP_MERLIN_DOCS_URL") || [
    {
      href: "https://github.com/caraml-dev/merlin/blob/main/docs/getting-started/README.md",
      label: "Getting Started with Merlin",
    },
  ],
  upiDocumentationUrl: getEnv("REACT_APP_UPI_DOC_URL"),
  dockerRegistries: getEnv("REACT_APP_DOCKER_REGISTRIES")
    ? getEnv("REACT_APP_DOCKER_REGISTRIES").split(",")
    : [],
  defaultDockerRegistry:
    process.env.REACT_APP_DEFAULT_DOCKER_REGISTRY || "docker.io", // User Docker Hub as the default
  defaultFeastSource: getEnv("REACT_APP_DEFAULT_FEAST_SOURCE") || "BIGTABLE",
  feastServingUrls: getEnv("REACT_APP_FEAST_SERVING_URLS") || [
    {
      host: "online-serving-redis.feast.dev",
      label: "Online Serving with Redis",
      icon: "redis",
      source_type: "REDIS",
    },
    {
      host: "online-serving-bigtable.feast.dev",
      label: "Online Serving with BigTable",
      icon: "bigtable",
      source_type: "BIGTABLE",
    },
  ],
  imagebuilder:{
    cluster:  getEnv("REACT_APP_IMAGE_BUILDER_CLUSTER"),
    gcp_project: getEnv("REACT_APP_IMAGE_BUILDER_GCP_PROJECT"),
    namespace: getEnv("REACT_APP_IMAGE_BUILDER_NAMESPACE"),
  },
  pagination: {
    defaultPageSize: 10,
  }
};

export const featureToggleConfig = {
  alertEnabled: getEnv("REACT_APP_ALERT_ENABLED")
    ? !(getEnv("REACT_APP_ALERT_ENABLED").toString().toLowerCase() === "false")
    : false,
  monitoringEnabled: getEnv("REACT_APP_MONITORING_DASHBOARD_ENABLED")
    ? !(
        getEnv("REACT_APP_MONITORING_DASHBOARD_ENABLED")
          .toString()
          .toLowerCase() === "false"
      )
    : false,
  monitoringDashboardJobBaseURL: getEnv(
    "REACT_APP_MONITORING_DASHBOARD_JOB_BASE_URL"
  ),
  modelDeletionEnabled: getEnv("REACT_APP_MODEL_DELETION_ENABLED")
    ? !(
        getEnv("REACT_APP_MODEL_DELETION_ENABLED").toString().toLowerCase() ===
        "false"
      )
    : false,
  logImageBuilderURL: getEnv(
      "REACT_APP_LOG_IMAGE_BUILDER_URL"
  ),
  logModelURL: getEnv(
      "REACT_APP_LOG_MODEL_URL"
  ),
};

export const costEstimationConfig = {
  // Default based on N1 on-demand pricing in asia-southeast1 / Singapore
  // Data taken on January 2023
  cpuCost: getEnv("REACT_APP_CPU_COST") || 28.46927,
  memoryCost: getEnv("REACT_APP_MEMORY_COST") || 3.81498,
};

const config = {
  HOMEPAGE: getEnv("REACT_APP_HOMEPAGE") || process.env.PUBLIC_URL,
  USE_MOCK_DATA: false,
  TIMEOUT: 20000,
  MERLIN_API: getEnv("REACT_APP_MERLIN_API"),
  MLP_API: getEnv("REACT_APP_MLP_API"),
  FEAST_CORE_API: getEnv("REACT_APP_FEAST_CORE_API"),
  OAUTH_CLIENT_ID: getEnv("REACT_APP_OAUTH_CLIENT_ID"),
};

export default config;

const ConfigContext = React.createContext({});

export const ConfigProvider = ({ children }) => {
  const runTimeConfig = window.config;
  const config = objectAssignDeep({}, runTimeConfig);

  return (
    <ConfigContext.Provider value={config}>{children}</ConfigContext.Provider>
  );
};

export const useConfig = () => React.useContext(ConfigContext);

