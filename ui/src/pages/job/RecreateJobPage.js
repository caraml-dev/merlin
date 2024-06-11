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

import {
  EuiCallOut,
  EuiLoadingChart,
  EuiPageTemplate,
  EuiSpacer,
  EuiTextAlign,
} from "@elastic/eui";
import React from "react";
import { useParams } from "react-router-dom";
import { useMerlinApi } from "../../hooks/useMerlinApi";
import { Job } from "../../services/job/Job";
import { JobForm } from "./form/JobForm";
import { JobFormContextProvider } from "./form/context";

const RecreateJobPage = () => {
  const { projectId, modelId, versionId, jobId } = useParams();

  const [{ data, isLoaded, error }] = useMerlinApi(
    `models/${modelId}/versions/${versionId}/jobs/${jobId}`,
    {},
    [],
  );

  return (
    <EuiPageTemplate restrictWidth="90%" paddingSize="none">
      <EuiSpacer size="l" />
      <EuiPageTemplate.Header
        bottomBorder={false}
        iconType={"machineLearningApp"}
        pageTitle="Prediction Jobs"
      />
      <EuiPageTemplate.Section color="transparent">
        {!isLoaded ? (
          <EuiTextAlign textAlign="center">
            <EuiSpacer size="xxl" />
            <EuiLoadingChart size="xl" mono />
          </EuiTextAlign>
        ) : error ? (
          <EuiCallOut
            title="Sorry, there was an error"
            color="danger"
            iconType="alert"
          >
            <p>{error.message}</p>
          </EuiCallOut>
        ) : (
          <JobFormContextProvider job={Job.from(data)}>
            <JobForm
              projectId={projectId}
              modelId={modelId}
              versionId={versionId}
              isNewJob={false}
            />
          </JobFormContextProvider>
        )}
      </EuiPageTemplate.Section>
    </EuiPageTemplate>
  );
};

export default RecreateJobPage;
