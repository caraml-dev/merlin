import { FormContextProvider, replaceBreadcrumbs } from "@caraml-dev/ui-lib";
import {
  EuiPageTemplate,
  EuiPanel,
  EuiSkeletonText,
  EuiSpacer,
} from "@elastic/eui";
import React, { Fragment, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useMerlinApi } from "../../../hooks/useMerlinApi";
import mocks from "../../../mocks";
import { Version } from "../../../services/version/Version";
import { VersionEndpoint } from "../../../services/version_endpoint/VersionEndpoint";

import { DeployModelVersionForm } from "../components/forms/DeployModelVersionForm";

const RedeployModelVersionView = () => {
  const navigate = useNavigate();
  const { projectId, modelId, versionId, endpointId } = useParams();
  const [{ data: model, isLoaded: modelLoaded }] = useMerlinApi(
    `/projects/${projectId}/models/${modelId}`,
    { mock: mocks.model },
    {},
  );

  const [{ data: version, isLoaded: versionLoaded }] = useMerlinApi(
    `/models/${modelId}/versions/${versionId}`,
    { mock: mocks.versionList[0] },
    {},
  );

  useEffect(() => {
    if (model && version) {
      replaceBreadcrumbs([
        { text: `Models`, href: `/merlin/projects/${model.project_id}/models` },
        {
          text: `${model.name}`,
          href: `/merlin/projects/${model.project_id}/models/${model.id}`,
        },
        {
          text: `Version ${version.id}`,
          href: `/merlin/projects/${model.project_id}/models/${model.id}/versions/${version.id}`,
        },
        { text: `Redeploy` },
      ]);
    }
  }, [model, version]);

  const [submissionResponse, submitForm] = useMerlinApi(
    `/models/${model.id}/versions/${version.id}/endpoint/${endpointId}`,
    { method: "PUT" },
    {},
    false,
  );

  return (
    <EuiPageTemplate restrictWidth="90%" paddingSize="none">
      <EuiSpacer size="l" />
      <EuiPageTemplate.Header
        bottomBorder={false}
        iconType={"machineLearningApp"}
        pageTitle={
          <Fragment>
            {"Redeploy "}
            {model.name}
            {" version "}
            {version.id}
          </Fragment>
        }
      />

      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
        <EuiPanel color={"transparent"}>
          {modelLoaded && versionLoaded ? (
            <FormContextProvider
              data={VersionEndpoint.fromJson(
                version.endpoints.find((e) => e.id === endpointId),
              )}
            >
              <DeployModelVersionForm
                model={model}
                version={Version.fromJson(version)}
                onCancel={() => window.history.back()}
                onSuccess={(redirectUrl) => navigate(redirectUrl)}
                submissionResponse={submissionResponse}
                submitForm={submitForm}
                actionTitle="Redeploy"
                isEnvironmentDisabled={true}
              />
            </FormContextProvider>
          ) : (
            <EuiSkeletonText lines={3} />
          )}
        </EuiPanel>
      </EuiPageTemplate.Section>
      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};

export default RedeployModelVersionView;
