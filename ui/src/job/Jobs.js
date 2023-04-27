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

import React, { useEffect } from "react";
import {
  EuiButton,
  EuiPageTemplate,
  EuiPanel,
  EuiSpacer
} from "@elastic/eui";
import { useParams } from "react-router-dom";
import { replaceBreadcrumbs } from "@caraml-dev/ui-lib";
import { useMerlinApi } from "../hooks/useMerlinApi";
import mocks from "../mocks";
import JobListTable from "../job/JobListTable";

const Jobs = () => {
  const { projectId, modelId } = useParams();
  const createJobURL = `/merlin/projects/${projectId}/models/${modelId}/create-job`;

  const [{ data, isLoaded, error }] = useMerlinApi(
    `/projects/${projectId}/jobs?model_id=${modelId}`,
    { mock: mocks.jobList },
    []
  );

  const [model] = useMerlinApi(
    `/projects/${projectId}/models/${modelId}`,
    { mock: mocks.model },
    []
  );

  useEffect(() => {
    model.data.name &&
      replaceBreadcrumbs([
        {
          text: "Models",
          href: `/merlin/projects/${projectId}/models`
        },
        {
          text: model.data.name,
          href: `/merlin/projects/${projectId}/models/${modelId}`
        },
        { text: "Jobs" }
      ]);
  }, [projectId, modelId, model.data.name]);

  return (
    <EuiPageTemplate restrictWidth="90%" paddingSize="none">
      <EuiSpacer size="l" />
      <EuiPageTemplate.Header
        bottomBorder={false}
        iconType={"machineLearningApp"}
        pageTitle="Prediction Jobs"
        rightSideItems={[
          <EuiButton fill href={createJobURL}>
            Start Batch Job
          </EuiButton>,
        ]}
      />
      
      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
        <EuiPanel>
          <JobListTable
            projectId={projectId}
            modelId={modelId}
            jobs={data}
            isLoaded={isLoaded}
            error={error}
          />
        </EuiPanel>
      </EuiPageTemplate.Section>
      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};

export default Jobs;
