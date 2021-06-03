import React from "react";
import { EuiButton, EuiText, EuiToolTip } from "@elastic/eui";

export const AddButton = ({
  title,
  description,
  onClick,
  fullWidth = true,
  titleSize = "s"
}) => {
  return (
    <EuiToolTip position="bottom" content={description}>
      <EuiButton onClick={onClick} size="s" fullWidth={fullWidth}>
        <EuiText size={titleSize}>{title}</EuiText>
      </EuiButton>
    </EuiToolTip>
  );
};
