import React from "react";
import { EuiButton, EuiText, EuiToolTip } from "@elastic/eui";

export const AddButton = ({
  title,
  description,
  onClick,
  fullWidth = true
}) => {
  return (
    <EuiToolTip position="bottom" content={description}>
      <EuiButton onClick={onClick} size="s" fullWidth={fullWidth}>
        <EuiText size="s">{title}</EuiText>
      </EuiButton>
    </EuiToolTip>
  );
};
