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

export class Entity {
  constructor(name, valueType, jsonPath, udf) {
    this.name = name;
    this.valueType = valueType;
    this.jsonPath = jsonPath;
    this.udf = udf;
  }
}

export class Feature {
  constructor(name, valueType, defaultValue) {
    this.name = name;
    this.valueType = valueType;
    this.defaultValue = defaultValue;
  }
}

export class FeastConfig {
  constructor(project, entities, features) {
    this.project = project;
    this.entities = entities; // Array of Entity
    this.features = features; // Array of Features
  }
}

export class TransformerConfig {
  constructor(feast) {
    this.feast = feast; // Array of FeastConfig

    this.preprocess = new Pipeline();
    this.postprocess = new Pipeline();
  }
}

// TODO: Delete it
export class Config {
  constructor(transformerConfig) {
    this.transformerConfig = transformerConfig;
  }

  static from(jsonObject) {
    return objectAssignDeep(new Config(), jsonObject);
  }
}

export const newConfig = () =>
  new Config(new TransformerConfig([new FeastConfig("", [], [])]));

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
  constructor() {
    this.tableName = "";
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

export class FromJson {
  constructor() {
    this.jsonPath = "";
    this.addRowNumber = false;
  }
}

export class FromTable {
  constructor() {
    this.tableName = "";
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
    this.leftTable = "";
    this.rightTable = "";
    this.outputTable = "";
    this.how = "";
  }
}

export class JsonOutput {
  constructor() {
    this.jsonTemplate = new JsonTemplate();
  }
}

export class JsonTemplate {
  constructor() {
    this.fields = undefined;
    this.data = undefined;
  }
}
