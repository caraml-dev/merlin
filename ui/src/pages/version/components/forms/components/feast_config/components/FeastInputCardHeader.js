import React from "react";
import {
  EuiButtonIcon,
  EuiFlexGroup,
  EuiFlexItem,
  EuiIcon
} from "@elastic/eui";

export const FeastInputCardHeader = ({ onDelete, dragHandleProps }) => (
  <EuiFlexGroup
    {...dragHandleProps}
    justifyContent="spaceBetween"
    alignItems="center"
    gutterSize="none"
    style={{ backgroundColor: "ghostwhite" }}
    direction="row">
    <EuiFlexItem grow={false}>
      <EuiIcon type="empty" size="l" />
    </EuiFlexItem>
    <EuiFlexItem grow={false}>
      <EuiIcon type="grab" color="subdued" size="m" />
    </EuiFlexItem>
    <EuiFlexItem grow={false}>
      {!!onDelete ? (
        <EuiButtonIcon
          iconType="cross"
          onClick={onDelete}
          aria-label="delete-route"
        />
      ) : (
        <EuiIcon type="empty" size="l" />
      )}
    </EuiFlexItem>
  </EuiFlexGroup>
);
