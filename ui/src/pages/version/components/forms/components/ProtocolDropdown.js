import React, { Fragment } from "react";
import { EuiSuperSelect, EuiText } from "@elastic/eui";

/**
 * Dropdown menu for selecting protocol
 *
 * @param {*} endpointStatus current endpoint status
 * @param {*} protocol current selected protocol
 * @param {*} onProtocolChange callback to be made when protocol selection changed
 *
 */
export const ProtocolDropdown = ({
  endpointStatus,
  protocol,
  onProtocolChange,
}) => {
  // Users should not be able to change protocol when the model endpoint status is serving
  const isDisabled = (endpointStatus) => {
    return endpointStatus === "serving";
  };

  const onChange = (value) => {
    onProtocolChange(value);
  };

  return (
    <EuiSuperSelect
      fullWidth
      options={[
        {
          value: "HTTP_JSON",
          inputDisplay: "JSON over HTTP (REST)",
          dropdownDisplay: (
            <Fragment>
              <strong>JSON over HTTP (REST)</strong>
              <EuiText size="s" color="subdued">
                <p className="euiTextColor--subdued">
                  (Default) Expose model as HTTP service
                </p>
              </EuiText>
            </Fragment>
          ),
        },
        {
          value: "UPI_V1",
          inputDisplay: "Universal Prediction Interface (gRPC)",
          dropdownDisplay: (
            <Fragment>
              <strong>Universal Prediction Interface (gRPC)</strong>
              <EuiText size="s" color="subdued">
                <p className="euiTextColor--subdued">
                  Expose model as gRPC service compatible with UPI
                </p>
              </EuiText>
            </Fragment>
          ),
        },
      ]}
      valueOfSelected={protocol || "HTTP_JSON"}
      onChange={onChange}
      disabled={isDisabled(endpointStatus)}
      hasDividers
    />
  );
};
