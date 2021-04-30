export class VersionEndpoint {
  constructor() {
    this.id = "";
    this.version_id = 0;
    this.model_id = 0;
    this.status = "";
    this.url = "";
    this.service_name = "";
    this.monitoring_url = "";

    this.environment = {
      id: 0,
      name: "",
      cluster: "",
      is_default: false,
      created_at: "",
      updated_at: ""
    };
    this.environment_name = "";

    this.message = "";

    this.resource_request = {
      min_replica: process.env.REACT_APP_ENVIRONMENT === "production" ? 2 : 0,
      max_replica: process.env.REACT_APP_ENVIRONMENT === "production" ? 4 : 2,
      cpu_request: "500m",
      memory_request: "512Mi"
    };

    this.env_vars = [];

    // TODO
    this.transformer = {};

    this.created_at = "";
    this.updated_at = "";
  }
}
