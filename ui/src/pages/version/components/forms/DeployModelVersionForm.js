import React, { useContext, useEffect, useState } from "react";
import {
  addToast,
  ConfirmationModal,
  FormContext,
  StepsWizardHorizontal
} from "@gojek/mlp-ui";
import { useMerlinApi } from "../../../../hooks/useMerlinApi";
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
  onSuccess
}) => {
  const { data: modelVersion } = useContext(FormContext);

  useEffect(() => {
    console.log(JSON.stringify(modelVersion, null, 2));
  }, [modelVersion]);

  const [submissionResponse, submitForm] = useMerlinApi(
    `/models/${model.id}/versions/${version.id}/endpointasd`, // TODO: Use the correct endpoint once ready
    { method: "POST" },
    {},
    false
  );

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

  const onSubmit = () => submitForm({ body: JSON.stringify(modelVersion) });

  const mainSteps = [
    {
      title: "Model",
      children: <ModelStep />,
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
      validationSchema: preprocessTransformerSchema
    },
    {
      title: "Postprocess",
      children: <PipelineStep stage="postprocess" />,
      validationSchema: postprocessTransformerSchema
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
    validationSchema: feastEnricherTransformerSchema
  };

  const [steps, setSteps] = useState(mainSteps);
  useEffect(
    () => {
      if (modelVersion.transformer) {
        switch (modelVersion.transformer.type_on_ui) {
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
    [modelVersion]
  );

  return (
    <ConfirmationModal
      title="Deploy Model Version"
      content={
        <DeploymentSummary modelName={model.name} versionId={version.id} />
      }
      isLoading={submissionResponse.isLoading}
      onConfirm={onSubmit}
      confirmButtonText="Deploy"
      confirmButtonColor="primary">
      {onSubmit => (
        <StepsWizardHorizontal
          steps={steps}
          onCancel={onCancel}
          onSubmit={onSubmit}
          submitLabel="Deploy"
        />
      )}
    </ConfirmationModal>
  );
};
