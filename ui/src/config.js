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

const getEnv = env => {
  return window.env && env in window.env ? window.env[env] : process.env[env];
};

export const sentryConfig = {
  dsn: getEnv("REACT_APP_SENTRY_DSN"),
  environment: getEnv("REACT_APP_ENVIRONMENT")
};

export const appConfig = {
  appIcon: "machineLearningApp",
  docsUrl: getEnv("REACT_APP_MERLIN_DOCS_URL") || [
    {
      href:
        "https://github.com/gojek/merlin/blob/main/docs/getting-started/README.md",
      label: "Getting Started with Merlin"
    }
  ],
  dockerRegistries: getEnv("REACT_APP_DOCKER_REGISTRIES")
    ? getEnv("REACT_APP_DOCKER_REGISTRIES").split(",")
    : [],
  defaultDockerRegistry:
    process.env.REACT_APP_DEFAULT_DOCKER_REGISTRY || "docker.io", // User Docker Hub as the default
  scaling: {
    maxAllowedReplica: 10
  }
};

export const featureToggleConfig = {
  alertEnabled: getEnv("REACT_APP_ALERT_ENABLED")
    ? !(
        getEnv("REACT_APP_ALERT_ENABLED")
          .toString()
          .toLowerCase() === "false"
      )
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
  )
};

const config = {
  HOMEPAGE: getEnv("REACT_APP_HOMEPAGE") || process.env.PUBLIC_URL,
  USE_MOCK_DATA: false,
  TIMEOUT: 20000,
  MERLIN_API: getEnv("REACT_APP_MERLIN_API"),
  MLP_API: getEnv("REACT_APP_MLP_API"),
  FEAST_CORE_API: getEnv("REACT_APP_FEAST_CORE_API"),
  OAUTH_CLIENT_ID: getEnv("REACT_APP_OAUTH_CLIENT_ID")
};

export default {
  // Add common config values here
  ...config
};
