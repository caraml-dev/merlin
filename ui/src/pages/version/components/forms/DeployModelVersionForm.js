import React, { useContext, useEffect, useState } from "react";
import {
  addToast,
  ConfirmationModal,
  FormContext,
  StepsWizardHorizontal,
} from "@caraml-dev/ui-lib";
import { DeploymentSummary } from "./components/DeploymentSummary";
import { CustomTransformerStep } from "./steps/CustomTransformerStep";
import { FeastTransformerStep } from "./steps/FeastTransformerStep";
import { ModelStep } from "./steps/ModelStep";
import { PredictionLoggerStep } from "./steps/PredictionLoggerStep";
import { StandardTransformerStep } from "./steps/StandardTransformerStep";
import { TransformerStep } from "./steps/TransformerStep";
import {
  customTransformerSchema,
  feastEnricherTransformerSchema,
  standardTransformerSchema,
  transformerConfigSchema,
  versionEndpointSchema,
} from "./validation/schema";
import { PROTOCOL } from "../../../../services/version_endpoint/VersionEndpoint";
import EnvironmentsContext from "../../../../providers/environments/context";

const _ = require('lodash');
const targetRequestStatus = (currentStatus) => {
  return currentStatus === "serving" ? "serving" : "running";
};

export const DeployModelVersionForm = ({
  model,
  version,
  onCancel,
  onSuccess,
  submissionResponse,
  submitForm,
  actionTitle = "Deploy",
  isEnvironmentDisabled = false,
}) => {
  const { data: versionEndpoint } = useContext(FormContext);

  useEffect(() => {
    if (submissionResponse.isLoaded && !submissionResponse.error) {
      addToast({
        id: "submit-success-deploy",
        title: "The deployment process is starting",
        color: "success",
        iconType: "check",
      });
      onSuccess(
        `/merlin/projects/${model.project_id}/models/${model.id}/versions/${version.id}/endpoints/${submissionResponse.data.id}/details`
      );
    }
  }, [submissionResponse, onSuccess, model, version]);

  const onSubmit = () => {

    // versionEndpoint toJSON() is not invoked, binding that causes many issues
    if (versionEndpoint?.resource_request?.cpu_limit === "") {
      delete versionEndpoint.resource_request.cpu_limit;
    }
    if (versionEndpoint?.image_builder_resource_request?.cpu_request === "") {
      delete versionEndpoint.image_builder_resource_request.cpu_request;
    }
    if (versionEndpoint?.image_builder_resource_request?.memory_request === "") {
      delete versionEndpoint.image_builder_resource_request.memory_request;
    }
    if (_.isEmpty(versionEndpoint.image_builder_resource_request)) {
      delete versionEndpoint.image_builder_resource_request;
    }
    submitForm({
      body: JSON.stringify({
        ...versionEndpoint,
        status: targetRequestStatus(versionEndpoint.status),
      }),
    });
  }

  const { data } = useContext(FormContext);
  const environments = useContext(EnvironmentsContext);

  const [maxAllowedReplica, setMaxAllowedReplica] = useState(() => {
    if (data.environment_name !== "") {
      return environments.find((env) => env.name === data.environment_name).max_allowed_replica;
    }
    return undefined;
  });

  const mainSteps = [
    {
      title: "Model",
      children: (
        <ModelStep
          version={version}
          isEnvironmentDisabled={isEnvironmentDisabled}
          maxAllowedReplica={maxAllowedReplica}
          setMaxAllowedReplica={setMaxAllowedReplica}
        />
      ),
      validationSchema: versionEndpointSchema(maxAllowedReplica),
    },
    {
      title: "Transformer",
      children: <TransformerStep
        maxAllowedReplica={maxAllowedReplica}
      />,
      validationSchema: transformerConfigSchema(maxAllowedReplica),
    },
  ];

  const standardTransformerStep = {
    title: "Standard Transformer",
    children: <StandardTransformerStep />,
    validationSchema: standardTransformerSchema,
    width: "100%",
  };

  const predictionLoggerStep = {
    title: "Logging",
    children: <PredictionLoggerStep />,
    validationSchema: versionEndpointSchema(maxAllowedReplica),
  };

  const customTransformerStep = {
    title: "Custom Transformer",
    children: <CustomTransformerStep />,
    validationSchema: customTransformerSchema,
  };

  const feastTransformerStep = {
    title: "Feast Enricher",
    children: <FeastTransformerStep />,
    validationSchema: feastEnricherTransformerSchema,
    width: "100%",
  };

  const [steps, setSteps] = useState(mainSteps);
  useEffect(
    () => {
      if (versionEndpoint.transformer && versionEndpoint.transformer.enabled) {
        switch (versionEndpoint.transformer.type_on_ui) {
          case "standard":
            if (versionEndpoint.protocol === PROTOCOL.UPI_V1) {
              setSteps([
                ...mainSteps,
                standardTransformerStep,
                predictionLoggerStep,
              ]);
            } else {
              setSteps([...mainSteps, standardTransformerStep]);
            }
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
      } else {
        setSteps(mainSteps);
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
          versionEndpoint={versionEndpoint}
        />
      }
      isLoading={submissionResponse.isLoading}
      onConfirm={onSubmit}
      confirmButtonText={actionTitle}
      confirmButtonColor="primary"
    >
      {(onSubmit) => (
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
