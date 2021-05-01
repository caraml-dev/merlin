export class Transformer {
  constructor() {
    this.id = undefined;
    this.enabled = undefined;
    this.version_endpoint_id = undefined;
    this.transformer_type = "";

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
    this.created_at = undefined;
    this.updated_at = undefined;
  }
}
