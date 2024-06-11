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

import { EuiPageTemplate, EuiSkeletonText, EuiSpacer } from "@elastic/eui";
import React from "react";
import { Route, Routes, useParams } from "react-router-dom";
import { featureToggleConfig } from "../config";
import { useMerlinApi } from "../hooks/useMerlinApi";
import mocks from "../mocks";
import { ModelAlert } from "./alert/ModelAlert";

const LoadingContent = () => (
  <EuiPageTemplate.Section>
    <EuiSkeletonText lines={3} />
  </EuiPageTemplate.Section>
);

export const ModelDetails = () => {
  const { projectId, modelId } = useParams();

  const [{ data: model, isLoaded: modelLoaded }] = useMerlinApi(
    `/projects/${projectId}/models/${modelId}`,
    { mock: mocks.model },
    [],
  );

  const breadcrumbs = [
    {
      text: "Models",
      href: `/merlin/projects/${projectId}/models`,
    },
    {
      text: model.name,
      href: `/merlin/projects/${projectId}/models/${modelId}`,
    },
  ];

  return (
    <EuiPageTemplate restrictWidth="90%" paddingSize="none">
      <EuiSpacer size="l" />
      <EuiPageTemplate.Header
        bottomBorder={false}
        iconType={"machineLearningApp"}
        pageTitle={model.name}
      />

      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
        {featureToggleConfig.alertEnabled && (
          <Routes>
            <Route index element={<LoadingContent default />} />
            {modelLoaded && model && (
              <Route
                path="endpoints/:endpointId/alert"
                element={
                  <ModelAlert
                    path="endpoints/:endpointId/alert"
                    breadcrumbs={breadcrumbs}
                    model={model}
                  />
                }
              />
            )}
          </Routes>
        )}
      </EuiPageTemplate.Section>
      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};
