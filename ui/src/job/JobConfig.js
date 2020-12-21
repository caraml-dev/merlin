/**
 * Copyright 2020 The Merlin Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React, { useEffect } from "react";
import {
  EuiFlexItem,
  EuiFlexGroup,
  EuiTextAlign,
  EuiLoadingChart,
  EuiCallOut,
  EuiCard,
  EuiDescriptionList,
  EuiDescriptionListTitle,
  EuiDescriptionListDescription
} from "@elastic/eui";
import mocks from "../mocks";
import { useMerlinApi } from "../hooks/useMerlinApi";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import PropTypes from "prop-types";

const JobConfig = ({ projectId, modelId, versionId, jobId }) => {
  const [{ data, isLoaded, error }] = useMerlinApi(
    `/models/${modelId}/versions/${versionId}/jobs/${jobId}`,
    { mock: mocks.job },
    []
  );

  const [model] = useMerlinApi(
    `/projects/${projectId}/models/${modelId}`,
    { mock: mocks.model },
    []
  );

  useEffect(() => {
    replaceBreadcrumbs([
      {
        text: "Models",
        href: `/merlin/projects/${projectId}/models`
      },
      {
        text: model.data.name,
        href: `/merlin/projects/${projectId}/models/${modelId}`
      },
      {
        text: "Versions",
        href: `/merlin/projects/${projectId}/models/${modelId}/versions`
      },
      {
        text: versionId,
        href: `/merlin/projects/${projectId}/models/${modelId}/versions/${versionId}`
      },
      {
        text: "Jobs",
        href: `/merlin/projects/${projectId}/models/${modelId}/versions/${versionId}/jobs`
      },
      {
        text: `${jobId}`,
        href: `/merlin/projects/${projectId}/models/${modelId}/versions/${versionId}/jobs/${jobId}`
      }
    ]);
  }, [projectId, modelId, versionId, jobId, model]);

  const ConfigEntry = ({ name, value }) => {
    return (
      <EuiDescriptionList style={{ marginBottom: "8px" }}>
        <EuiDescriptionListTitle>{name}</EuiDescriptionListTitle>
        <EuiDescriptionListDescription>{value}</EuiDescriptionListDescription>
      </EuiDescriptionList>
    );
  };

  const ConfigEntryLink = ({ name, value, url }) => {
    return (
      <EuiDescriptionList style={{ marginBottom: "8px" }}>
        <EuiDescriptionListTitle>{name}</EuiDescriptionListTitle>
        <a href={url} target="_blank" rel="noopener noreferrer">
          <EuiDescriptionListDescription>{value}</EuiDescriptionListDescription>
        </a>
      </EuiDescriptionList>
    );
  };

  function getGcsDashboardUrl(gcsBucketUri) {
    const uri = gcsBucketUri.replace("gs://", "");
    return `https://console.cloud.google.com/storage/browser/${uri}/?authuser=1`;
  }

  function getBigQueryDashboardUrl(tableId) {
    const segment = tableId.split(".");
    const project = segment[0];
    const dataset = segment[1];
    const table = segment[2];
    return `https://console.cloud.google.com/bigquery?authuser=1&project=${project}&p=${project}&d=${dataset}&t=${table}&page=table`;
  }

  const SourceConfig = ({ job }) => {
    const BQSourceConfig = ({ job }) => {
      return (
        <EuiCard textAlign="left" title="Source Config" description="">
          <ConfigEntry name="Source Type" value="BigQuery" />
          <ConfigEntryLink
            name="Table Name"
            value={job.config.job_config.bigquerySource.table}
            url={getBigQueryDashboardUrl(
              job.config.job_config.bigquerySource.table
            )}
          />
          <ConfigEntry
            name="Features"
            value={job.config.job_config.bigquerySource.features.join(", ")}
          />
          <ConfigEntry
            name="Options"
            value={JSON.stringify(
              job.config.job_config.bigquerySource.options,
              undefined,
              2
            )}
          />
        </EuiCard>
      );
    };

    // TODO: when GCS is introduced we need to make condition to return the correct component
    return <BQSourceConfig job={job} />;
  };

  function getResultType(result) {
    if (!result) {
      return "DOUBLE";
    }
    let resultType = result.type;

    if (!resultType) {
      resultType = "DOUBLE";
    }
    if (resultType === "ARRAY") {
      let itemType = result.item_type;
      if (!itemType) {
        itemType = "DOUBLE";
      }
      return "ARRAY OF " + itemType;
    }

    return resultType;
  }

  const SinkConfig = ({ job }) => {
    const BQSinkConfig = ({ job }) => {
      return (
        <EuiCard textAlign="left" title="Sink Config" description="">
          <ConfigEntry name="Sink Type" value="BigQuery" />
          <ConfigEntryLink
            name="Table Name"
            value={job.config.job_config.bigquerySink.table}
            url={getBigQueryDashboardUrl(
              job.config.job_config.bigquerySink.table
            )}
          />
          <ConfigEntry
            name="GCS Staging Bucket"
            value={job.config.job_config.bigquerySink.stagingBucket}
          />
          <ConfigEntry
            name="Result Column Name"
            value={job.config.job_config.bigquerySink.resultColumn}
          />
          <ConfigEntry
            name="Save Mode"
            value={job.config.job_config.bigquerySink.saveMode}
          />
          <ConfigEntry
            name="Options"
            value={JSON.stringify(
              job.config.job_config.bigquerySink.options,
              undefined,
              2
            )}
          />
          <ConfigEntry
            name="Result Type"
            value={getResultType(data.config.job_config.model.result)}
          />
        </EuiCard>
      );
    };

    // TODO: when GCS is introduced we need to make condition to return the correct component
    return <BQSinkConfig job={job} />;
  };

  return !isLoaded ? (
    <EuiTextAlign textAlign="center">
      <EuiLoadingChart size="xl" mono />
    </EuiTextAlign>
  ) : error ? (
    <EuiCallOut
      title="Sorry, there was an error"
      color="danger"
      iconType="alert">
      <p>{error.message}</p>
    </EuiCallOut>
  ) : (
    <EuiFlexGroup direction="column">
      <EuiFlexItem>
        <EuiFlexGroup gutterSize="l">
          <EuiFlexItem>
            <EuiCard textAlign="left" title="Job Info" description="">
              <ConfigEntry name="Job Id" value={data.id} />
              <ConfigEntry name="Job Name" value={data.name} />
              <ConfigEntry name="Model Name" value={model.data.name} />
              <ConfigEntry name="Model Version" value={data.version_id} />
              <ConfigEntry
                name="Model Type"
                value={data.config.job_config.model.type}
              />
              <ConfigEntryLink
                name="Model Artifact URI"
                value={data.config.job_config.model.uri}
                url={getGcsDashboardUrl(data.config.job_config.model.uri)}
              />
            </EuiCard>
          </EuiFlexItem>
          <EuiFlexItem>
            <EuiCard textAlign="left" title="Job Status" description="">
              <ConfigEntry name="Status" value={data.status} />
              <ConfigEntry name="Error" value={data.error} />
            </EuiCard>
          </EuiFlexItem>
          <EuiFlexItem>
            <EuiCard textAlign="left" title="Miscellaneous Info" description="">
              <ConfigEntry
                name="Service Account"
                value={data.config.service_account_name}
              />
              <ConfigEntry name="Image" value={data.config.image_ref} />
            </EuiCard>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiFlexItem>
      <EuiFlexItem>
        <EuiFlexGroup>
          <EuiFlexItem>
            <SourceConfig job={data} />
          </EuiFlexItem>
          <EuiFlexItem>
            <SinkConfig job={data} />
          </EuiFlexItem>

          <EuiFlexItem>
            <EuiFlexGroup direction="column">
              <EuiFlexItem>
                <EuiCard
                  textAlign="left"
                  title="Resource Requests"
                  description="">
                  <ConfigEntry
                    name="Driver CPU"
                    value={data.config.resource_request.driver_cpu_request}
                  />
                  <ConfigEntry
                    name="Driver Memory"
                    value={data.config.resource_request.driver_memory_request}
                  />
                  <ConfigEntry
                    name="Executor Replica Number"
                    value={data.config.resource_request.executor_replica}
                  />
                  <ConfigEntry
                    name="Executor CPU"
                    value={data.config.resource_request.executor_cpu_request}
                  />
                  <ConfigEntry
                    name="Executor Memory"
                    value={data.config.resource_request.executor_memory_request}
                  />
                </EuiCard>
              </EuiFlexItem>

              {data.config.env_vars && (
                <EuiFlexItem>
                  <EuiCard
                    textAlign="left"
                    title="Environment Variables"
                    description="">
                    {data.config.env_vars.map((variable, i) => (
                      <ConfigEntry
                        key={i}
                        name={variable.name}
                        value={variable.value}
                      />
                    ))}
                  </EuiCard>
                </EuiFlexItem>
              )}
            </EuiFlexGroup>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};

JobConfig.propTypes = {
  projectId: PropTypes.string,
  modelId: PropTypes.string,
  versionId: PropTypes.string,
  jobId: PropTypes.string
};

export default JobConfig;
