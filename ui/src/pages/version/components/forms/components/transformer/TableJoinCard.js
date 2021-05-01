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

export const TableJoinCard = ({ errors = {} }) => {
  return (
    <EuiCard title="#2 - Table Join" titleSize="xs" textAlign="left">
      <EuiForm>
        <EuiFlexGroup direction="row">
          <EuiFlexItem>
            <EuiFormRow
              label={
                <FormLabelWithToolTip
                  label="Column On Left"
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
                  label="Column On Right"
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

        <EuiFlexGroup direction="row">
          <EuiFlexItem>
            <EuiFormRow
              label={
                <FormLabelWithToolTip
                  label="Left Table"
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
                  label="Right Table"
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

        <EuiFlexGroup direction="row">
          <EuiFlexItem>
            <EuiFormRow
              label={
                <FormLabelWithToolTip
                  label="Output Table"
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
                  label="Join Operation"
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
