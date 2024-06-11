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

import { EuiPageTemplate, EuiSpacer } from "@elastic/eui";
import PropTypes from "prop-types";
import React from "react";
import { useParams } from "react-router-dom";
import { Job } from "../../services/job/Job";
import { JobForm } from "./form/JobForm";
import { JobFormContextProvider } from "./form/context";

const CreateJobPage = () => {
  const { projectId, modelId, versionId } = useParams();
  return (
    <EuiPageTemplate restrictWidth="90%" paddingSize="none">
      <EuiSpacer size="l" />
      <EuiPageTemplate.Header
        bottomBorder={false}
        iconType={"machineLearningApp"}
        pageTitle="Prediction Jobs"
      />
      <EuiPageTemplate.Section color="transparent">
        <JobFormContextProvider job={new Job(projectId, modelId, versionId)}>
          <JobForm
            projectId={projectId}
            modelId={modelId}
            versionId={versionId}
          />
        </JobFormContextProvider>
      </EuiPageTemplate.Section>
    </EuiPageTemplate>
  );
};

CreateJobPage.propTypes = {
  projectId: PropTypes.string,
  modelId: PropTypes.string,
  versionId: PropTypes.string,
};

export default CreateJobPage;
