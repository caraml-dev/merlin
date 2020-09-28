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

import React, { Fragment } from "react";
import { Link } from "@reach/router";
import {
  EuiButtonEmpty,
  EuiFlexGroup,
  EuiFlexItem,
  EuiText
} from "@elastic/eui";
import { useToggle } from "@gojek/mlp-ui";
import StopServeModelEndpointModal from "./modals/StopServeModelEndpointModal";
import { featureToggleConfig } from "../config";

const ModelEndpointActions = ({ model, modelEndpoint, fetchModels }) => {
  const [
    isStopServeModelEndpointModalVisible,
    toggleStopServeModelEndpointModal
  ] = useToggle();

  const actions = {
    alert: (
      <EuiFlexItem grow={false} key={`alert-${modelEndpoint.id}`}>
        <Link
          to={`${model.id}/endpoints/${modelEndpoint.id}/alert`}
          state={{
            model: model,
            endpoint: modelEndpoint
          }}>
          <EuiButtonEmpty iconType="bell" size="xs">
            <EuiText size="xs">Alert</EuiText>
          </EuiButtonEmpty>
        </Link>
      </EuiFlexItem>
    ),

    stopServing: (
      <EuiFlexItem grow={false} key={`stop-serving-${modelEndpoint.id}`}>
        <EuiButtonEmpty
          onClick={() => toggleStopServeModelEndpointModal()}
          color="danger"
          iconType="minusInCircle"
          size="xs">
          <EuiText size="xs">Stop serving</EuiText>
        </EuiButtonEmpty>
      </EuiFlexItem>
    )
  };

  const tailorActions = actions => {
    let list = [];

    if (featureToggleConfig.alertEnabled) {
      // Alert
      list.push(actions.alert);
    }

    // Stop serving
    list.push(actions.stopServing);

    return list;
  };

  return (
    <Fragment>
      <EuiFlexGroup
        alignItems="flexStart"
        direction="column"
        gutterSize="s"
        style={{ marginTop: "4px" }}>
        {tailorActions(actions)}
      </EuiFlexGroup>

      {isStopServeModelEndpointModalVisible && (
        <StopServeModelEndpointModal
          model={model}
          endpoint={modelEndpoint}
          callback={fetchModels}
          closeModal={() => toggleStopServeModelEndpointModal()}
        />
      )}
    </Fragment>
  );
};

export default ModelEndpointActions;
