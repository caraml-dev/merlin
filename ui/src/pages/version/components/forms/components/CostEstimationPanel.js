import React from "react";
import {
  EuiSpacer,
  EuiCallOut,
  EuiHorizontalRule,
  EuiFlexGroup,
  EuiFlexItem,
  EuiText,
} from "@elastic/eui";
import { calculateCost } from "../../../../../utils/costEstimation";

/**
 * Panel for showing cost estimation
 * @param {VersionEndpoint} versionEndpoint version endpoint to be deployed
 * @returns
 */
export const CostEstimationPanel = ({ versionEndpoint }) => {
  const modelMinCost = calculateCost(
    versionEndpoint.resource_request.min_replica,
    versionEndpoint.resource_request.cpu_request,
    versionEndpoint.resource_request.memory_request
  );

  const modelMaxCost = calculateCost(
    versionEndpoint.resource_request.max_replica,
    versionEndpoint.resource_request.cpu_request,
    versionEndpoint.resource_request.memory_request
  );

  const transformerMinCost =
    versionEndpoint.transformer && versionEndpoint.transformer.enabled
      ? calculateCost(
          versionEndpoint.transformer.resource_request.min_replica,
          versionEndpoint.transformer.resource_request.cpu_request,
          versionEndpoint.transformer.resource_request.memory_request
        )
      : 0;

  const transformerMaxCost =
    versionEndpoint.transformer && versionEndpoint.transformer.enabled
      ? calculateCost(
          versionEndpoint.transformer.resource_request.max_replica,
          versionEndpoint.transformer.resource_request.cpu_request,
          versionEndpoint.transformer.resource_request.memory_request
        )
      : 0;

  const totalMinCost = modelMinCost + transformerMinCost;
  const totalMaxCost = modelMaxCost + transformerMaxCost;

  return (
    <EuiCallOut title="Cost Estimation" iconType="tag">
      <EuiSpacer size="m" />
      <div>
        <EuiFlexGroup>
          <EuiFlexItem>
            <h4>Model</h4>
          </EuiFlexItem>
          <EuiFlexItem style={{ textAlign: "right" }}>
            {" "}
            <h4>
              ${modelMinCost.toFixed(2)}-{modelMaxCost.toFixed(2)} / Month
            </h4>
          </EuiFlexItem>
        </EuiFlexGroup>
        <EuiSpacer size="s" />
        <EuiText>
          {" "}
          {versionEndpoint.resource_request.min_replica}-
          {versionEndpoint.resource_request.max_replica} Replica
        </EuiText>
        <EuiText>{versionEndpoint.resource_request.cpu_request} CPU</EuiText>
        <EuiText>
          {versionEndpoint.resource_request.memory_request} Memory
        </EuiText>
        <EuiSpacer size="m" />
      </div>

      {/* Don't show transformer cost estimate if it's disabled */}
      {versionEndpoint.transformer && versionEndpoint.transformer.enabled && (
        <div>
          <EuiFlexGroup>
            <EuiFlexItem>
              <h4>Transformer</h4>
            </EuiFlexItem>
            <EuiFlexItem style={{ textAlign: "right" }}>
              {" "}
              <h4>
                ${transformerMinCost.toFixed(2)}-{transformerMaxCost.toFixed(2)}{" "}
                / Month
              </h4>
            </EuiFlexItem>
          </EuiFlexGroup>
          <EuiSpacer size="s" />
          <EuiText>
            {" "}
            {versionEndpoint.transformer.resource_request.min_replica}-
            {versionEndpoint.transformer.resource_request.max_replica} Replica
          </EuiText>
          <EuiText>
            {versionEndpoint.transformer.resource_request.cpu_request} CPU
          </EuiText>
          <EuiText>
            {versionEndpoint.transformer.resource_request.memory_request} Memory
          </EuiText>
        </div>
      )}

      <EuiHorizontalRule />
      <EuiFlexGroup>
        <EuiFlexItem>
          <h4>Total</h4>
        </EuiFlexItem>
        <EuiFlexItem style={{ textAlign: "right" }}>
          <h4>
            ${totalMinCost.toFixed(2)}-{totalMaxCost.toFixed(2)} / Month
          </h4>
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiCallOut>
  );
};
