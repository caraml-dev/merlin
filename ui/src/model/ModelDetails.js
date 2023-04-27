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
import { Route, Routes, useLocation, useParams } from "react-router-dom";
import {
  EuiLoadingContent,
  EuiPageTemplate,
  EuiSpacer
} from "@elastic/eui";
import { get } from "@caraml-dev/ui-lib";
import { useMerlinApi } from "../hooks/useMerlinApi";
import { ModelAlert } from "./alert/ModelAlert";
import { featureToggleConfig } from "../config";

const LoadingContent = () => (
  <EuiPageTemplate.Section>
    <EuiLoadingContent lines={3} />
  </EuiPageTemplate.Section>
);

export const ModelDetails = () => {
  const { projectId, modelId } = useParams();
  const location = useLocation();
  const [model, setModel] = useState(get(location.state, "model"));
  const [breadcrumbs, setBreadcrumbs] = useState([]);

  const [{ data: models, isLoaded: modelsLoaded }] = useMerlinApi(
    `/projects/${projectId}/models`,
    {},
    [],
    !model
  );

  useEffect(() => {
    modelsLoaded && setModel(models.find(m => m.id.toString() === modelId));
  }, [models, modelsLoaded, modelId, setModel]);

  useEffect(() => {
    model &&
      setBreadcrumbs([
        {
          text: "Models",
          href: `/merlin/projects/${projectId}/models`
        },
        {
          text: model.name,
          href: `/merlin/projects/${projectId}/models/${model.id}`
        }
      ]);
  }, [projectId, model]);

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
            {model && (<Route path="endpoints/:endpointId/alert" element={<ModelAlert
                path="endpoints/:endpointId/alert"
                breadcrumbs={breadcrumbs}
                model={model}
              />} />)}
          </Routes>
        )}
       
      </EuiPageTemplate.Section>
      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};
