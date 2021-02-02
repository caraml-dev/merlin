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
import { Link, Router } from "@reach/router";
import {
  EuiButton,
  EuiEmptyPrompt,
  EuiFlexGroup,
  EuiFlexItem,
  EuiLoadingContent,
  EuiPage,
  EuiPageBody,
  EuiPageHeader,
  EuiPageHeaderSection,
  EuiSpacer,
  EuiText
} from "@elastic/eui";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import config from "../../config";
import { useMerlinApi } from "../../hooks/useMerlinApi";
import Log from "../../log/Log";
import mocks from "../../mocks";
import { PageTitle } from "../../components/PageTitle";
import VersionDeploy from "../../version/deployment/VersionDeploy";
import VersionRedeploy from "../../version/deployment/VersionRedeploy";
import { DeploymentPanelHeader } from "./DeploymentPanelHeader";
import { ModelVersionPanelHeader } from "./ModelVersionPanelHeader";
import { VersionConfig } from "./VersionConfig";
import { VersionTabNavigation } from "./VersionTabNavigation";

const Version = ({ projectId, modelId, versionId, endpointId, ...props }) => {
  const [{ data: model, isLoaded: modelLoaded }] = useMerlinApi(
    `/projects/${projectId}/models/${modelId}`,
    { mock: mocks.model },
    {}
  );

  const [{ data: version, isLoaded: versionLoaded }] = useMerlinApi(
    `/models/${modelId}/versions/${versionId}`,
    { mock: mocks.versionList[0] },
    {}
  );

  const [endpoint, setEndpoint] = useState();
  const [environments, setEnvironments] = useState([]);
  const [isDeployed, setIsDeployed] = useState(false);

  useEffect(() => {
    if (version) {
      if (version.endpoints && version.endpoints.length > 0) {
        setIsDeployed(true);
        setEnvironments(
          version.endpoints.map(endpoint => endpoint.environment)
        );

        if (endpointId) {
          setEndpoint(
            version.endpoints.find(endpoint => endpoint.id === endpointId)
          );
        }
      }
    }
  }, [version, endpointId]);

  useEffect(() => {
    let breadCrumbs = [];

    if (modelLoaded && versionLoaded) {
      breadCrumbs.push(
        {
          text: "Models",
          href: `/merlin/projects/${model.project_id}/models`
        },
        {
          text: model.name || "",
          href: `/merlin/projects/${model.project_id}/models/${model.id}`
        },
        {
          text: `Model Version ${version.id}`,
          href: `/merlin/projects/${model.project_id}/models/${model.id}/versions/${version.id}`
        }
      );
    }

    if (endpoint) {
      breadCrumbs.push({
        text: endpoint.environment_name
      });
    }

    replaceBreadcrumbs(breadCrumbs);
  }, [modelLoaded, model, versionLoaded, version, endpoint]);

  return (
    <EuiPage>
      <EuiPageBody>
        {!modelLoaded && !versionLoaded ? (
          <EuiFlexGroup direction="row">
            <EuiFlexItem grow={true}>
              <EuiLoadingContent lines={3} />
            </EuiFlexItem>
          </EuiFlexGroup>
        ) : (
          <Fragment>
            <EuiPageHeader>
              <EuiPageHeaderSection>
                <PageTitle
                  title={
                    <Fragment>
                      {model.name}
                      {" version "}
                      <strong>{version.id}</strong>
                    </Fragment>
                  }
                />
              </EuiPageHeaderSection>
            </EuiPageHeader>

            {!(props["*"] === "deploy" || props["*"] === "redeploy") &&
              model &&
              modelLoaded &&
              version &&
              versionLoaded && (
                <Fragment>
                  <ModelVersionPanelHeader model={model} version={version} />
                  <EuiSpacer size="m" />
                </Fragment>
              )}

            {!(props["*"] === "deploy" || props["*"] === "redeploy") &&
              model &&
              modelLoaded &&
              version &&
              versionLoaded &&
              environments &&
              isDeployed && (
                <Fragment>
                  <DeploymentPanelHeader
                    model={model}
                    version={version}
                    endpoint={endpoint}
                    environments={environments}
                  />
                  <EuiSpacer size="m" />
                </Fragment>
              )}

            {!(props["*"] === "deploy" || props["*"] === "redeploy") &&
              endpoint &&
              isDeployed && (
                <Fragment>
                  <VersionTabNavigation endpoint={endpoint} {...props} />
                  <EuiSpacer size="m" />
                </Fragment>
              )}

            {!(props["*"] === "deploy" || props["*"] === "redeploy") &&
              model &&
              modelLoaded &&
              version &&
              versionLoaded &&
              !isDeployed && (
                <EuiEmptyPrompt
                  title={<h2>Model version is not deployed</h2>}
                  body={
                    <Fragment>
                      <p>
                        Deploy it first and wait for deployment to complete
                        before you can see the configuration details here.
                      </p>
                      <Link
                        to={`${config.HOMEPAGE}/projects/${model.project_id}/models/${model.id}/versions/${version.id}/deploy`}
                        state={{ model: model, version: version }}>
                        <EuiButton iconType="importAction" size="s">
                          <EuiText size="xs">
                            {model.type !== "pyfunc_v2"
                              ? "Deploy"
                              : "Deploy Endpoint"}
                          </EuiText>
                        </EuiButton>
                      </Link>
                    </Fragment>
                  }
                />
              )}

            <Router primary={false}>
              {model && modelLoaded && version && versionLoaded && endpoint && (
                <VersionConfig
                  path="details"
                  model={model}
                  version={version}
                  endpoint={{
                    ...endpoint,
                    transformer: {
                      enabled: true,
                      image: "ghcr.io/gojek/merlin-transformer:latest",
                      resource_request: {
                        cpu_request: "500m",
                        memory_request: "500Mi",
                        min_replica: 1,
                        max_replica: 4
                      },
                      transformer_type: "standard",
                      env_vars: [
                        {
                          name: "TRANSFORMER_CONFIG",
                          value:
                            '{"transformerConfig":{"feast":[{"project":"project_1","entities":[{"name":"user_id_1","valueType":"STRING","jsonPath":"user_id_1"}],"features":[{"name":"feast_test_metrics:string_feature","valueType":"STRING","defaultValue":"1"}]},{"project":"project_2","entities":[{"name":"user_id_2","valueType":"STRING","jsonPath":"user_id_2"}],"features":[{"name":"feast_test_metrics:string_feature","valueType":"STRING","defaultValue":"2"}]}]}}'
                        }
                      ]
                    }
                  }}
                />
              )}

              <Log
                path="logs"
                modelId={modelId}
                versionId={versionId}
                fetchContainerURL={`/models/${modelId}/versions/${versionId}/endpoint/${endpointId}/containers`}
              />

              {model && modelLoaded && version && versionLoaded && (
                <VersionDeploy path="deploy" model={model} version={version} />
              )}

              {model && modelLoaded && version && versionLoaded && (
                <VersionRedeploy
                  path="redeploy"
                  model={model}
                  version={version}
                />
              )}
            </Router>
          </Fragment>
        )}
      </EuiPageBody>
    </EuiPage>
  );
};

export default Version;
