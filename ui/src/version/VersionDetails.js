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
  EuiBadge,
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
import { useMerlinApi } from "../hooks/useMerlinApi";
import { Router } from "@reach/router";
import { get } from "@gojek/mlp-ui";
import VersionDeploy from "./deployment/VersionDeploy";
import VersionRedeploy from "./deployment/VersionRedeploy";
import Log from "../log/Log";
import PropTypes from "prop-types";

export const VersionDetails = ({
  projectId,
  modelId,
  versionId,
  location: { state }
}) => {
  const [model, setModel] = useState(get(state, "model"));
  const [version, setVersion] = useState(get(state, "version"));

  const [breadcrumbs, setBreadcrumbs] = useState([]);

  const [{ data: models, isLoaded: modelsLoaded }] = useMerlinApi(
    `/projects/${projectId}/models`,
    {},
    [],
    !model
  );

  const [{ data: versions, isLoaded: versionsLoaded }] = useMerlinApi(
    `/models/${modelId}/versions`,
    {},
    [],
    !version
  );

  useEffect(() => {
    modelsLoaded && setModel(models.find(m => m.id.toString() === modelId));
  }, [models, modelsLoaded, modelId, setModel]);

  useEffect(() => {
    versionsLoaded &&
      setVersion(versions.find(v => v.id.toString() === versionId));
  }, [versions, versionsLoaded, versionId, setVersion]);

  useEffect(() => {
    model &&
      version &&
      setBreadcrumbs([
        {
          text: "Models",
          href: `/merlin/projects/${projectId}/models`
        },
        {
          text: model.name,
          href: `/merlin/projects/${projectId}/models/${model.id}`
        },
        {
          text: `Model Version ${version.id}`,
          href: `/merlin/projects/${projectId}/models/${model.id}/versions/${version.id}`
        }
      ]);
  }, [projectId, model, version]);

  return (
    <EuiPage>
      <EuiPageBody>
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <EuiFlexGroup alignItems="center" gutterSize="s">
              <EuiFlexItem grow={false}>
                <EuiIcon type="graphApp" size="xl" />
              </EuiFlexItem>

              {version && (
                <EuiFlexItem grow={2}>
                  <EuiTitle size="m">
                    <h1>
                      Model Version <strong>{version.id}</strong>
                    </h1>
                  </EuiTitle>
                </EuiFlexItem>
              )}

              {model && (
                <EuiFlexItem grow={false} style={{ maxWidth: 300 }}>
                  {<EuiBadge color="secondary">{model.name}</EuiBadge>}
                </EuiFlexItem>
              )}
            </EuiFlexGroup>
          </EuiPageHeaderSection>
        </EuiPageHeader>

        <Router>
          {model && version && (
            <VersionDeploy
              path="deploy"
              model={model}
              version={version}
              breadcrumbs={breadcrumbs}
            />
          )}

          {model && version && (
            <VersionRedeploy
              path="endpoints/:endpointId/redeploy"
              model={model}
              version={version}
              breadcrumbs={breadcrumbs}
            />
          )}

          {model && version && (
            <VersionLog
              path="endpoints/:endpointId/logs"
              modelId={model.id}
              versionId={version.id}
              breadcrumbs={breadcrumbs}
            />
          )}

          <LoadingContent default />
        </Router>
      </EuiPageBody>
    </EuiPage>
  );
};

VersionDetails.propTypes = {
  projectId: PropTypes.string,
  modelId: PropTypes.string,
  versionId: PropTypes.string,
  location: PropTypes.object,
  state: PropTypes.object
};

const VersionLog = ({ modelId, versionId, endpointId, breadcrumbs }) => {
  const containerURL = `/models/${modelId}/versions/${versionId}/endpoint/${endpointId}/containers`;
  return (
    <Log
      path="endpoints/:endpointId/logs"
      modelId={modelId}
      fetchContainerURL={containerURL}
      breadcrumbs={breadcrumbs}
    />
  );
};

const LoadingContent = () => (
  <EuiPageContent>
    <EuiLoadingContent lines={3} />
  </EuiPageContent>
);
