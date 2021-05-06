import React, { useContext, useEffect, useState } from "react";
import {
  addToast,
  ConfirmationModal,
  FormContext,
  StepsWizardHorizontal
} from "@gojek/mlp-ui";
import { useMerlinApi } from "../../../../hooks/useMerlinApi";
import { DeploymentSummary } from "./components/DeploymentSummary";
import { ModelStep } from "./steps/ModelStep";
import { PostprocessStep } from "./steps/PostprocessStep";
import { PreprocessStep } from "./steps/PreprocessStep";
import { TransformerStep } from "./steps/TransformerStep";
import {
  customTransformerSchema,
  transformerConfigSchema,
  versionEndpointSchema
} from "./validation/schema";
import { CustomTransformerStep } from "./steps/CustomTransformerStep";
import { FeastTransformerStep } from "./steps/FeastTransformerStep";

export const DeployModelVersionForm = ({
  model,
  version,
  onCancel,
  onSuccess
}) => {
  const { data: modelVersion } = useContext(FormContext);
  useEffect(() => {
    console.log(JSON.stringify(modelVersion.transformer.config));
  }, [modelVersion]);

  const [submissionResponse, submitForm] = useMerlinApi(
    `/models/${model.id}/versions/${version.id}/endpointasd`,
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
    // {
    //   title: "Model",
    //   children: <ModelStep />,
    //   validationSchema: versionEndpointSchema
    // },
    // {
    //   title: "Transformer",
    //   children: <TransformerStep />,
    //   validationSchema: transformerConfigSchema,
    // },
    {
      title: "Preprocess",
      children: <PreprocessStep />
      // validationSchema: schema[1],
    },
    {
      title: "Postprocess",
      children: <PostprocessStep />
      // validationSchema: schema[1],
    }
  ];

  const standardTransformerSteps = [
    {
      title: "Preprocess",
      children: <PreprocessStep />
      // validationSchema: schema[1],
    },
    {
      title: "Postprocess",
      children: <PostprocessStep />
      // validationSchema: schema[1],
    }
  ];

  const customTransformerStep = {
    title: "Custom Transformer",
    children: <CustomTransformerStep />,
    validationSchema: customTransformerSchema
  };

  const feastTransformerStep = {
    title: "Feast Enricher",
    children: <FeastTransformerStep />
    // validationSchema: schema[1], // TODO
  };

  const [steps, setSteps] = useState(mainSteps);
  useEffect(
    () => {
      if (modelVersion.transformer) {
        switch (modelVersion.transformer.transformer_type) {
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
