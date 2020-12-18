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

import React from "react";
import { useMerlinApi } from "../../hooks/useMerlinApi";
import { EndpointDeployment } from "./components/EndpointDeployment";
import PropTypes from "prop-types";

const VersionRedeploy = ({ breadcrumbs, model, version, endpointId }) => {
  const [redeploymentResponse, redeployVersionEndpoint] = useMerlinApi(
    `/models/${model.id}/versions/${version.id}/endpoint/${endpointId}`,
    { method: "PUT", addToast: true },
    {},
    false
  );

  return (
    <EndpointDeployment
      actionTitle="Redeploy"
      breadcrumbs={breadcrumbs}
      model={model}
      version={version}
      endpointId={endpointId}
      disableEnvironment={true}
      modalContent={`You're about to redeploy an existing endpoint of model version ${version.id}.`}
      onDeploy={redeployVersionEndpoint}
      response={redeploymentResponse}
    />
  );
};

VersionRedeploy.propTypes = {
  breadcrumbs: PropTypes.array,
  model: PropTypes.object,
  version: PropTypes.object,
  endpointId: PropTypes.string
};

export default VersionRedeploy;
