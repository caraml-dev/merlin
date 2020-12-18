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

import React, { useEffect, useState } from "react";
import {
  EuiIcon,
  EuiPage,
  EuiPageBody,
  EuiPageContent,
  EuiPageHeader,
  EuiPageHeaderSection,
  EuiTitle
} from "@elastic/eui";
import VersionListTable from "./VersionListTable";
import { get, replaceBreadcrumbs, useToggle } from "@gojek/mlp-ui";
import VersionUndeployEndpointModal from "./modals/VersionUndeployEndpointModal";
import VersionServeEndpointModal from "./modals/VersionServeEndpointModal";
import mocks from "../mocks";
import { useMerlinApi } from "../hooks/useMerlinApi";
import PropTypes from "prop-types";

const Versions = ({ projectId, modelId, ...props }) => {
  const [
    isUndeployEndpointModalVisible,
    toggleUndeployEndpointModal
  ] = useToggle();
  const [isServeEndpointModalVisible, toggleServeEndpointModal] = useToggle();

  const [activeVersionEndpoint, setActiveVersionEndpoint] = useState(null);
  const [activeVersion, setActiveVersion] = useState(null);
  const [activeModel, setActiveModel] = useState(
    get(props, "location.state.activeModel")
  );

  /**
   * API Usage
   */
  const [versions, fetchVersions] = useMerlinApi(
    `/models/${modelId}/versions`,
    { mock: mocks.versionList },
    []
  );

  const [models, fetchModels] = useMerlinApi(
    `/projects/${projectId}/models`,
    { mock: mocks.modelList },
    []
  );

  useEffect(() => {
    if (activeModel) {
      replaceBreadcrumbs([
        {
          text: "Models",
          href: `/merlin/projects/${projectId}/models`
        },
        {
          text: activeModel.name,
          href: `/merlin/projects/${projectId}/models/${activeModel.id}`
        },
        { text: "Versions" }
      ]);
    } else {
      fetchModels();
    }
  }, [activeModel, projectId, fetchModels]);

  useEffect(() => {
    const found = models.data.filter(m => m.id.toString() === modelId);

    if (found.length > 0) {
      setActiveModel(found[0]);
    }
  }, [models.data, modelId]);

  return (
    <EuiPage>
      <EuiPageBody>
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <EuiTitle size="l">
              <h1>
                <EuiIcon type="graphApp" size="xl" />{" "}
                {activeModel ? activeModel.name : ""}
              </h1>
            </EuiTitle>
          </EuiPageHeaderSection>
        </EuiPageHeader>

        <EuiPageContent>
          <VersionListTable
            projectId={projectId}
            versions={versions.data}
            fetchVersions={fetchVersions}
            isLoaded={versions.isLoaded}
            error={versions.error}
            activeModel={activeModel}
            setActiveModel={setActiveModel}
            activeVersion={activeVersion}
            setActiveVersion={setActiveVersion}
            activeVersionEndpoint={activeVersionEndpoint}
            setActiveVersionEndpoint={setActiveVersionEndpoint}
            toggleUndeployEndpointModal={toggleUndeployEndpointModal}
            toggleServeEndpointModal={toggleServeEndpointModal}
          />
        </EuiPageContent>
      </EuiPageBody>

      {isUndeployEndpointModalVisible && (
        <VersionUndeployEndpointModal
          versionEndpoint={activeVersionEndpoint}
          version={activeVersion}
          model={activeModel}
          updateVersionsCallback={fetchVersions}
          closeModal={toggleUndeployEndpointModal}
        />
      )}

      {isServeEndpointModalVisible && (
        <VersionServeEndpointModal
          versionEndpoint={activeVersionEndpoint}
          version={activeVersion}
          model={activeModel}
          updateVersionsCallback={fetchVersions}
          closeModal={toggleServeEndpointModal}
        />
      )}
    </EuiPage>
  );
};

Versions.propTypes = {
  projectId: PropTypes.string.isRequired,
  modelId: PropTypes.string.isRequired
};

export default Versions;
