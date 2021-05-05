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
  }
}

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
