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

import { useToggle } from "@caraml-dev/ui-lib";
import {
  EuiButtonEmpty,
  EuiFlexGroup,
  EuiFlexItem,
  EuiText,
} from "@elastic/eui";
import React, { Fragment, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { StopServeVersionEndpointModal } from "../components/modals";

const VersionEndpointActions = ({
  versionEndpoint,
  activeVersion,
  activeModel,
  fetchVersions,
  ...props
}) => {
  const openModal = (toggleFn) => {
    return (event) => {
      event.preventDefault();
      props.setActiveVersionEndpoint(versionEndpoint);
      props.setActiveVersion(activeVersion);
      props.setActiveModel(activeModel);
      toggleFn();
    };
  };

  const [
    isStopServeModelEndpointModalVisible,
    toggleStopServeModelEndpointModal,
  ] = useToggle();
  const [modelEndpoint, setModelEndpoint] = useState({});

  useEffect(() => {
    if (activeModel.endpoints && versionEndpoint.status === "serving") {
      const modelEndpoint = activeModel.endpoints.find(
        (endpoint) =>
          endpoint.environment_name === versionEndpoint.environment_name
      );
      setModelEndpoint(modelEndpoint);
    }
  }, [activeModel.endpoints, versionEndpoint, setModelEndpoint]);

  const actions = {
    startServing: (
      <EuiFlexItem grow={false} key={`serving-${versionEndpoint.id}`}>
        <EuiButtonEmpty
          onClick={openModal(props.toggleServeEndpointModal)}
          iconType="play"
          size="xs"
        >
          <EuiText size="xs">Start Serving</EuiText>
        </EuiButtonEmpty>
      </EuiFlexItem>
    ),

    stopServing: (
      <EuiFlexItem grow={false} key={`stop-serving-${versionEndpoint.id}`}>
        <EuiButtonEmpty
          onClick={() => toggleStopServeModelEndpointModal()}
          color="danger"
          iconType="minusInCircle"
          size="xs"
        >
          <EuiText size="xs">Stop serving</EuiText>
        </EuiButtonEmpty>
      </EuiFlexItem>
    ),

    redeploy: (
      <EuiFlexItem grow={false} key={`deploy-${versionEndpoint.id}`}>
        <Link
          to={`${activeVersion.id}/endpoints/${versionEndpoint.id}/redeploy`}
          state={{
            model: activeModel,
            version: activeVersion,
          }}
        >
          <EuiButtonEmpty iconType="importAction" size="xs">
            <EuiText size="xs">Redeploy</EuiText>
          </EuiButtonEmpty>
        </Link>
      </EuiFlexItem>
    ),

    undeploy: (
      <EuiFlexItem grow={false} key={`undeploy-${versionEndpoint.id}`}>
        <EuiButtonEmpty
          onClick={openModal(props.toggleUndeployEndpointModal)}
          color="danger"
          iconType="exportAction"
          size="xs"
        >
          <EuiText size="xs">Undeploy</EuiText>
        </EuiButtonEmpty>
      </EuiFlexItem>
    ),

    monitoring: (
      <EuiFlexItem grow={false} key={`monitoring-${versionEndpoint.id}`}>
        <EuiButtonEmpty
          href={versionEndpoint.monitoring_url}
          iconType="visLine"
          size="xs"
          target="_blank"
        >
          <EuiText size="xs">Monitoring</EuiText>
        </EuiButtonEmpty>
      </EuiFlexItem>
    ),

    logging: (
      <EuiFlexItem grow={false} key={`logging-${versionEndpoint.id}`}>
        <Link to={`${activeVersion.id}/endpoints/${versionEndpoint.id}/logs`}>
          <EuiButtonEmpty iconType="logstashQueue" size="xs">
            <EuiText size="xs">Logging</EuiText>
          </EuiButtonEmpty>
        </Link>
      </EuiFlexItem>
    ),

    details: (
      <EuiFlexItem grow={false} key={`details-${versionEndpoint.id}`}>
        <Link to={`${activeVersion.id}/endpoints/${versionEndpoint.id}`}>
          <EuiButtonEmpty iconType="inspect" size="xs">
            <EuiText size="xs">Details</EuiText>
          </EuiButtonEmpty>
        </Link>
      </EuiFlexItem>
    ),
  };

  const tailorActions = (actions) => {
    let list = [];

    // Start serving
    if (versionEndpoint.status === "running") {
      list.push(actions.startServing);
    }

    // Redeploy
    if (versionEndpoint.status !== "pending") {
      list.push(actions.redeploy);
    }

    // Monitoring
    list.push(actions.monitoring);

    // Logging
    list.push(actions.logging);

    // Details
    list.push(actions.details);

    // Undeploy
    if (
      versionEndpoint.status === "running" ||
      versionEndpoint.status === "failed"
    ) {
      list.push(actions.undeploy);
    }

    // Stop serving
    if (versionEndpoint.status === "serving") {
      list.push(actions.stopServing);
    }

    return list;
  };

  return (
    <Fragment>
      <EuiFlexGroup alignItems="flexStart" direction="column" gutterSize="xs">
        {tailorActions(actions)}
      </EuiFlexGroup>

      {isStopServeModelEndpointModalVisible && (
        <StopServeVersionEndpointModal
          model={activeModel}
          modelEndpoint={modelEndpoint}
          callback={fetchVersions}
          closeModal={() => toggleStopServeModelEndpointModal()}
        />
      )}
    </Fragment>
  );
};

export default VersionEndpointActions;
