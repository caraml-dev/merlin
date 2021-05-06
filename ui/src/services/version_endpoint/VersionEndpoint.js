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

    this.environment = {
      id: undefined,
      name: undefined,
      cluster: undefined,
      is_default: false,
      created_at: undefined,
      updated_at: undefined
    };
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

  toJSON() {
    let obj = objectAssignDeep({}, this);
    return obj;
  }
}
