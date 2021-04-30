import React, { useContext, useEffect, useState } from "react";
import {
  ConfirmationModal,
  FormContext,
  StepsWizardHorizontal
} from "@gojek/mlp-ui";
import { useMerlinApi } from "../../../hooks/useMerlinApi";
import { DeploymentSummary } from "./DeploymentSummary";
import { ModelStep } from "../components/steps/ModelStep";
import { TransformerStep } from "../components/steps/TransformerStep";
import { versionEndpointSchema } from "../components/validation/schema";

export const DeployModelVersionForm = ({ model, version, onCancel }) => {
  const { data: modelVersion } = useContext(FormContext);
  useEffect(() => {
    console.table(modelVersion);
  }, [modelVersion]);

  const [submissionResponse, submitForm] = useMerlinApi(
    `/models/${model.id}/versions/${version.id}/endpoint`,
    { method: "POST", addToast: true },
    {},
    false
  );

  const onSubmit = () => submitForm({ body: JSON.stringify(modelVersion) });

  const [isStandardTransformer, setIsStandardTransformer] = useState(false);

  const allSteps = [
    {
      title: "Model",
      children: <ModelStep />,
      validationSchema: versionEndpointSchema
    },
    {
      title: "Transformer",
      children: (
        <TransformerStep setIsStandardTransformer={setIsStandardTransformer} />
      )
      // validationSchema: schema[1],
      // validationContext: { experimentEngineOptions }
    },
    {
      title: "Preprocess",
      children: <TransformerStep />
      // validationSchema: schema[1],
      // validationContext: { experimentEngineOptions }
    },
    {
      title: "Postprocess",
      children: <TransformerStep />
      // validationSchema: schema[1],
      // validationContext: { experimentEngineOptions }
    }
  ];

  const [steps, setSteps] = useState(allSteps.slice(0, 2));
  useEffect(
    () => {
      setSteps(isStandardTransformer ? allSteps : allSteps.slice(0, 2));
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [isStandardTransformer]
  );

  return (
    <ConfirmationModal
      title="Deploy Model Version"
      content={<DeploymentSummary />}
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
    // <EuiText>WTF</EuiText>
  );
};
