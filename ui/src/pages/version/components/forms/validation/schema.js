import * as yup from "yup";

const cpuRequestRegex = /^(\d{1,3}(\.\d{1,3})?)$|^(\d{2,5}m)$/,
  memRequestRegex = /^\d+(Ei?|Pi?|Ti?|Gi?|Mi?|Ki?)?$/,
  gpuRequestRegex= /^\d+$/,
  envVariableNameRegex = /^[a-z0-9_]*$/i,
  dockerImageRegex =
    /^([a-z0-9]+(?:[._-][a-z0-9]+)*(?::\d{2,5})?\/)?([a-z0-9]+(?:[._-][a-z0-9]+)*\/)*([a-z0-9]+(?:[._-][a-z0-9]+)*)(?::[a-z0-9]+(?:[._-][a-z0-9]+)*)?$/i,
  configMapNameRegex = /^[-._a-zA-Z0-9]+$/;

const resourceRequestSchema = (maxAllowedReplica) => yup.object().shape({
  cpu_request: yup
    .string()
    .matches(cpuRequestRegex, 'Valid CPU value is required, e.g "2" or "500m"'),
  cpu_limit: yup
    .string()
    .matches(cpuRequestRegex, { message: 'Valid CPU value is required, e.g "2" or "500m"', excludeEmptyString: true }),
  memory_request: yup
    .string()
    .matches(memRequestRegex, "Valid RAM value is required, e.g. 512Mi"),
  gpu_request: yup
    .string()
    .typeError("GPU value is required")
    .when("gpu_name", {
      is: (v) => v !== undefined,
      then: (_) => yup.string().matches(gpuRequestRegex, { message:'Valid GPU value is required'}).required("GPU value is required"),
    }),
  min_replica: yup
    .number()
    .typeError("Min Replicas value is required")
    .min(0, "Min Replicas cannot be less than 0"),
  max_replica: yup
    .number()
    .typeError("Max Replicas value is required")
    .min(
      yup.ref(`min_replica`),
      "Max Replicas cannot be less than Min Replicas"
    )
    .max(
      maxAllowedReplica,
      // eslint-disable-next-line no-template-curly-in-string
      "Max Replicas value has exceeded allowed number of replicas: ${max}"
    )
    .when("min_replica", ([minReplica], schema) =>
      minReplica === 0
        ? schema.positive("Max Replica should be positive")
        : schema
    ),
});

const imagebuilderRequestSchema = yup.object().nullable().shape({
  cpu_request: yup
  .string()
  .matches(cpuRequestRegex, { message:'Valid CPU value is required, e.g "2" or "500m"', excludeEmptyString: true }),
  memory_request: yup
  .string()
  .matches(memRequestRegex, { message:"Valid RAM value is required, e.g. 512Mi", excludeEmptyString: true }),
});

const environmentVariableSchema = yup.object().shape({
  name: yup
    .string()
    .required("Variable name is required")
    .matches(
      envVariableNameRegex,
      "The name of an environment variable must contain only alphanumeric characters or '_'"
    ),
  value: yup.string(),
});

const secretSchema = yup.object().shape({
  mlp_secret_name: yup
    .string()
    .required("MLP secret name is required")
    .matches(
      configMapNameRegex,
      "The name of the MLP secret must contain only alphanumeric characters, '-', '_' or '.'"
    ),
  env_var_name: yup
    .string()
    .required("Environment variable name is required")
    .matches(
      envVariableNameRegex,
      "The name of an environment variable must contain only alphanumeric characters or '_'"
    ),
});


const stringValueSchema = (name, key) => yup.mixed().test(
  `value input schema`,
  `${name} must be a \`string\` but the final value was $\{value}`,
  (value, { createError }) => {
    try {
      if (typeof value == "object" && key in value) {
        return yup.string().required(`${name} is required`).validateSync(value[key])
      }
      return yup.string().required(`${name} is required`).validateSync(value)
    } catch (e) {
      return createError({ message: e.message })
    }
  }
)

export const versionEndpointSchema = (maxAllowedReplica) => yup.object().shape({
  environment_name: yup.string().required("Environment is required"),
  resource_request: resourceRequestSchema(maxAllowedReplica),
  image_builder_resource_request: imagebuilderRequestSchema,
  env_vars: yup.array(environmentVariableSchema),
  secrets: yup.array(secretSchema),
});

export const transformerConfigSchema = (maxAllowedReplica) => yup.object().shape({
  transformer: yup.object().shape({
    resource_request: resourceRequestSchema(maxAllowedReplica),
    env_vars: yup.array(environmentVariableSchema),
    secrets: yup.array(secretSchema),
  }),
});

const dockerImageSchema = yup
  .string()
  .matches(
    dockerImageRegex,
    "Valid Docker Image value should be provided, e.g. kennethreitz/httpbin:latest"
  );

const feastEntitiesSchema = yup.object().shape({
  name: yup.string().required("Entity Name is required"),
  valueType: yup.string().required("Entity Value Type is required"),
  fieldType: yup.string().required("Input Type is required"),
  field: stringValueSchema("Input Value", "jsonPath"),
});

const feastFeaturesSchema = yup.object().shape({
  name: yup.string().required("Feature Name is required"),
});

export const feastInputSchema = yup.object().shape({
  tableName: yup.string().when("isTableNameEditable", {
    is: true,
    then: (_) => yup.string().required("Table name is required"),
  }),
  project: yup.string().required("Project name is required"),
  source: yup.string().required("Source is required"),
  entities: yup.array(feastEntitiesSchema),
  features: yup.array(feastFeaturesSchema),
});

const variableInputSchema = yup.object().shape({
  name: yup.string().required("Name is required"),
  type: yup.string().required("Type is required"),
  value: stringValueSchema("Value", "jsonPath"),
});

const tablesInputSchema = yup.object().shape({
  name: yup.string().required("Table Name is required"),
  baseTable: yup.object().shape({
    fromJson: yup
      .object()
      .nullable()
      .default(undefined)
      .shape({
        jsonPath: yup.string().required("JSONPath is required"),
      }),
    fromFile: yup
      .object()
      .nullable()
      .default(undefined)
      .shape({
        uri: yup.string().required("URI to file is required"),
      }),
    fromTable: yup
      .object()
      .nullable()
      .default(undefined)
      .shape({
        tableName: yup.string().required("Table Name Source is required"),
      }),
  }),
  columns: yup
    .array()
    .when("baseTable.fromTable.tableName", ([tableName], schema) => {
      return tableName !== undefined ? yup.array(variableInputSchema) : schema;
    }),
});

const encodersInputSchema = yup.object().shape({
  name: yup.string().required("Encoder Name is required"),
  ordinalEncoderConfig: yup
    .object()
    .nullable()
    .default(undefined)
    .shape({
      defaultValue: yup.string().required("Default value required"),
    }),
});

const inputPipelineSchema = yup.object().shape({
  feast: yup.array(feastInputSchema),
  tables: yup.array(tablesInputSchema),
  variables: yup.array(variableInputSchema),
  encoders: yup.array(encodersInputSchema),
});

const tableTransformationStep = yup.object().shape({
  operation: yup.string().required("Operation is required"),
  dropColumns: yup
    .array()
    .of(yup.string())
    .when("operation", {
      is: (v) => v !== undefined && v === "dropColumns",
      then: (_) => yup.array().required("List of columns to be deleted is required"),
    }),
  selectColumns: yup
    .array()
    .of(yup.string())
    .when("operation", {
      is: (v) => v !== undefined && v === "selectColumns",
      then: (_) => yup.array().required("List of columns to be selected is required"),
    }),
});

const transformationPipelineSchema = yup.object().shape({
  tableTransformation: yup
    .object()
    .nullable()
    .default(undefined)
    .shape({
      inputTable: yup.string().required("Input Table Name is required"),
      outputTable: yup.string().required("Output Table Name is required"),
      steps: yup.array(tableTransformationStep),
    }),
  tableJoin: yup
    .object()
    .nullable()
    .default(undefined)
    .shape({
      leftTable: yup.string().required("Left Table is required"),
      rightTable: yup.string().required("Right Table is required"),
      outputTable: yup.string().required("Output Table is required"),
      how: yup.string().required("Join Method is required"),
      onColumns: yup
        .array()
        .of(yup.string())
        .when("how", {
          is: (v) =>
            v !== undefined && v !== "" && v !== "CROSS" && v !== "CONCAT",
          then: (_) => yup.array().required("On Columns is required"),
        }),
    }),
});

const outputPipelineSchema = yup.object().shape({
  jsonOutput: yup.object().shape({
    jsonTemplate: yup.object().shape({
      baseJson: yup
        .object()
        .nullable()
        .default(undefined)
        .shape({
          jsonPath: yup.string().required("JSONPath is required"),
        }),
    }),
  }),
});

export const pipelineSchema = yup.object().shape({
  inputs: yup.array(inputPipelineSchema),
  // .required("One of inputs should be specified"),
  transformations: yup.array(transformationPipelineSchema),
  // .required("One of transformations should be specified"),
  outputs: yup.array(outputPipelineSchema),
  // .required("One of outputs should be specified")
});

export const customTransformerSchema = yup.object().shape({
  transformer: yup.object().shape({
    image: dockerImageSchema.required("Docker Image is required"),
    command: yup.string(),
    args: yup.string(),
  }),
});

export const feastEnricherTransformerSchema = yup.object().shape({
  transformer: yup.object().shape({
    config: yup.object().shape({
      transformerConfig: yup.object().shape({
        feast: yup.array(feastInputSchema),
      }),
    }),
  }),
});

export const standardTransformerSchema = yup.object().shape({
  transformer: yup.object().shape({
    config: yup.object().shape({
      transformerConfig: yup.object().shape({
        preprocess: pipelineSchema,
        postprocess: pipelineSchema,
      }),
    }),
  }),
});
