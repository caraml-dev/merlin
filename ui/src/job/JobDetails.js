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
import {
  EuiIcon,
  EuiPage,
  EuiPageBody,
  EuiPageHeader,
  EuiPageHeaderSection,
  EuiTitle,
  EuiFlexItem,
  EuiFlexGroup
} from "@elastic/eui";
import { Router } from "@reach/router";
import JobConfig from "./JobConfig";
import RecreateJobView from "./RecreateJobView";
import mocks from "../mocks";
import { useMerlinApi } from "../hooks/useMerlinApi";
import PropTypes from "prop-types";
import { ContainerLogsView } from "../components/logs/ContainerLogsView";

const JobLog = ({ projectId, model, versionId, jobId }) => {
  const containerURL = `/models/${model.id}/versions/${versionId}/jobs/${jobId}/containers`;
  return (
    <ContainerLogsView
      projectId={projectId}
      model={model}
      versionId={versionId}
      jobId={jobId}
      fetchContainerURL={containerURL}
    />
  );
};

const JobDetails = ({ projectId, modelId, versionId, jobId }) => {
  const [{ data: model, isLoaded: modelLoaded }] = useMerlinApi(
    `/projects/${projectId}/models/${modelId}`,
    { mock: mocks.model },
    []
  );

  const breadcrumbs = [
    {
      text: "Models",
      href: `/merlin/projects/${projectId}/models`
    },
    {
      text: model && model.data ? model.data.name : "",
      href: `/merlin/projects/${projectId}/models/${modelId}`
    },
    {
      text: "Versions",
      href: `/merlin/projects/${projectId}/models/${modelId}/versions`
    },
    {
      text: versionId,
      href: `/merlin/projects/${projectId}/models/${modelId}/versions/${versionId}`
    },
    {
      text: "Jobs",
      href: `/merlin/projects/${projectId}/models/${modelId}/versions/${versionId}/jobs`
    },
    {
      text: `${jobId}`,
      href: `/merlin/projects/${projectId}/models/${modelId}/versions/${versionId}/jobs/${jobId}`
    }
  ];

  return (
    <EuiPage>
      <EuiPageBody>
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <EuiFlexGroup alignItems="center" gutterSize="s">
              <EuiFlexItem grow={false}>
                <EuiIcon type="graphApp" size="xl" />
              </EuiFlexItem>

              <EuiFlexItem grow={2}>
                <EuiTitle size="m">
                  <h1>
                    Prediction Job <strong>{jobId}</strong>
                  </h1>
                </EuiTitle>
              </EuiFlexItem>
            </EuiFlexGroup>
          </EuiPageHeaderSection>
        </EuiPageHeader>
        <Router>
          <JobConfig path="/" breadcrumbs={breadcrumbs} />

          {projectId && modelLoaded && model && (
            <JobLog
              path="logs"
              projectId={projectId}
              model={model}
              versionId={versionId}
              jobId={jobId}
            />
          )}

          <RecreateJobView path="recreate" breadcrumbs={breadcrumbs} />
        </Router>
      </EuiPageBody>
    </EuiPage>
  );
};

JobDetails.propTypes = {
  projectId: PropTypes.string,
  modelId: PropTypes.string,
  versionId: PropTypes.string,
  jobId: PropTypes.string
};

export default JobDetails;
