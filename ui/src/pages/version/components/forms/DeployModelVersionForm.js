import React, { useContext, useEffect, useState } from "react";
import {
  addToast,
  ConfirmationModal,
  FormContext,
  StepsWizardHorizontal
} from "@gojek/mlp-ui";
import { DeploymentSummary } from "./components/DeploymentSummary";
import { CustomTransformerStep } from "./steps/CustomTransformerStep";
import { FeastTransformerStep } from "./steps/FeastTransformerStep";
import { ModelStep } from "./steps/ModelStep";
import { PipelineStep } from "./steps/PipelineStep";
import { TransformerStep } from "./steps/TransformerStep";
import {
  customTransformerSchema,
  feastEnricherTransformerSchema,
  postprocessTransformerSchema,
  preprocessTransformerSchema,
  transformerConfigSchema,
  versionEndpointSchema
} from "./validation/schema";

export const DeployModelVersionForm = ({
  model,
  version,
  onCancel,
  onSuccess,
  submissionResponse,
  submitForm,
  actionTitle = "Deploy",
  isEnvironmentDisabled = false
}) => {
  const { data: versionEndpoint } = useContext(FormContext);

  useEffect(() => {
    if (submissionResponse.isLoaded && !submissionResponse.error) {
      addToast({
        id: "submit-success-deploy",
        title: "The deployment process is starting",
        color: "success",
        iconType: "check"
      });
      onSuccess(
        `/merlin/projects/${model.project_id}/models/${model.id}/versions/${version.id}/endpoints/${submissionResponse.data.id}/details`
      );
    }
  }, [submissionResponse, onSuccess, model, version]);

  const onSubmit = () => submitForm({ body: JSON.stringify(versionEndpoint) });

  const mainSteps = [
    {
      title: "Model",
      children: (
        <ModelStep
          version={version}
          isEnvironmentDisabled={isEnvironmentDisabled}
        />
      ),
      validationSchema: versionEndpointSchema
    },
    {
      title: "Transformer",
      children: <TransformerStep />,
      validationSchema: transformerConfigSchema
    }
  ];

  const standardTransformerSteps = [
    {
      title: "Preprocess",
      children: <PipelineStep stage="preprocess" />,
      validationSchema: preprocessTransformerSchema,
      width: "100%"
    },
    {
      title: "Postprocess",
      children: <PipelineStep stage="postprocess" />,
      validationSchema: postprocessTransformerSchema,
      width: "100%"
    }
  ];

  const customTransformerStep = {
    title: "Custom Transformer",
    children: <CustomTransformerStep />,
    validationSchema: customTransformerSchema
  };

  const feastTransformerStep = {
    title: "Feast Enricher",
    children: <FeastTransformerStep />,
    validationSchema: feastEnricherTransformerSchema,
    width: "100%"
  };

  const [steps, setSteps] = useState(mainSteps);
  useEffect(
    () => {
      if (versionEndpoint.transformer) {
        switch (versionEndpoint.transformer.type_on_ui) {
          case "standard":
            setSteps([...mainSteps, ...standardTransformerSteps]);
            break;
          case "custom":
            setSteps([...mainSteps, customTransformerStep]);
            break;
          case "feast":
            setSteps([...mainSteps, feastTransformerStep]);
            break;
          default:
            setSteps(mainSteps);
            break;
        }
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [versionEndpoint]
  );

  return (
    <ConfirmationModal
      title={`${actionTitle} Model Version`}
      content={
        <DeploymentSummary
          actionTitle={actionTitle}
          modelName={model.name}
          versionId={version.id}
        />
      }
      isLoading={submissionResponse.isLoading}
      onConfirm={onSubmit}
      confirmButtonText={actionTitle}
      confirmButtonColor="primary">
      {onSubmit => (
        <StepsWizardHorizontal
          steps={steps}
          onCancel={onCancel}
          onSubmit={onSubmit}
          submitLabel={actionTitle}
        />
      )}
    </ConfirmationModal>
  );
};
