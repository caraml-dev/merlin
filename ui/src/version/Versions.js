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

import { get, replaceBreadcrumbs, useToggle } from "@caraml-dev/ui-lib";
import { EuiButton, EuiPageTemplate, EuiPanel, EuiSpacer } from "@elastic/eui";
import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { CursorPagination } from "../components/CursorPagination";
import {
  ServeVersionEndpointModal,
  UndeployVersionEndpointModal,
} from "../components/modals";
import { useMerlinApi } from "../hooks/useMerlinApi";
import mocks from "../mocks";
import VersionListTable from "./VersionListTable";

const Versions = ({ ...props }) => {
  const { projectId, modelId } = useParams();
  const [isUndeployEndpointModalVisible, toggleUndeployEndpointModal] =
    useToggle();
  const [isServeEndpointModalVisible, toggleServeEndpointModal] = useToggle();

  const [activeVersionEndpoint, setActiveVersionEndpoint] = useState(null);
  const [activeVersion, setActiveVersion] = useState(null);
  const [activeModel, setActiveModel] = useState(
    get(props, "location.state.activeModel"),
  );

  /**
   * API Usage
   */
  const limitPerPage = 50;
  const [limit, setLimit] = useState(limitPerPage);
  const [searchQuery, setSearchQuery] = useState(null);

  const [selectedCursor, setSelectedCursor] = useState("");
  const [versions, fetchVersions] = useMerlinApi(
    `/models/${modelId}/versions`,
    {
      mock: mocks.versionList,
      query: { limit: limit, search: searchQuery, cursor: selectedCursor },
      method: "GET",
    },
    [],
  );

  const [environments] = useMerlinApi(`/environments`, {}, []);

  const [models, fetchModels] = useMerlinApi(
    `/projects/${projectId}/models`,
    { mock: mocks.modelList },
    [],
  );

  const listJobURL = `/merlin/projects/${projectId}/models/${modelId}/versions/all/jobs`;

  useEffect(() => {
    if (activeModel) {
      replaceBreadcrumbs([
        {
          text: "Models",
          href: `/merlin/projects/${projectId}/models`,
        },
        {
          text: activeModel.name,
          href: `/merlin/projects/${projectId}/models/${activeModel.id}`,
        },
        { text: "Versions" },
      ]);
    } else {
      fetchModels();
    }
  }, [activeModel, projectId, fetchModels]);

  useEffect(() => {
    const found = models.data.filter((m) => m.id.toString() === modelId);

    if (found.length > 0) {
      setActiveModel(found[0]);
    }
  }, [models.data, modelId]);

  return (
    <EuiPageTemplate restrictWidth="90%" paddingSize="none">
      <EuiSpacer size="l" />
      <EuiPageTemplate.Header
        bottomBorder={false}
        iconType={"machineLearningApp"}
        pageTitle={activeModel ? activeModel.name : ""}
        rightSideItems={[
          activeModel && activeModel.type === "pyfunc_v2" && (
            <EuiButton href={listJobURL}>View Batch Jobs</EuiButton>
          ),
        ]}
      />

      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
        <EuiPanel>
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
            searchCallback={setSearchQuery}
            searchQuery={searchQuery}
            environments={environments.data}
          />
          <EuiSpacer size="m" />
          {activeModel && (
            <CursorPagination
              apiResponse={versions}
              defaultLimit={limitPerPage}
              search={searchQuery}
              limit={limit}
              setLimit={setLimit}
              setSelectedCursor={setSelectedCursor}
            />
          )}
        </EuiPanel>
      </EuiPageTemplate.Section>
      {isUndeployEndpointModalVisible && (
        <UndeployVersionEndpointModal
          versionEndpoint={activeVersionEndpoint}
          version={activeVersion}
          model={activeModel}
          callback={fetchVersions}
          closeModal={toggleUndeployEndpointModal}
        />
      )}

      {isServeEndpointModalVisible && (
        <ServeVersionEndpointModal
          versionEndpoint={activeVersionEndpoint}
          version={activeVersion}
          model={activeModel}
          callback={fetchVersions}
          closeModal={toggleServeEndpointModal}
        />
      )}
      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};

export default Versions;
