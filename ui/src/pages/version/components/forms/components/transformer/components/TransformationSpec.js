import React, { useContext } from "react";
import { EuiCodeBlock } from "@elastic/eui";
import { FormContext } from "@gojek/mlp-ui";
import "./TransformationSpec.scss";

const yaml = require("js-yaml");

export const TransformationSpec = () => {
  const {
    data: {
      transformer: { config }
    }
  } = useContext(FormContext);

  return (
    <EuiCodeBlock
      language="yaml"
      fontSize="s"
      paddingSize="s"
      overflowHeight={640}
      isCopyable>
      {config && yaml.dump(JSON.parse(JSON.stringify(config)))}
    </EuiCodeBlock>
  );
};
