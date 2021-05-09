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

export const STANDARD_TRANSFORMER_CONFIG_ENV_NAME =
  "STANDARD_TRANSFORMER_CONFIG";

export class Config {
  constructor(transformerConfig) {
    this.transformerConfig = transformerConfig;
  }

  // Deprecated.
  static from(jsonObject) {
    return objectAssignDeep(new Config(), jsonObject);
  }

  static fromJson(json) {
    const config = objectAssignDeep(new Config(), json);
    config.transformerConfig = TransformerConfig.fromJson(
      json.transformerConfig
    );
    return config;
  }
}

export class TransformerConfig {
  constructor(feast, preprocess, postprocess) {
    this.feast = feast;
    this.preprocess = preprocess;
    this.postprocess = postprocess;

    this.toJSON = this.toJSON.bind(this);
  }

  static fromJson(json) {
    const transformerConfig = objectAssignDeep(new TransformerConfig(), json);

    transformerConfig.feast &&
      transformerConfig.feast.forEach(feast => {
        feast.entities &&
          feast.entities.forEach(entity => {
            if (entity.udf) {
              entity["fieldType"] = "UDF";
              entity["field"] = entity.udf;
            } else {
              entity["fieldType"] = "JSONPath";
              entity["field"] = entity.jsonPath;
            }
          });
      });

    return transformerConfig;
  }

  toJSON() {
    let obj = objectAssignDeep({}, this);

    obj.feast &&
      obj.feast.forEach(feast => {
        feast.entities &&
          feast.entities.forEach(entity => {
            if (entity.fieldType === "UDF") {
              entity["udf"] = entity.field;
            } else {
              entity["jsonPath"] = entity.field;
            }
            delete entity["fieldType"];
            delete entity["field"];
          });
      });

    return obj;
  }
}

export class Pipeline {
  constructor() {
    this.inputs = [];
    this.transformations = [];
    this.outputs = [];

    this.toJSON = this.toJSON.bind(this);
  }

  toJSON() {
    let obj = objectAssignDeep({}, this);

    if (obj.inputs.length === 0) {
      delete obj["inputs"];
    } else {
      obj.inputs.forEach(input => {
        input.feast &&
          input.feast.forEach(feast => {
            delete feast["isTableNameEditable"];

            input.feast.entities &&
              input.feast.entities.forEach(entity => {
                delete entity["fieldType"];
                delete entity["field"];
              });
          });

        input.tables &&
          input.tables.forEach(table => {
            table.columns &&
              table.columns.forEach(column => {
                delete column["idx"];
                delete column["type"];
                delete column["value"];
              });
          });

        input.variables &&
          input.variables.forEach(variable => {
            delete variable["idx"];
            delete variable["type"];
            delete variable["value"];
          });
      });
    }

    if (obj.transformations.length === 0) {
      delete obj["transformations"];
    }
    if (obj.outputs.length === 0) {
      delete obj["outputs"];
    }

    return obj;
  }
}

export class Input {
  constructor() {
    this.feast = undefined;
    this.tables = undefined;
    this.variables = undefined;
  }
}

export class FeastInput {
  constructor(isTableNameEditable) {
    this.tableName = undefined;
    if (isTableNameEditable) {
      this.isTableNameEditable = isTableNameEditable;
      this.tableName = "";
    }

    this.project = "";
    this.entities = [];
    this.features = [];
  }
}

export class TablesInput {
  constructor() {
    this.name = "";
    this.baseTable = undefined;
    this.columns = [];
  }
}

export class Transformations {
  constructor() {
    this.tableTransformation = new TableTransformation();
    this.tableJoin = new TableJoin();
  }
}

export class TableTransformation {
  constructor() {
    this.inputTable = "";
    this.outputTable = "";
    this.steps = [{}];
  }
}

export class TableJoin {
  constructor() {
    this.leftTable = undefined;
    this.rightTable = undefined;
    this.outputTable = undefined;
    this.how = undefined;
    this.onColumn = undefined;
  }
}

export class Output {
  constructor() {
    this.jsonOutput = new JsonOutput();
  }
}

export class JsonOutput {
  constructor() {
    this.jsonTemplate = new JsonTemplate();
  }
}

export class BaseJson {
  constructor() {
    this.jsonPath = "";
  }
}

export class JsonTemplate {
  constructor() {
    this.baseJson = undefined;
    this.fields = [];
  }
}

export class Field {
  constructor() {
    this.fieldName = "";
    this.fields = [];
    this.value = undefined;
  }
}

export class FieldFromJson {
  constructor() {
    this.fromJson = new FromJson();
  }
}

export class FieldFromTable {
  constructor() {
    this.fromTable = new FromTable();
  }
}

export class FieldFromExpression {
  constructor() {
    this.expression = undefined;
  }
}

export class FromJson {
  constructor() {
    this.jsonPath = "";
    this.addRowNumber = false;
  }
}

export class FromTable {
  constructor() {
    this.tableName = "";
    this.format = "";
  }
}
