import React from "react";
import {
  EuiButton,
  EuiCard,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
  EuiSpacer
} from "@elastic/eui";
import { FormLabelWithToolTip } from "@gojek/mlp-ui";

export const TransformationStepCard = ({ errors = {} }) => {
  return (
    <EuiCard title="Step 1" titleSize="xs" textAlign="left" display="subdued">
      <EuiForm>
        <EuiFlexGroup direction="row">
          <EuiFlexItem>
            <EuiFormRow
              label={
                <FormLabelWithToolTip
                  label="Operation"
                  content="Specify the total amount of CPU available for the component"
                />
              }
              isInvalid={!!errors.cpu_request}
              error={errors.cpu_request}
              fullWidth>
              <EuiFieldText
                fullWidth
                value=""
                onChange={() => {}}
                isInvalid={!!errors.cpu_request}
                name="cpu"
              />
            </EuiFormRow>
          </EuiFlexItem>

          <EuiFlexItem>
            <EuiFormRow
              label={
                <FormLabelWithToolTip
                  label="Column Name"
                  content="Specify the total amount of RAM available for the component"
                />
              }
              isInvalid={!!errors.memory_request}
              error={errors.memory_request}
              fullWidth>
              <EuiFieldText
                fullWidth
                value=""
                onChange={() => {}}
                isInvalid={!!errors.memory_request}
                name="memory"
              />
            </EuiFormRow>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiForm>
    </EuiCard>
  );
};
