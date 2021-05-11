import { Config } from "./TransformerConfig";

const fs = require("fs");
const path = require("path");
const yaml = require("js-yaml");

test("test Config.fromJson()", () => {
  const content = fs
    .readFileSync(path.join(__dirname + "/testdata/standard_transformer.yaml"))
    .toString();
  const config = Config.fromJson(yaml.load(content));

  expect(config.transformerConfig).toBeDefined();

  expect(config.transformerConfig.feast).toBeUndefined();
  expect(config.transformerConfig.preprocess).toBeDefined();
  expect(config.transformerConfig.postprocess).toBeDefined();
});
