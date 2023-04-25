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

import React, { Fragment, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  EuiButtonEmpty,
  EuiFlexGroup,
  EuiFlexItem,
  EuiText
} from "@elastic/eui";
import { useToggle } from "@gojek/mlp-ui";
import config from "../../config";
import {
  ServeVersionEndpointModal,
  StopServeVersionEndpointModal,
  UndeployVersionEndpointModal
} from "../../components/modals";

const ActionButton = ({ action }) => (
  <EuiButtonEmpty
    color={action.color}
    iconType={action.iconType}
    flush="left"
    size="xs"
    onClick={() => action.onClick()}>
    <EuiText size="xs">{action.name}</EuiText>
  </EuiButtonEmpty>
);

export const DeploymentActions = ({ model, version, endpoint }) => {
  const navigate = useNavigate();
  const [isServeEndpointModalVisible, toggleServeEndpointModal] = useToggle();
  const [
    isStopServeEndpointModalVisible,
    toggleStopServeEndpointModal
  ] = useToggle();
  const [
    isUndeployEndpointModalVisible,
    toggleUndeployEndpointModal
  ] = useToggle();

  const [modelEndpoint, setModelEndpoint] = useState({});
  useEffect(() => {
    if (model.endpoints && endpoint.status === "serving") {
      const modelEndpoint = model.endpoints.find(
        modelEndpoint =>
          modelEndpoint.environment_name === endpoint.environment_name
      );
      setModelEndpoint(modelEndpoint);
    }
  }, [model.endpoints, endpoint, setModelEndpoint]);

  const actions = [
    {
      name: "Start serving",
      iconType: "play",
      enabled: endpoint.status === "running",
      onClick: () => toggleServeEndpointModal()
    },
    {
      name: "Redeploy",
      iconType: "importAction",
      enabled: endpoint.status !== "pending",
      onClick: () =>
        navigate(
          `${config.HOMEPAGE}/projects/${model.project_id}/models/${model.id}/versions/${version.id}/endpoints/${endpoint.id}/redeploy`
        )
    },
    {
      name: "Undeploy",
      iconType: "exportAction",
      color: "danger",
      enabled:
        endpoint.status === "pending" ||
        endpoint.status === "running" ||
        endpoint.status === "failed",
      onClick: () => toggleUndeployEndpointModal()
    },
    {
      name: "Stop serving",
      iconType: "minusInCircle",
      color: "danger",
      enabled: endpoint.status === "serving",
      onClick: () => toggleStopServeEndpointModal()
    }
  ];

  return (
    <Fragment>
      <EuiFlexGroup alignItems="flexStart" direction="column" gutterSize="s">
        {actions.map(
          action =>
            action.enabled && (
              <EuiFlexItem grow={false} key={`${action.name}-${endpoint.id}`}>
                <ActionButton action={action} />
              </EuiFlexItem>
            )
        )}
      </EuiFlexGroup>

      {isServeEndpointModalVisible && (
        <ServeVersionEndpointModal
          model={model}
          version={version}
          versionEndpoint={endpoint}
          callback={() => window.location.reload()}
          closeModal={toggleServeEndpointModal}
        />
      )}

      {isStopServeEndpointModalVisible && (
        <StopServeVersionEndpointModal
          model={model}
          modelEndpoint={modelEndpoint}
          callback={() => window.location.reload()}
          closeModal={toggleStopServeEndpointModal}
        />
      )}

      {isUndeployEndpointModalVisible && (
        <UndeployVersionEndpointModal
          model={model}
          version={version}
          versionEndpoint={endpoint}
          callback={() => window.location.reload()}
          closeModal={toggleUndeployEndpointModal}
        />
      )}
    </Fragment>
  );
};
