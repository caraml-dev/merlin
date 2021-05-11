import { Logger } from "../logger/Logger";
import { Transformer } from "../transformer/Transformer";

const objectAssignDeep = require(`object-assign-deep`);

export class VersionEndpoint {
  constructor() {
    this.id = undefined;
    this.version_id = undefined;
    this.model_id = undefined;
    this.status = undefined;
    this.url = undefined;
    this.service_name = undefined;
    this.monitoring_url = undefined;

    this.environment_name = "";

    this.message = undefined;

    this.resource_request = {
      min_replica: process.env.REACT_APP_ENVIRONMENT === "production" ? 2 : 0,
      max_replica: process.env.REACT_APP_ENVIRONMENT === "production" ? 4 : 2,
      cpu_request: "500m",
      memory_request: "512Mi"
    };

    this.env_vars = [];

    this.transformer = new Transformer();

    this.logger = new Logger();

    this.created_at = undefined;
    this.updated_at = undefined;
  }

  static fromJson(json) {
    const versionEndpoint = objectAssignDeep(new VersionEndpoint(), json);

    if (!versionEndpoint.env_vars) {
      versionEndpoint.env_vars = [];
    } else {
      versionEndpoint.env_vars = versionEndpoint.env_vars.filter(
        e => e.name !== "MODEL_NAME" && e.name !== "MODEL_DIR"
      );
    }

    if (json.transformer) {
      versionEndpoint.transformer = Transformer.fromJson(json.transformer);
    }

    if (json.logger) {
      versionEndpoint.logger = Logger.fromJson(json.logger);
    }

    return versionEndpoint;
  }

  toJSON() {
    let obj = objectAssignDeep({}, this);

    if (obj.env_vars && obj.env_vars.length > 0) {
      obj["env_vars"] = obj.env_vars.filter(
        e => e.name !== "MODEL_NAME" && e.name !== "MODEL_DIR"
      );
    }

    return obj;
  }
}
