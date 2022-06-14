import {
  EuiCodeBlock,
  EuiFilePicker,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiSpacer,
  EuiText
} from "@elastic/eui";
import { FormContext, useOnChangeHandler } from "@gojek/mlp-ui";
import React, { useContext } from "react";
import { Config } from "../../../../../../../services/transformer/TransformerConfig";

const yaml = require("js-yaml");

export const TransformationSpec = ({ importEnabled = true }) => {
  const {
    data: {
      transformer: { config }
    },
    onChangeHandler
  } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const onFilePickerChange = files => {
    if (files.length > 0) {
      var reader = new FileReader();

      reader.onload = (function() {
        return function(e) {
          const content = e.target.result;
          const transformerConfig = Config.fromJson(yaml.load(content));
          if (transformerConfig.validate()) {
            onChange("transformer.config")(transformerConfig);
          }
        };
      })(files[0]);

      reader.readAsText(files[0]);
    }
  };

  return (
    <EuiFlexGroup direction="column" gutterSize="s">
      {importEnabled && (
        <EuiFlexItem>
          <EuiFormRow label="Import Configuration" display="row" fullWidth>
            <EuiFilePicker
              initialPromptText={
                <EuiText size="xs">
                  Import Transformer YAML Specification
                </EuiText>
              }
              onChange={files => {
                onFilePickerChange(files);
              }}
              compressed
              fullWidth
            />
          </EuiFormRow>
        </EuiFlexItem>
      )}

      {config && (
        <EuiFlexItem grow={false}>
          <EuiSpacer size="m" />
          <EuiCodeBlock
            language="yaml"
            fontSize="s"
            paddingSize="s"
            overflowHeight={640}
            style={{ maxWidth: 480, maxHeight: 480 }}
            isCopyable>
            {yaml.dump(JSON.parse(JSON.stringify(config)))}
          </EuiCodeBlock>
        </EuiFlexItem>
      )}
    </EuiFlexGroup>
  );
};
