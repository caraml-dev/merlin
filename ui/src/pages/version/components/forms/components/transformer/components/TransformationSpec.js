import React, { useContext } from "react";
import {
  EuiCodeBlock,
  EuiFilePicker,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiIcon,
  EuiToolTip
} from "@elastic/eui";
import { FormContext, useOnChangeHandler } from "@gojek/mlp-ui";
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
      <EuiFlexItem>
        <EuiCodeBlock
          language="yaml"
          fontSize="s"
          paddingSize="s"
          overflowHeight={640}
          style={{ maxWidth: 480 }}
          isCopyable>
          {config && yaml.dump(JSON.parse(JSON.stringify(config)))}
        </EuiCodeBlock>
      </EuiFlexItem>

      {importEnabled && (
        <EuiFlexItem>
          <EuiFormRow
            label={
              <EuiToolTip content="This is beta feature. Proceed with caution 😉">
                <span>
                  Import <EuiIcon type="questionInCircle" color="subdued" />
                </span>
              </EuiToolTip>
            }
            display="row"
            fullWidth>
            <EuiFilePicker
              initialPromptText="Import Transformer YAML Specification"
              onChange={files => {
                onFilePickerChange(files);
              }}
              compressed
              fullWidth
            />
          </EuiFormRow>
        </EuiFlexItem>
      )}
    </EuiFlexGroup>
  );
};
