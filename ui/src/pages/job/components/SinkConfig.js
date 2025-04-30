import { EuiCodeBlock, EuiDescriptionList } from "@elastic/eui";
import React, { Fragment } from "react";
import { ConfigSectionPanelTitle } from "../../../components/section";
import { getBigQueryDashboardUrl } from "../utils/bigquery";

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

const BQSinkConfig = ({ job }) => {
  const items = [
    { title: "Sink Type", description: "BigQuery" },
    {
      title: "Table Name",
      description: (
        <a
          href={getBigQueryDashboardUrl(
            job.config.job_config.bigquerySink.table
          )}
          target="_blank"
          rel="noopener noreferrer"
        >
          {job.config.job_config.bigquerySink.table}
        </a>
      ),
    },
    {
      title: "GCS Staging Bucket",
      description: job.config.job_config.bigquerySink.stagingBucket,
    },
    {
      title: "Result Column Name",
      description: job.config.job_config.bigquerySink.resultColumn,
    },
    {
      title: "Save Mode",
      description: job.config.job_config.bigquerySink.saveMode,
    },
    {
      title: "Result Type",
      description: getResultType(job.config.job_config.model.result),
    },
    {
      title: "Options",
      description: job.config.job_config.bigquerySink.options ? (
        <EuiCodeBlock paddingSize="s">
          {JSON.stringify(
            job.config.job_config.bigquerySink.options,
            undefined,
            2
          )}
        </EuiCodeBlock>
      ) : (
        "-"
      ),
    },
  ];

  return (
    <Fragment>
      <ConfigSectionPanelTitle title="Sink Config" />

      <EuiDescriptionList
        compressed
        columnWidths={[2, 4]}
        type="responsiveColumn"
        listItems={items}
      />
    </Fragment>
  );
};

const MCSinkConfig = ({ job }) => {
  const items = [
    { title: "Sink Type", description: "MaxCompute" },
    {
      title: "Table Name",
      description: job.config.job_config.maxcomputeSink.table,
    },
    {
      title: "Endpoint",
      description: job.config.job_config.maxcomputeSource.endpoint
    },
    {
      title: "Result Column Name",
      description: job.config.job_config.maxcomputeSink.resultColumn,
    },
    {
      title: "Save Mode",
      description: job.config.job_config.maxcomputeSink.saveMode,
    },
    {
      title: "Result Type",
      description: getResultType(job.config.job_config.model.result),
    },
    {
      title: "Options",
      description: job.config.job_config.maxcomputeSink.options ? (
        <EuiCodeBlock paddingSize="s">
          {JSON.stringify(
            job.config.job_config.maxcomputeSink.options,
            undefined,
            2
          )}
        </EuiCodeBlock>
      ) : (
        "-"
      ),
    },
  ];

  return (
    <Fragment>
      <ConfigSectionPanelTitle title="Sink Config" />

      <EuiDescriptionList
        compressed
        columnWidths={[2, 4]}
        type="responsiveColumn"
        listItems={items}
      />
    </Fragment>
  );
};

const SinkConfig = ({ job }) => {
  return (
      job.config.job_config.bigquerySink ?
          <BQSinkConfig job={job} /> :
          <MCSinkConfig job={job} />
  );
};

export default SinkConfig;
