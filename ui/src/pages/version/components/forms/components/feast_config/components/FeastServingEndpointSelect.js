import React from "react";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiIcon,
  EuiSuperSelect,
  EuiText
} from "@elastic/eui";
import { appConfig } from "../../../../../../../config";

import logoBigtable from "../../../../../../../assets/icon/gcp-bigtable.svg";

const logoMaps = {
  redis: "logoRedis",
  bigtable: logoBigtable
};

export const FeastServingEndpointSelect = ({
  servingEndpoint,
  onChange,
  isInvalid
}) => {
  const options = appConfig.feastServingEndpoints.map(endpoint => {
    return {
      value: endpoint.host,
      inputDisplay: `${endpoint.label} (${endpoint.host})`,
      dropdownDisplay: (
        <EuiFlexGroup alignItems="center" gutterSize="m">
          <EuiFlexItem grow={false}>
            <EuiIcon size="l" type={logoMaps[endpoint.icon]} />
          </EuiFlexItem>
          <EuiFlexItem>
            <strong>{endpoint.label}</strong>
            <EuiText size="xs" color="subdued">
              <p className="euiTextColor--subdued">{endpoint.host}</p>
            </EuiText>
          </EuiFlexItem>
        </EuiFlexGroup>
      )
    };
  });

  return (
    <EuiSuperSelect
      options={options}
      valueOfSelected={servingEndpoint || ""}
      onChange={onChange}
      placeholder="Select Feast Serving Endpoint"
      isInvalid={isInvalid}
      hasDividers
      fullWidth
    />
  );
};
