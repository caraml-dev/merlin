import React, { useEffect } from "react";
import {
  EuiLoadingContent,
  EuiPage,
  EuiPageBody,
  EuiPageSection,
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

const DeployModelVersionView = ({
  projectId,
  modelId,
  versionId,
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
        { text: `Deploy` }
      ]);
    }
  }, [model, version]);

  const [submissionResponse, submitForm] = useMerlinApi(
    `/models/${model.id}/versions/${version.id}/endpoint`,
    { method: "POST" },
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
                  {"Deploy "}
                  {model.name}
                  {" version "}
                  <strong>{version.id}</strong>
                </>
              }
            />
          </EuiPageHeaderSection>
        </EuiPageHeader>

        <EuiPageSection>
          {modelLoaded && versionLoaded ? (
            <FormContextProvider data={new VersionEndpoint()}>
              <DeployModelVersionForm
                model={model}
                version={Version.fromJson(version)}
                onCancel={() => window.history.back()}
                onSuccess={redirectUrl => props.navigate(redirectUrl)}
                submissionResponse={submissionResponse}
                submitForm={submitForm}
                actionTitle="Deploy"
              />
            </FormContextProvider>
          ) : (
            <EuiLoadingContent lines={3} />
          )}
        </EuiPageSection>
      </EuiPageBody>
    </EuiPage>
  );
};

export default DeployModelVersionView;
