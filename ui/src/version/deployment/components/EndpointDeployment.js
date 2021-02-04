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

import React, { Fragment, useCallback, useEffect, useState } from "react";
import { navigate } from "@reach/router";
import {
  EuiButton,
  EuiButtonEmpty,
  EuiConfirmModal,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiOverlayMask,
  EuiPanel,
  EuiProgress,
  EuiSpacer,
  EuiTitle
} from "@elastic/eui";
import { replaceBreadcrumbs, useToggle } from "@gojek/mlp-ui";
import { useMerlinApi } from "../../../hooks/useMerlinApi";
import mocks from "../../../mocks";
import { EndpointEnvironment } from "./EndpointEnvironment";
import { EndpointVariables } from "./EndpointVariables";
import { ResourceRequest } from "./ResourceRequest";
import { Transformer } from "./Transformer";
import { LoggerForm, DEFAULT_LOGGER_CONFIG } from "./LoggerForm";
import PropTypes from "prop-types";

const DeployConfirmationModal = ({
  actionTitle,
  content,
  isLoading,
  onConfirm,
  onCancel
}) => (
  <EuiOverlayMask>
    <EuiConfirmModal
      title={`${actionTitle} model version`}
      onCancel={onCancel}
      onConfirm={onConfirm}
      cancelButtonText="Cancel"
      confirmButtonText={actionTitle}>
      {content}
      {isLoading && <EuiProgress size="xs" color="accent" />}
    </EuiConfirmModal>
  </EuiOverlayMask>
);

DeployConfirmationModal.propTypes = {
  actionTitle: PropTypes.string,
  content: PropTypes.string,
  isLoading: PropTypes.bool,
  onConfirm: PropTypes.func,
  onCancel: PropTypes.func
};

const defaultResourceRequest = {
  cpu_request: "500m",
  memory_request: "500Mi",
  min_replica: process.env.REACT_APP_ENVIRONMENT === "production" ? 2 : 0,
  max_replica: process.env.REACT_APP_ENVIRONMENT === "production" ? 4 : 2
};

const targetRequestStatus = currentStatus => {
  return currentStatus === "serving" ? "serving" : "running";
};

const isRequestConfigured = request => {
  if (request.transformer && request.transformer.enabled) {
    if (
      (request.transformer.transformer_type === "" ||
        request.transformer.transformer_type === "custom") &&
      !request.transformer.image
    ) {
      return false;
    }
  }

  if (!request.environment_name) {
    return false;
  }

  return true;
};

export const EndpointDeployment = ({
  actionTitle,
  breadcrumbs,
  model,
  version,
  endpointId,
  disableEnvironment,
  modalContent,
  onDeploy,
  redirectUrl,
  response
}) => {
  useEffect(() => {
    breadcrumbs && replaceBreadcrumbs([...breadcrumbs, { text: actionTitle }]);
  }, [actionTitle, breadcrumbs]);

  const [request, setRequest] = useState({});

  useEffect(() => {
    version.endpoints &&
      endpointId &&
      setRequest(version.endpoints.find(e => e.id === endpointId));
  }, [version.endpoints, endpointId, setRequest]);

  const [{ data: environments }] = useMerlinApi(
    `/environments`,
    { mock: mocks.environmentList },
    [],
    true
  );

  const [isModalVisible, toggleModalVisible] = useToggle();

  const openModal = () => {
    if (!isModalVisible && response.isLoaded) {
      response.isLoaded = false;
    }
    toggleModalVisible();
  };

  useEffect(() => {
    if (response.isLoaded) {
      isModalVisible && toggleModalVisible();
      if (!response.error) {
        navigate(redirectUrl);
      }
    }
  }, [response, isModalVisible, toggleModalVisible, redirectUrl]);

  // first, get selected environment's default resource request
  // next, check if endpoint in selected environment already exists or not
  // if exists and status is not pending, it's a redeployment so let's get previous configurations
  useEffect(() => {
    const selectedEnvironment = environments.find(
      e => e.name === request.environment_name
    );

    let targetResourceRequest = defaultResourceRequest;
    let targetEnvVars = [];
    let targetTransformer = undefined;
    let targetLogger = undefined;

    if (selectedEnvironment) {
      if (selectedEnvironment.default_resource_request) {
        targetResourceRequest = selectedEnvironment.default_resource_request;
      }
    }

    const endpoint = version.endpoints.find(
      e => e.environment_name === request.environment_name
    );

    if (endpoint) {
      if (endpoint.resource_request) {
        targetResourceRequest = endpoint.resource_request;
      }
      if (endpoint.env_vars) {
        targetEnvVars = endpoint.env_vars.filter(
          envVar => envVar.name !== "MODEL_NAME" && envVar.name !== "MODEL_DIR"
        );
      }
      if (endpoint.transformer) {
        targetTransformer = endpoint.transformer;
      }
      if (endpoint.logger) {
        targetLogger = endpoint.logger;
      }
    }

    setRequest(r => ({
      ...r,
      resource_request: targetResourceRequest,
      env_vars: targetEnvVars,
      transformer: targetTransformer,
      logger: targetLogger
    }));
  }, [environments, version.endpoints, request.environment_name, setRequest]);

  const onChange = field => {
    return value => setRequest(r => ({ ...r, [field]: value }));
  };

  const onVariablesChange = useCallback(onChange("env_vars"), []);

  const defaultLogger = {
    model: DEFAULT_LOGGER_CONFIG,
    transformer: DEFAULT_LOGGER_CONFIG
  };

  return (
    <Fragment>
      <EuiSpacer size="l" />

      <EuiFlexGroup justifyContent="spaceAround">
        <EuiFlexItem style={{ maxWidth: 700 }}>
          <EuiForm
            isInvalid={!!response.error}
            error={response.error ? [response.error.message] : ""}>
            <EuiFlexGroup direction="column">
              <EuiFlexItem>
                <EndpointEnvironment
                  version={version}
                  selected={request.environment_name}
                  environments={environments}
                  disabled={disableEnvironment}
                  onChange={onChange("environment_name")}
                />
              </EuiFlexItem>

              <EuiFlexItem grow={false}>
                <EuiSpacer size="s" />
                <EuiPanel grow={false}>
                  <EuiTitle size="xs">
                    <h4>Model Configuration</h4>
                  </EuiTitle>
                  <ResourceRequest
                    resourceRequest={
                      request.resource_request || defaultResourceRequest
                    }
                    onChange={onChange("resource_request")}
                  />
                  <EuiSpacer size="l" />
                  <LoggerForm
                    logger={request.logger || defaultLogger}
                    config_type="model"
                    onChange={onChange("logger")}
                  />
                </EuiPanel>
              </EuiFlexItem>

              {model.type === "pyfunc" && (
                <EuiFlexItem grow={false}>
                  <EuiSpacer size="s" />
                  {/* TODO: Use new EnvironmentVariables.js */}
                  <EndpointVariables
                    variables={request.env_vars || []}
                    onChange={onVariablesChange}
                  />
                </EuiFlexItem>
              )}

              <EuiFlexItem grow={false}>
                <EuiSpacer size="s" />
                <Transformer
                  transformer={
                    request.transformer || {
                      image: "",
                      resource_request: defaultResourceRequest
                    }
                  }
                  onChange={onChange("transformer")}
                  logger={request.logger || defaultLogger}
                  onLoggerChange={onChange("logger")}
                />
              </EuiFlexItem>

              <EuiFlexItem grow={false}>
                <EuiFlexGroup direction="row" justifyContent="flexEnd">
                  <EuiFlexItem grow={false}>
                    <EuiButtonEmpty
                      size="s"
                      onClick={() => navigate(redirectUrl)}>
                      Cancel
                    </EuiButtonEmpty>
                  </EuiFlexItem>

                  <EuiFlexItem grow={false}>
                    <EuiButton
                      size="s"
                      color="primary"
                      fill
                      disabled={!isRequestConfigured(request)}
                      onClick={openModal}>
                      {actionTitle}
                    </EuiButton>
                  </EuiFlexItem>
                </EuiFlexGroup>
              </EuiFlexItem>
            </EuiFlexGroup>
          </EuiForm>
        </EuiFlexItem>

        {isModalVisible && (
          <DeployConfirmationModal
            actionTitle={actionTitle}
            content={modalContent}
            isLoading={response.isLoading}
            onConfirm={() =>
              onDeploy({
                body: JSON.stringify({
                  ...request,
                  status: targetRequestStatus(request.status)
                })
              })
            }
            onCancel={toggleModalVisible}
          />
        )}
      </EuiFlexGroup>
    </Fragment>
  );
};

EndpointDeployment.propTypes = {
  actionTitle: PropTypes.string,
  breadcrumbs: PropTypes.array,
  model: PropTypes.object,
  version: PropTypes.object,
  endpointId: PropTypes.string,
  disableEnvironment: PropTypes.bool,
  modalContent: PropTypes.string,
  onDeploy: PropTypes.func,
  response: PropTypes.object
};
