import React from "react";
import { EuiCard } from "@elastic/eui";
import { VariablesPanel } from "./Variables";

export const VariableInputCard = () => {
  return (
    <EuiCard title="Variable Declaration" titleSize="xs" textAlign="left">
      <VariablesPanel variables={[]} onChangeHandler={() => {}} errors={{}} />
    </EuiCard>
  );
};
