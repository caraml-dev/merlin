/**
 * Copyright 2020 The Merlin Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

const objectAssignDeep = require(`object-assign-deep`);

export class Job {
  constructor(project_id, model_id, version_id) {
    this.id = 0;
    this.name = "";

    this.project_id = parseInt(project_id) ? parseInt(project_id) : 0;
    this.model_id = parseInt(model_id) ? parseInt(model_id) : 0;
    this.version_id = parseInt(version_id) ? parseInt(version_id) : 0;

    // this.status = "";
    // this.error = "";

    this.config = {
      service_account_name: "",

      resource_request: {},

      env_vars: [],

      job_config: {
        model: {
          type: "",
          uri: "",
          result: {
            type: "DOUBLE",
            item_type: "DOUBLE"
          },
          options: {}
        },
        bigquerySource: {
          table: "",
          features: [],
          options: {
            // parentProject: "",
            // maxParallelism: "0",
            viewsEnabled: "false",
            // viewMaterializationProject: "",
            // viewMaterializationDataset: "",
            readDataFormat: "AVRO",
            optimizedEmptyProjection: "true"
            // filter: ""
          }
        },
        bigquerySink: {
          table: "",
          stagingBucket: "",
          resultColumn: "",
          saveMode: "ERRORIFEXISTS",
          options: {
            createDisposition: "CREATE_IF_NEEDED",
            intermediateFormat: "parquet",
            // partitionField: "",
            // partitionExpirationMs: "0",
            partitionType: "DAY",
            // clusteredFields: "",
            allowFieldAddition: "false",
            allowFieldRelaxation: "false"
          }
        }
      }
    };
  }

  static from(json) {
    if (!json.config.env_vars) {
      json.config.env_vars = [];
    }

    return objectAssignDeep(new Job(), json);
  }
}
