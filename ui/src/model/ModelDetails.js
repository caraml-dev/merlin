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
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiIcon,
  EuiLoadingContent,
  EuiPage,
  EuiPageBody,
  EuiPageContent,
  EuiPageHeader,
  EuiPageHeaderSection,
  EuiTitle
} from "@elastic/eui";
import { Router } from "@reach/router";
import { get } from "@gojek/mlp-ui";
import { useMerlinApi } from "../hooks/useMerlinApi";
import { ModelAlert } from "./alert/ModelAlert";
import { featureToggleConfig } from "../config";
import PropTypes from "prop-types";

const LoadingContent = () => (
  <EuiPageContent>
    <EuiLoadingContent lines={3} />
  </EuiPageContent>
);

export const ModelDetails = ({ projectId, modelId, location: { state } }) => {
  const [model, setModel] = useState(get(state, "model"));
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
    <EuiPage>
      <EuiPageBody>
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <EuiFlexGroup alignItems="center" gutterSize="s">
              <EuiFlexItem grow={false}>
                <EuiIcon type="graphApp" size="xl" />
              </EuiFlexItem>

              {model && (
                <EuiFlexItem grow={2}>
                  <EuiTitle size="l">
                    <h1>{model.name}</h1>
                  </EuiTitle>
                </EuiFlexItem>
              )}
            </EuiFlexGroup>
          </EuiPageHeaderSection>
        </EuiPageHeader>
        {featureToggleConfig.alertEnabled && (
          <Router>
            {model && (
              <ModelAlert
                path="endpoints/:endpointId/alert"
                breadcrumbs={breadcrumbs}
                model={model}
              />
            )}

            <LoadingContent default />
          </Router>
        )}
      </EuiPageBody>
    </EuiPage>
  );
};

ModelDetails.propTypes = {
  projectId: PropTypes.string,
  modelId: PropTypes.string,
  state: PropTypes.object
};
