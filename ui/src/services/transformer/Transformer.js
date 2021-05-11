import {
  Config,
  STANDARD_TRANSFORMER_CONFIG_ENV_NAME
} from "./TransformerConfig";

const objectAssignDeep = require(`object-assign-deep`);

export class Transformer {
  constructor() {
    this.id = undefined;
    this.enabled = false;
    this.version_endpoint_id = undefined;
    this.transformer_type = undefined;
    this.type_on_ui = ""; // Only used on UI side

    // Custom transformer's properties.
    this.image = undefined;
    this.command = undefined;
    this.args = undefined;

    this.resource_request = {
      min_replica: process.env.REACT_APP_ENVIRONMENT === "production" ? 2 : 0,
      max_replica: process.env.REACT_APP_ENVIRONMENT === "production" ? 4 : 2,
      cpu_request: "500m",
      memory_request: "512Mi"
    };

    this.env_vars = [];

    this.config = undefined; // Config

    this.created_at = undefined;
    this.updated_at = undefined;

    this.toJSON = this.toJSON.bind(this);
  }

  static fromJson(json) {
    const transformer = objectAssignDeep(new Transformer(), json);
    transformer.type_on_ui = json.transformer_type;

    // Parse transformer config
    const envVarIndex = json.env_vars.findIndex(
      e => e.name === STANDARD_TRANSFORMER_CONFIG_ENV_NAME
    );

    if (envVarIndex !== -1) {
      transformer.config = Config.fromJson(
        JSON.parse(json.env_vars[envVarIndex].value)
      );

      if (transformer.config.transformerConfig.feast !== undefined) {
        transformer.type_on_ui = "feast";
      }
    }

    return transformer;
  }

  toJSON() {
    let obj = objectAssignDeep({}, this);

    // Update config to env vars
    //
    if (obj.config) {
      let configJson = JSON.stringify(obj.config);

      // Find the index of env_var that contains transformer config
      // If it's not exist, create new env var
      // If it's exist, update it
      const envVarIndex = obj.env_vars.findIndex(
        e => e.name === STANDARD_TRANSFORMER_CONFIG_ENV_NAME
      );
      if (envVarIndex === -1) {
        obj.env_vars.push({
          name: STANDARD_TRANSFORMER_CONFIG_ENV_NAME,
          value: configJson
        });
      } else {
        obj.env_vars[envVarIndex] = {
          ...obj.env_vars[envVarIndex],
          value: configJson
        };
      }
    }

    // Remove properties for optional fields, if not relevant
    delete obj["type_on_ui"];
    if (obj.config) {
      delete obj["config"];
    }

    return obj;
  }
}
