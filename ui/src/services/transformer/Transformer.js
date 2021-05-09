import {
  Config,
  STANDARD_TRANSFORMER_CONFIG_ENV_NAME
} from "./TransformerConfig";

const objectAssignDeep = require(`object-assign-deep`);

export class Transformer {
  constructor() {
    this.id = undefined;
    this.enabled = undefined;
    this.version_endpoint_id = undefined;
    this.transformer_type = "";
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

    this.config = undefined; // TransformerConfig

    this.created_at = undefined;
    this.updated_at = undefined;

    this.toJSON = this.toJSON.bind(this);
  }

  toJSON() {
    let obj = objectAssignDeep({}, this);

    // Update config to env vars
    //
    if (obj.config) {
      let configJson = JSON.stringify(new Config(obj.config));

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
    //
    // Delete config and feast_enricher_config, because we already set the config to env vars
    if (obj.type_on_ui) {
      // delete obj["type_on_ui"];
    }
    if (obj.config) {
      // delete obj["config"];
    }
    if (obj.feast_enricher_config) {
      // delete obj["feast_enricher_config"];
    }

    return obj;
  }
}
