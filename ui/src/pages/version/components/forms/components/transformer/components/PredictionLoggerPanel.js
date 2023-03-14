import React from "react";
import { 
    EuiFieldText, 
    EuiFlexGroup, 
    EuiFlexItem, 
    EuiFormRow, 
    EuiSpacer, 
    EuiSwitch, 
    EuiText
} from "@elastic/eui";

import { useOnChangeHandler, FormLabelWithToolTip } from "@gojek/mlp-ui";
import { Panel } from "../../Panel";

export const PredictionLoggerPanel = ({ predictionLogger, onChangeHandler, errors = {} }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);
  return (
    <Panel>
        <EuiFlexGroup direction="column" gutterSize="m">
            <EuiSpacer size="s"/>
            <EuiFlexItem>
                <EuiFormRow
                    fullWidth
                    display="row">
                    <EuiSwitch
                        id="enable-prediction-logging"
                        label={
                            <EuiText size="s" color="subdued"> Enable Prediction Logging </EuiText>
                        }
                        checked={predictionLogger.enabled}
                        onChange={e => onChange("enabled")(e.target.checked)}
                    />
                </EuiFormRow>
            </EuiFlexItem>
            <EuiSpacer size="s"/>
            <EuiFlexItem>
                <EuiFormRow
                    fullWidth
                    label={
                        <FormLabelWithToolTip 
                            label="Raw Features Table"
                            content="Table that has pre transformed feature value"
                        />
                    }
                    error={errors.raw_features_table}
                    isInvalid={!!errors.raw_features_table}
                    display="columnCompressed"
                >
                    <EuiFieldText 
                        value={!!predictionLogger ? predictionLogger.raw_features_table : ""}
                        onChange={(e) => onChange("raw_features_table")(e.target.value)}
                    />
                </EuiFormRow>
            </EuiFlexItem>
            <EuiSpacer size="s"/>
            <EuiFlexItem>
                <EuiFormRow
                    fullWidth
                    label={
                        <FormLabelWithToolTip 
                            label="Entities Table"
                            content="Table that has entities information which are involve in the prediction process"
                        />
                    }
                    error={errors.entities_table}
                    isInvalid={!!errors.entities_table}
                    display="columnCompressed"
                >
                    <EuiFieldText 
                        value={!!predictionLogger ? predictionLogger.entities_table : ""}
                        onChange={(e) => onChange("entities_table")(e.target.value)}
                    />
                </EuiFormRow>
            </EuiFlexItem>
        </EuiFlexGroup>
    </Panel>
  );
};
