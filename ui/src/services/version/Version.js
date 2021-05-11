import objectAssignDeep from "object-assign-deep";
import { VersionEndpoint } from "../version_endpoint/VersionEndpoint";

export class Version {
  constructor() {
    this.id = undefined;

    this.model_id = undefined;
    this.model = undefined;

    this.mlflow_run_id = undefined;
    this.mlflow_url = undefined;
    this.artifact_uri = undefined;

    this.endpoints = undefined;

    this.properties = undefined;
    this.labels = undefined;

    this.created_at = undefined;
    this.updated_at = undefined;
  }

  static fromJson(json) {
    const version = objectAssignDeep(new Version(), json);
    version.endpoints &&
      version.endpoints.forEach((endpoint, idx) => {
        version.endpoints[idx] = VersionEndpoint.fromJson(endpoint);
      });
    return version;
  }
}
