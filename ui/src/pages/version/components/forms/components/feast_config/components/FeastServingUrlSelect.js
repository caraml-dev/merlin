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

export const FeastServingUrlSelect = ({
  servingUrl,
  servingSource,
  onChange,
  isInvalid
}) => {
  const confByUrl = appConfig.feastServingUrls.reduce(
    (res, val) => ({ ...res, [val.host]: val }),
    {}
  );
  const confBySource = appConfig.feastServingUrls.reduce(
    (res, val) => ({ ...res, [val.source_type]: val }),
    {}
  );

  // find serving source if it is not set before
  if (!servingSource || servingSource === "") {
    if (servingUrl !== "") {
      servingSource = confByUrl[servingUrl].source_type;
    } else {
      servingSource = appConfig.defaultFeastSource;
    }
  }

  const options = appConfig.feastServingUrls.map(url => {
    return {
      value: url,
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
      valueOfSelected={confBySource[servingSource] || {}}
      onChange={servingConf => {
        onChange("servingUrl", servingConf.host);
        onChange("source", servingConf.source_type);
      }}
      placeholder="Select Feast Serving Url"
      isInvalid={isInvalid}
      hasDividers
      fullWidth
    />
  );
};
