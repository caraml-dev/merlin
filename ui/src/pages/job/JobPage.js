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
  EuiCallOut,
  EuiFlexGroup,
  EuiFlexItem,
  EuiLoadingChart,
  EuiPageTemplate,
  EuiSpacer,
  EuiTextAlign,
} from "@elastic/eui";
import React, { Fragment, useEffect } from "react";
import { useParams } from "react-router-dom";
import { useMerlinApi } from "../../hooks/useMerlinApi";
import mocks from "../../mocks";
import JobConfig from "./components/JobConfig";
import { JobDetailTabNavigation } from "./components/JobDetailTabNavigation";
import JobErrorMessage from "./components/JobErrorMessage";
import { JobInfoPanel } from "./components/JobInfoPanel";
import JobLog from "./components/JobLog";
import { JobRunPanel } from "./components/JobRunPanel";
import { generateBreadcrumbs } from "./utils/breadcrumbs";

const JobPage = () => {
  const { projectId, modelId, versionId, jobId, "*": section } = useParams();

  const [{ data: project, isLoaded: projectLoaded }] = useMerlinApi(
    `/projects/${projectId}`,
    { mock: mocks.project },
    [],
  );

  const [{ data: model }] = useMerlinApi(
    `/projects/${projectId}/models/${modelId}`,
    { mock: mocks.model },
    [],
  );

  const [{ data: version }] = useMerlinApi(
    `/models/${modelId}/versions/${versionId}`,
    { mock: mocks.versionList[0] },
    [],
  );

  const [{ data: job, isLoaded: jobLoaded, error: fetchJobError }] =
    useMerlinApi(
      `/models/${modelId}/versions/${versionId}/jobs/${jobId}`,
      { mock: mocks.job },
      [],
    );

  useEffect(() => {
    replaceBreadcrumbs(
      generateBreadcrumbs(projectId, modelId, model, versionId, jobId),
    );
  }, [projectId, modelId, versionId, jobId, model]);

  return (
    <EuiPageTemplate restrictWidth="90%" paddingSize="none">
      <EuiSpacer size="l" />
      <EuiPageTemplate.Header
        bottomBorder={false}
        iconType={"machineLearningApp"}
        pageTitle={<Fragment>Prediction Job {jobId}</Fragment>}
      />

      <EuiPageTemplate.Section color={"transparent"}>
        <EuiSpacer size="l" />

        {projectLoaded && jobLoaded ? (
          <EuiFlexGroup direction="column" gutterSize="l">
            <EuiFlexItem>
              <JobInfoPanel model={model} version={version} job={job} />
            </EuiFlexItem>

            <EuiFlexItem>
              <JobRunPanel
                project={project}
                model={model}
                version={version}
                job={job}
              />
            </EuiFlexItem>

            <EuiFlexItem>
              <JobDetailTabNavigation
                project={project}
                model={model}
                job={job}
                selectedTab={section}
              />
            </EuiFlexItem>

            <EuiFlexItem>
              {section === "details" && <JobConfig job={job} />}

              {section === "error" && <JobErrorMessage error={job.error} />}

              {section === "logs" && (
                <JobLog
                  projectId={projectId}
                  model={model}
                  versionId={versionId}
                  jobId={jobId}
                />
              )}
            </EuiFlexItem>
          </EuiFlexGroup>
        ) : fetchJobError ? (
          <EuiCallOut
            title="Sorry, there was an error"
            color="danger"
            iconType="alert"
          >
            <p>{fetchJobError.message}</p>
          </EuiCallOut>
        ) : (
          <EuiTextAlign textAlign="center">
            <EuiLoadingChart size="xl" mono />
          </EuiTextAlign>
        )}
      </EuiPageTemplate.Section>
      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};

export default JobPage;
