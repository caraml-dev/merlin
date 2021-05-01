import React from "react";
import {
  EuiButton,
  EuiCard,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
  EuiSpacer,
  EuiText
} from "@elastic/eui";
import { FormLabelWithToolTip } from "@gojek/mlp-ui";
import { TransformationStepCard } from "./TransformationStepCard";

export const TableTransformationCard = ({ errors = {} }) => {
  return (
    <EuiCard title="#1 - Table Transformation" titleSize="xs" textAlign="left">
      <EuiForm>
        <EuiFlexGroup direction="row">
          <EuiFlexItem>
            <EuiFormRow
              label={
                <FormLabelWithToolTip
                  label="Input Table"
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
                  label="Output Table"
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

        <EuiSpacer size="m" />

        <EuiFlexGroup direction="column" gutterSize="s">
          <EuiFlexItem>
            <TransformationStepCard />
            <EuiSpacer size="s" />
          </EuiFlexItem>
          <EuiFlexItem>
            <EuiButton onClick={() => {}} size="s">
              <EuiText size="s">+ Add Step</EuiText>
            </EuiButton>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiForm>
    </EuiCard>
  );
};
