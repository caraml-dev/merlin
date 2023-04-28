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

import React, { Fragment, useEffect } from "react";
import {
  EuiPageTemplate,
  EuiSpacer
} from "@elastic/eui";
import { Route, Routes, useParams } from "react-router-dom";
import JobConfig from "./JobConfig";
import RecreateJobView from "./RecreateJobView";
import mocks from "../mocks";
import { useMerlinApi } from "../hooks/useMerlinApi";
import { ContainerLogsView } from "../components/logs/ContainerLogsView";
import { replaceBreadcrumbs } from "@caraml-dev/ui-lib";

const JobLog = ({ model, breadcrumbs }) => {
  const { projectId, versionId, jobId } = useParams();
  useEffect(() => {
    breadcrumbs && replaceBreadcrumbs([...breadcrumbs, { text: "Logs" }]);
  }, [breadcrumbs]);

  const containerURL = `/models/${model.id}/versions/${versionId}/jobs/${jobId}/containers`;

  return (
    <Fragment>
      <EuiSpacer size="l" />
      <ContainerLogsView
        projectId={projectId}
        model={model}
        versionId={versionId}
        jobId={jobId}
        fetchContainerURL={containerURL}
      />
    </Fragment>
  );
};

const JobDetails = () => {
  const { projectId, modelId, versionId, jobId } = useParams();
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
    <EuiPageTemplate restrictWidth="90%" paddingSize="none">
      <EuiSpacer size="l" />
      <EuiPageTemplate.Header
        bottomBorder={false}
        iconType={"machineLearningApp"}
        pageTitle={
          <Fragment>
            Prediction Job {jobId}
          </Fragment>
        }
      />
    
      <EuiPageTemplate.Section color={"transparent"}>
        <Routes>
          <Route
            index
            element={
              <JobConfig path="/" breadcrumbs={breadcrumbs} />
            }
          />
          {projectId && modelLoaded && model && (
            <Route
              path="logs"
              element={
                <JobLog
                  projectId={projectId}
                  model={model}
                  versionId={versionId}
                  jobId={jobId}
                  breadcrumbs={breadcrumbs}
                />
              }
            />
          )}
          <Route
            path="recreate"
            element={
              <RecreateJobView breadcrumbs={breadcrumbs} />
            }
          />
        </Routes>
      </EuiPageTemplate.Section>
      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};

export default JobDetails;
