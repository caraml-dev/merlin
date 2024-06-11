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

import { replaceBreadcrumbs } from "@caraml-dev/ui-lib";
import {
  EuiButton,
  EuiEmptyPrompt,
  EuiFlexGroup,
  EuiFlexItem,
  EuiPageTemplate,
  EuiSkeletonText,
  EuiSpacer,
  EuiText,
} from "@elastic/eui";
import React, { Fragment, useEffect, useState } from "react";
import { Link, Route, Routes, useParams } from "react-router-dom";
import { ContainerLogsView } from "../../components/logs/ContainerLogsView";
import { ConfigSection, ConfigSectionPanel } from "../../components/section";
import config, { appConfig } from "../../config";
import { useMerlinApi } from "../../hooks/useMerlinApi";
import mocks from "../../mocks";
import JobsTable from "../job/components/JobsTable";
import { DeploymentPanelHeader } from "./DeploymentPanelHeader";
import { EndpointDetails } from "./EndpointDetails";
import { HistoryDetails } from "./HistoryDetails";
import { ModelVersionPanelHeader } from "./ModelVersionPanelHeader";
import { VersionTabNavigation } from "./VersionTabNavigation";

/**
 * VersionDetails page containing detailed information of a model version.
 * In this page users can also manage all deployed endpoint created from the model version.
 */
const VersionDetails = () => {
  const {
    projectId,
    modelId,
    versionId,
    endpointId,
    "*": section,
  } = useParams();
  const [{ data: model, isLoaded: modelLoaded }] = useMerlinApi(
    `/projects/${projectId}/models/${modelId}`,
    { mock: mocks.model },
    {},
  );

  const [{ data: version, isLoaded: versionLoaded }] = useMerlinApi(
    `/models/${modelId}/versions/${versionId}`,
    { mock: mocks.versionList[0] },
    {},
  );

  const [jobPage, setJobPage] = useState({
    index: 0,
    size: appConfig.pagination.defaultPageSize,
  });

  const [searchJobText, setSearchJobText] = useState(versionId);

  const [{ data: jobs, isLoaded: jobsLoaded, error: jobsError }, fetchJobs] =
    useMerlinApi(
      `/projects/${projectId}/jobs-by-page?model_id=${modelId}&version_id=${versionId}&page=${jobPage.index + 1}&page_size=${jobPage.size}&search=${searchJobText}`,
      { mock: mocks.jobList },
      [],
    );

  const onSearchJobTextChange = (text) => {
    setJobPage({ ...jobPage, index: 0 });
    setSearchJobText(text);
  };

  const [endpoint, setEndpoint] = useState();
  const [environments, setEnvironments] = useState([]);
  const [isDeployed, setIsDeployed] = useState(false);

  useEffect(() => {
    if (version) {
      if (version.endpoints && version.endpoints.length > 0) {
        setIsDeployed(true);
        setEnvironments(
          version.endpoints.map((endpoint) => endpoint.environment),
        );

        if (endpointId) {
          setEndpoint(
            version.endpoints.find((endpoint) => endpoint.id === endpointId),
          );
        }
      }
    }
  }, [version, endpointId]);

  useEffect(() => {
    let breadCrumbs = [];

    if (modelLoaded && versionLoaded) {
      breadCrumbs.push(
        {
          text: "Models",
          href: `/merlin/projects/${model.project_id}/models`,
        },
        {
          text: model.name || "",
          href: `/merlin/projects/${model.project_id}/models/${model.id}`,
        },
        {
          text: `Model Version ${version.id}`,
          href: `/merlin/projects/${model.project_id}/models/${model.id}/versions/${version.id}`,
        },
      );
    }

    if (endpoint) {
      breadCrumbs.push({
        text: endpoint.environment_name,
      });
    }

    replaceBreadcrumbs(breadCrumbs);
  }, [modelLoaded, model, versionLoaded, version, endpoint]);

  return (
    <EuiPageTemplate restrictWidth="90%" paddingSize="none">
      <EuiSpacer size="l" />
      {!modelLoaded && !versionLoaded ? (
        <EuiFlexGroup direction="row">
          <EuiFlexItem grow={true}>
            <EuiSkeletonText lines={3} />
          </EuiFlexItem>
        </EuiFlexGroup>
      ) : (
        <Fragment>
          <EuiPageTemplate.Header
            bottomBorder={false}
            iconType={"machineLearningApp"}
            pageTitle={
              <Fragment>
                {model.name}
                {" version "}
                <strong>{version.id}</strong>
              </Fragment>
            }
          />

          <EuiPageTemplate.Section color={"transparent"}>
            <EuiSpacer size="l" />
            {!(section === "deploy" || section === "redeploy") &&
              model &&
              modelLoaded &&
              version &&
              versionLoaded && (
                <Fragment>
                  <ModelVersionPanelHeader model={model} version={version} />
                  <EuiSpacer size="m" />
                </Fragment>
              )}

            {!(section === "deploy" || section === "redeploy") &&
              model &&
              modelLoaded &&
              version &&
              versionLoaded &&
              environments &&
              isDeployed && (
                <Fragment>
                  <DeploymentPanelHeader
                    model={model}
                    version={version}
                    endpoint={endpoint}
                    environments={environments}
                  />
                  <EuiSpacer size="m" />
                </Fragment>
              )}

            {model.type === "pyfunc_v2" && (
              <Fragment>
                <EuiSpacer size="m" />
                <ConfigSection title="Batch Prediction Jobs">
                  <ConfigSectionPanel>
                    <JobsTable
                      projectId={projectId}
                      modelId={modelId}
                      jobs={jobs.results || []}
                      isLoaded={jobsLoaded}
                      error={jobsError}
                      page={jobPage}
                      totalItemCount={jobs?.paging?.total || 0}
                      onPaginationChange={setJobPage}
                      fetchJobs={fetchJobs}
                    />
                  </ConfigSectionPanel>
                </ConfigSection>
              </Fragment>
            )}

            {!(section === "deploy" || section === "redeploy") &&
              endpoint &&
              isDeployed && (
                <Fragment>
                  <VersionTabNavigation
                    endpoint={endpoint}
                    selectedTab={section}
                  />
                  <EuiSpacer size="m" />
                </Fragment>
              )}

            {!(section === "deploy" || section === "redeploy") &&
              model &&
              modelLoaded &&
              version &&
              versionLoaded &&
              !isDeployed &&
              model.type !== "pyfunc_v2" && (
                <EuiEmptyPrompt
                  title={<h2>Model version is not deployed</h2>}
                  body={
                    <Fragment>
                      <p>
                        Deploy it first and wait for deployment to complete
                        before you can see the configuration details here.
                      </p>
                      <Link
                        to={`${config.HOMEPAGE}/projects/${model.project_id}/models/${model.id}/versions/${version.id}/deploy`}
                        state={{ model: model, version: version }}
                      >
                        <EuiButton iconType="importAction" size="s">
                          <EuiText size="xs">Deploy</EuiText>
                        </EuiButton>
                      </Link>
                    </Fragment>
                  }
                />
              )}

            {model && modelLoaded && version && versionLoaded && endpoint && (
              <Routes>
                <Route
                  path="details"
                  element={
                    <EndpointDetails
                      model={model}
                      version={version}
                      endpoint={endpoint}
                    />
                  }
                />
                <Route
                  path="history"
                  element={
                    <HistoryDetails
                      model={model}
                      version={version}
                      endpoint={endpoint}
                    />
                  }
                />
                <Route
                  path="logs"
                  element={
                    <ContainerLogsView
                      projectId={projectId}
                      model={model}
                      versionId={versionId}
                      revisionId={endpoint.revision_id}
                      fetchContainerURL={`/models/${modelId}/versions/${versionId}/endpoint/${endpointId}/containers`}
                    />
                  }
                />
              </Routes>
            )}
          </EuiPageTemplate.Section>
        </Fragment>
      )}
      <EuiSpacer size="xxl" />
    </EuiPageTemplate>
  );
};

export default VersionDetails;
