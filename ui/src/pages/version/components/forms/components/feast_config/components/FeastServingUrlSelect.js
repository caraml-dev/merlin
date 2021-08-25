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

export const FeastServingUrlSelect = ({ servingUrl, onChange, isInvalid }) => {
  const options = appConfig.feastServingUrls.map(url => {
    return {
      value: url.host,
      inputDisplay: `${url.label} (${url.host})`,
      dropdownDisplay: (
        <EuiFlexGroup alignItems="center" gutterSize="m">
          <EuiFlexItem grow={false}>
            <EuiIcon size="l" type={logoMaps[url.icon]} />
          </EuiFlexItem>
          <EuiFlexItem>
            <strong>{url.label}</strong>
            <EuiText size="xs" color="subdued">
              <p className="euiTextColor--subdued">{url.host}</p>
            </EuiText>
          </EuiFlexItem>
        </EuiFlexGroup>
      )
    };
  });

  return (
    <EuiSuperSelect
      options={options}
      valueOfSelected={servingUrl || ""}
      onChange={onChange}
      placeholder="Select Feast Serving Url"
      isInvalid={isInvalid}
      hasDividers
      fullWidth
    />
  );
};
