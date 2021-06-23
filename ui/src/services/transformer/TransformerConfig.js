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

import * as yup from "yup";
import {
  feastInputSchema,
  pipelineSchema
} from "../../pages/version/components/forms/validation/schema";

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

    if (!config.transformerConfig.preprocess) {
      config.transformerConfig.preprocess = new Pipeline();
    }
    if (!config.transformerConfig.postprocess) {
      config.transformerConfig.postprocess = new Pipeline();
    }

    return config;
  }

  validate() {
    let obj = objectAssignDeep({}, this);
    if (obj.feast) {
      if (!yup.array(feastInputSchema).isValidSync(obj.feast)) {
        return false;
      }
    }
    if (obj.preprocess) {
      if (!pipelineSchema.isValidSync(obj.preprocess)) {
        return false;
      }
    }
    if (obj.postprocess) {
      if (!pipelineSchema.isValidSync(obj.postprocess)) {
        return false;
      }
    }
    return true;
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
              /* For backward compatibility */
              entity["fieldType"] = "UDF";
              entity["field"] = entity.udf;
            } else if (entity.expression) {
              entity["fieldType"] = "Expression";
              entity["field"] = entity.expression;
            } else if (entity.jsonPath) {
              entity["fieldType"] = "JSONPath";
              entity["field"] = entity.jsonPath;
            }
          });
      });

    if (transformerConfig.preprocess) {
      transformerConfig.preprocess = Pipeline.fromJson(
        transformerConfig.preprocess
      );
    }

    if (transformerConfig.postprocess) {
      transformerConfig.postprocess = Pipeline.fromJson(
        transformerConfig.postprocess
      );
    }

    return transformerConfig;
  }

  toJSON() {
    let obj = objectAssignDeep({}, this);

    obj.feast &&
      obj.feast.forEach(feast => {
        feast.entities &&
          feast.entities.forEach(entity => {
            if (entity.fieldType === "UDF") {
              /* For backward compatibility */
              entity["udf"] = entity.field;
            } else if (entity.fieldType === "Expression") {
              entity["expression"] = entity.field;
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

  static fromJson(json) {
    const pipeline = objectAssignDeep(new Pipeline(), json);

    pipeline.inputs.forEach(input => {
      input.feast &&
        input.feast.forEach(feast => {
          feast.entities &&
            feast.entities.forEach(entity => {
              if (entity.udf) {
                /* For backward compatibility */
                entity["fieldType"] = "UDF";
                entity["field"] = entity.udf;
              } else if (entity.expression) {
                entity["fieldType"] = "Expression";
                entity["field"] = entity.expression;
              } else if (entity.jsonPath) {
                entity["fieldType"] = "JSONPath";
                entity["field"] = entity.jsonPath;
              }
            });
        });

      input.tables &&
        input.tables.forEach(table => {
          table.columns &&
            table.columns.forEach(column => {
              if (column.fromJson) {
                column["type"] = "jsonpath";
                column["value"] = column.fromJson.jsonPath;
              } else if (column.expression) {
                column["type"] = "expression";
                column["value"] = column.expression;
              } else if (column.literal) {
                if (column.literal.stringValue) {
                  column["type"] = "string";
                  column["value"] = column.literal.stringValue;
                } else if (column.literal.intValue) {
                  column["type"] = "int";
                  column["value"] = parseInt(column.literal.intValue);
                } else if (column.literal.floatValue) {
                  column["type"] = "float";
                  column["value"] = parseFloat(column.literal.floatValue);
                } else if (column.literal.boolValue !== undefined) {
                  column["type"] = "bool";
                  column["value"] = column.literal.boolValue.toString();
                }
              }
            });
        });

      input.variables &&
        input.variables.forEach(variable => {
          if (variable.jsonPath !== undefined && variable.jsonPath !== "") {
            variable["type"] = "jsonpath";
            variable["value"] = variable.jsonPath;
          } else if (
            variable.expression !== undefined &&
            variable.expression !== ""
          ) {
            variable["type"] = "expression";
            variable["value"] = variable.expression;
          } else if (variable.literal) {
            if (
              variable.literal.stringValue !== undefined &&
              variable.literal.stringValue !== ""
            ) {
              variable["type"] = "string";
              variable["value"] = variable.literal.stringValue;
            } else if (
              variable.literal.intValue !== undefined &&
              variable.literal.intValue !== 0
            ) {
              variable["type"] = "int";
              variable["value"] = variable.literal.intValue;
            } else if (
              variable.literal.floatValue !== undefined &&
              variable.literal.floatValue !== 0
            ) {
              variable["type"] = "float";
              variable["value"] = variable.literal.floatValue;
            } else if (variable.literal.boolValue !== undefined) {
              variable["type"] = "bool";
              variable["value"] = variable.literal.boolValue;
            }
          }
        });
    });

    pipeline.transformations.forEach(transformation => {
      transformation.tableTransformation &&
        transformation.tableTransformation.steps &&
        transformation.tableTransformation.steps.forEach(step => {
          if (step.dropColumns !== undefined) {
            step["operation"] = "dropColumns";
          } else if (step.renameColumns !== undefined) {
            step["operation"] = "renameColumns";
          } else if (step.selectColumns !== undefined) {
            step["operation"] = "selectColumns";
          } else if (step.sort !== undefined) {
            step["operation"] = "sort";
          } else if (step.updateColumns !== undefined) {
            step["operation"] = "updateColumns";
          }
        });
    });

    return pipeline;
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

            feast.entities &&
              feast.entities.forEach(entity => {
                if (entity.fieldType === "UDF") {
                  /* For backward compatibility */
                  entity["udf"] = entity.field;
                } else if (entity.fieldType === "Expression") {
                  entity["expression"] = entity.field;
                } else {
                  entity["jsonPath"] = entity.field;
                }
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
    } else {
      obj.transformations.forEach(transformation => {
        transformation.tableTransformation &&
          transformation.tableTransformation.steps &&
          transformation.tableTransformation.steps.forEach(step => {
            delete step["operation"];
          });
      });
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
