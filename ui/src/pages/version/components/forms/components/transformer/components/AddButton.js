import React from "react";
import { EuiButton, EuiText, EuiToolTip } from "@elastic/eui";

export const AddButton = ({
  title,
  description,
  onClick,
  fullWidth = true,
  titleSize = "s",
  disabled = false
}) => {
  return (
    <EuiToolTip position="bottom" content={description}>
      <EuiButton
        onClick={onClick}
        size="s"
        fullWidth={fullWidth}
        disabled={disabled}>
        <EuiText size={titleSize}>{title}</EuiText>
      </EuiButton>
    </EuiToolTip>
  );
};
