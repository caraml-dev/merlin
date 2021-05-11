import React from "react";
import {
  EuiForm,
  EuiFormRow,
  EuiIcon,
  EuiSuperSelect,
  EuiText,
  EuiToolTip
} from "@elastic/eui";
import { useOnChangeHandler } from "@gojek/mlp-ui";
import { Panel } from "./Panel";

export const LoggerPanel = ({ loggerConfig, onChangeHandler, errors = {} }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const setValue = value => {
    if (value === "disabled") {
      onChange("enabled")(false);
      onChange("mode")("");
    } else {
      onChange("enabled")(true);
      onChange("mode")(value);
    }
  };

  const loggerModes = [
    {
      name: "Disabled",
      value: "disabled",
      desc: "No logging"
    },
    {
      name: "All",
      value: "all",
      desc: "Log request and response"
    },
    {
      name: "Request",
      value: "request",
      desc: "Only log request"
    },
    {
      name: "Response",
      value: "response",
      desc: "Only log response"
    }
  ];

  const loggerModeOptions = loggerModes.map(mode => {
    return {
      value: mode.value,
      inputDisplay: mode.name,
      dropdownDisplay: (
        <>
          <strong>{mode.name}</strong>
          <EuiText size="s" color="subdued">
            <p className="euiTextColor--subdued">{mode.desc}</p>
          </EuiText>
        </>
      )
    };
  });

  return (
    <Panel title="Logger">
      <EuiForm>
        <EuiFormRow
          fullWidth
          label={
            <EuiToolTip content="Specify the logger mode for your deployment">
              <span>
                Logger Mode *{" "}
                <EuiIcon type="questionInCircle" color="subdued" />
              </span>
            </EuiToolTip>
          }
          isInvalid={!!errors.mode}
          error={errors.mode}
          display="row">
          <EuiSuperSelect
            fullWidth
            options={loggerModeOptions}
            valueOfSelected={loggerConfig.mode || "disabled"}
            onChange={value => setValue(value)}
            isInvalid={!!errors.logger}
            itemLayoutAlign="top"
            hasDividers
          />
        </EuiFormRow>
      </EuiForm>
    </Panel>
  );
};
