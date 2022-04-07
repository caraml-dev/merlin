import React, { Fragment, useState } from "react";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiFieldText,
  EuiSwitch
} from "@elastic/eui";
import { SelectValueType } from "./components/table_inputs/SelectValueType";
import _uniqueId from "lodash/uniqueId";

export const JsonPathConfigInput = ({
  jsonPathConfig,
  identifier,
  onChangeHandler,
  errors = {}
}) => {
  const [checkBoxId] = useState(_uniqueId(`${identifier}-`));

  const getDefaultValue = () => {
    if (jsonPathConfig === undefined) {
      return "";
    }
    if (jsonPathConfig.defaultValue === undefined) {
      return "";
    }
    return jsonPathConfig.defaultValue;
  };

  const getValueType = () => {
    if (jsonPathConfig === undefined) {
      return "";
    }
    if (jsonPathConfig.valueType === undefined) {
      return "";
    }
    return jsonPathConfig.valueType;
  };
  const [showDefaultValueOpt, setShowDefaultValueOpt] = useState(
    getDefaultValue() !== ""
  );
  const onChange = (field, value) => {
    onChangeHandler({
      ...jsonPathConfig,
      [field]: value
    });
  };

  const setShowingDefaultValue = checked => {
    setShowDefaultValueOpt(checked);
    if (jsonPathConfig) {
      delete jsonPathConfig["defaultValue"];
      delete jsonPathConfig["valueType"];
      onChangeHandler(jsonPathConfig);
    }
  };

  return (
    <Fragment>
      <EuiFlexGroup direction="column" gutterSize="s">
        <EuiFlexItem>
          <EuiFieldText
            value={jsonPathConfig !== undefined ? jsonPathConfig.jsonPath : ""}
            placeholder="JSONPath"
            onChange={e => onChange("jsonPath", e.target.value)}
          />
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiSwitch
            id={`addDefaultValue-${checkBoxId}`}
            label="Add default value"
            color="subdued"
            checked={showDefaultValueOpt}
            onChange={e => setShowingDefaultValue(e.target.checked)}
          />
        </EuiFlexItem>
        {showDefaultValueOpt && (
          <>
            <EuiFlexItem>
              <EuiFormRow label="Default Value">
                <EuiFieldText
                  placeholder="Default"
                  value={getDefaultValue()}
                  onChange={e =>
                    onChange("defaultValue", e.target.value)
                  }></EuiFieldText>
              </EuiFormRow>
            </EuiFlexItem>
            <EuiFlexItem>
              <SelectValueType
                valueType={getValueType()}
                onChangeHandler={value => onChange("valueType", value)}
                displayInColumn={false}
                errors={errors}
              />
            </EuiFlexItem>
          </>
        )}
      </EuiFlexGroup>
    </Fragment>
  );
};
