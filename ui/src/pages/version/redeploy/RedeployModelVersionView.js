import React, { useEffect } from "react";
import {
  EuiLoadingContent,
  EuiPage,
  EuiPageBody,
  EuiPageContentBody,
  EuiPageHeader,
  EuiPageHeaderSection
} from "@elastic/eui";
import { FormContextProvider, replaceBreadcrumbs } from "@gojek/mlp-ui";
import { PageTitle } from "../../../components/PageTitle";
import { useMerlinApi } from "../../../hooks/useMerlinApi";
import mocks from "../../../mocks";
import { Version } from "../../../services/version/Version";
import { VersionEndpoint } from "../../../services/version_endpoint/VersionEndpoint";

import { DeployModelVersionForm } from "../components/forms/DeployModelVersionForm";

const RedeployModelVersionView = ({
  projectId,
  modelId,
  versionId,
  endpointId,
  ...props
}) => {
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

  useEffect(() => {
    if (model && version) {
      replaceBreadcrumbs([
        { text: `Models`, href: `/merlin/projects/${model.project_id}/models` },
        {
          text: `${model.name}`,
          href: `/merlin/projects/${model.project_id}/models/${model.id}`
        },
        {
          text: `Version ${version.id}`,
          href: `/merlin/projects/${model.project_id}/models/${model.id}/versions/${version.id}`
        },
        { text: `Redeploy` }
      ]);
    }
  }, [model, version]);

  const [submissionResponse, submitForm] = useMerlinApi(
    `/models/${model.id}/versions/${version.id}/endpoint/${endpointId}`,
    { method: "PUT" },
    {},
    false
  );

  return (
    <EuiPage>
      <EuiPageBody>
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <PageTitle
              title={
                <>
                  {"Redeploy "}
                  {model.name}
                  {" version "}
                  <strong>{version.id}</strong>
                </>
              }
            />
          </EuiPageHeaderSection>
        </EuiPageHeader>

        <EuiPageContentBody>
          {modelLoaded && versionLoaded ? (
            <FormContextProvider
              data={VersionEndpoint.fromJson(
                version.endpoints.find(e => e.id === endpointId)
              )}>
              <DeployModelVersionForm
                model={model}
                version={Version.fromJson(version)}
                onCancel={() => window.history.back()}
                onSuccess={redirectUrl => props.navigate(redirectUrl)}
                submissionResponse={submissionResponse}
                submitForm={submitForm}
                actionTitle="Redeploy"
                isEnvironmentDisabled={true}
              />
            </FormContextProvider>
          ) : (
            <EuiLoadingContent lines={3} />
          )}
        </EuiPageContentBody>
      </EuiPageBody>
    </EuiPage>
  );
};

export default RedeployModelVersionView;
