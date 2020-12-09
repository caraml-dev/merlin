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

import React, { Fragment, useEffect, useState } from "react";
import { Redirect, Router } from "@reach/router";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiPage,
  EuiPageBody,
  EuiPageHeader,
  EuiPageHeaderSection,
  EuiSpacer
} from "@elastic/eui";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { PageTitle } from "../../components/PageTitle";
import { useMerlinApi } from "../../hooks/useMerlinApi";
import Log from "../../log/Log";
import mocks from "../../mocks";
import VersionRedeploy from "../../version/deployment/VersionRedeploy";
import { ConfigSection, ConfigSectionPanel } from "../../components/section";
import { ModelVersionConfigTable } from "../../components/ModelVersionConfigTable";
import { VersionConfig } from "./VersionConfig";
import { DeploymentInfoHeader } from "./DeploymentInfoHeader";
import { VersionEndpointTabNavigation } from "./VersionEndpointTabNavigation";

const Version = ({ projectId, modelId, versionId, endpointId, ...props }) => {
  const [{ data: version, isLoaded: versionLoaded }] = useMerlinApi(
    `/models/${modelId}/versions/${versionId}`,
    { mock: mocks.versionList[0] },
    {}
  );

  const [{ data: endpoint, isLoaded: endpointLoaded }] = useMerlinApi(
    `/models/${modelId}/versions/${versionId}/endpoint/${endpointId}`,
    { mock: mocks.versionEndpoint },
    {},
    endpointId !== "-"
  );

  const [environments, setEnvironments] = useState([]);
  useEffect(() => {
    version.endpoints &&
      setEnvironments(version.endpoints.map(endpoint => endpoint.environment));
  }, [version]);

  useEffect(() => {
    let breadCrumbs = [
      {
        text: "Models",
        href: `/merlin/projects/${projectId}/models`
      }
    ];

    if (version.model) {
      breadCrumbs.push({
        text: version.model.name,
        href: `/merlin/projects/${projectId}/models/${modelId}`
      });
      breadCrumbs.push({
        text: `Model Version ${versionId}`,
        href: `/merlin/projects/${projectId}/models/${modelId}/versions/${versionId}`
      });
    }

    if (endpoint.environment_name) {
      breadCrumbs.push({
        text: endpoint.environment_name
      });
    }

    replaceBreadcrumbs(breadCrumbs);
  }, [projectId, modelId, versionId, version, endpoint]);

  return (
    <EuiPage>
      <EuiPageBody>
        <Fragment>
          <EuiPageHeader>
            <EuiPageHeaderSection>
              <PageTitle
                title={
                  <Fragment>
                    {version.model && version.model.name}
                    {" version "}
                    <strong>{version.id}</strong>
                  </Fragment>
                }
              />
            </EuiPageHeaderSection>
          </EuiPageHeader>

          {versionLoaded && !(props["*"] === "redeploy") && (
            <Fragment>
              <ConfigSection title="Model">
                <ConfigSectionPanel>
                  <EuiFlexGroup direction="column" gutterSize="m">
                    <EuiFlexItem>
                      <ModelVersionConfigTable
                        model={version.model}
                        version={version}
                      />
                    </EuiFlexItem>
                  </EuiFlexGroup>
                </ConfigSectionPanel>
              </ConfigSection>
              <EuiSpacer size="m" />
            </Fragment>
          )}

          {versionLoaded &&
            endpointLoaded &&
            environments &&
            !(props["*"] === "redeploy") && (
              <Fragment>
                <ConfigSection title="Deployment">
                  <DeploymentInfoHeader
                    version={version}
                    endpoint={endpoint}
                    environments={environments}
                  />
                </ConfigSection>
                <EuiSpacer size="m" />
              </Fragment>
            )}

          {endpointLoaded && !(props["*"] === "redeploy") && (
            <Fragment>
              <VersionEndpointTabNavigation endpoint={endpoint} {...props} />
              <EuiSpacer size="m" />
            </Fragment>
          )}

          <Router primary={false}>
            <Redirect from="/" to="details" noThrow />
            {version && version.model && (
              <VersionConfig
                path="details"
                model={version.model}
                version={version}
                endpoint={endpoint}
              />
            )}

            <Log
              path="logs"
              modelId={modelId}
              versionId={versionId}
              fetchContainerURL={`/models/${modelId}/versions/${versionId}/endpoint/${endpointId}/containers`}
            />

            {version && version.model && (
              <VersionRedeploy
                path="redeploy"
                model={version.model}
                version={version}
              />
            )}
          </Router>
        </Fragment>
      </EuiPageBody>
    </EuiPage>
  );
};

export default Version;
