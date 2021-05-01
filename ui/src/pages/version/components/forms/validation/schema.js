import * as yup from "yup";
import { appConfig } from "../../../../../config";

const cpuRequestRegex = /^(\d{1,3}(\.\d{1,3})?)$|^(\d{2,5}m)$/,
  memRequestRegex = /^\d+(Ei?|Pi?|Ti?|Gi?|Mi?|Ki?)?$/,
  envVariableNameRegex = /^[a-z0-9_]*$/i,
  dockerImageRegex = /^([a-z0-9]+(?:[._-][a-z0-9]+)*(?::\d{2,5})?\/)?([a-z0-9]+(?:[._-][a-z0-9]+)*\/)*([a-z0-9]+(?:[._-][a-z0-9]+)*)(?::[a-z0-9]+(?:[._-][a-z0-9]+)*)?$/i;

const resourceRequestSchema = yup.object().shape({
  cpu_request: yup
    .string()
    .matches(cpuRequestRegex, 'Valid CPU value is required, e.g "2" or "500m"'),
  memory_request: yup
    .string()
    .matches(memRequestRegex, "Valid RAM value is required, e.g. 512Mi"),
  min_replica: yup
    .number()
    .typeError("Min Replicas value is required")
    .min(0, "Min Replicas can not be less than 0"),
  max_replica: yup
    .number()
    .typeError("Max Replicas value is required")
    .min(
      yup.ref(`min_replica`),
      "Max Replicas can not be less than Min Replicas"
    )
    .max(
      appConfig.scaling.maxAllowedReplica,
      // eslint-disable-next-line no-template-curly-in-string
      "Max Replicas value has exceeded allowed number of replicas: ${max}"
    )
    .when("min_replica", (minReplica, schema) =>
      minReplica === 0
        ? schema.positive("Max Replica should be positive")
        : schema
    )
});

const environmentVariableSchema = yup.object().shape({
  name: yup
    .string()
    .required("Variable name can not be empty")
    .matches(
      envVariableNameRegex,
      "The name of a variable can contain only alphanumeric character or the underscore"
    ),
  value: yup.string()
});

export const versionEndpointSchema = yup.object().shape({
  environment_name: yup.string().required("Environment is required"),
  resource_request: resourceRequestSchema,
  env_vars: yup.array(environmentVariableSchema)
});

export const transformerConfigSchema = yup.object().shape({
  transformer: yup.object().shape({
    resource_request: resourceRequestSchema,
    env_vars: yup.array(environmentVariableSchema)
  })
});

const dockerImageSchema = yup
  .string()
  .matches(
    dockerImageRegex,
    "Valid Docker Image value should be provided, e.g. kennethreitz/httpbin:latest"
  );

export const customTransformerSchema = yup.object().shape({
  transformer: yup.object().shape({
    image: dockerImageSchema.required("Docker Image is required")
  })
});
