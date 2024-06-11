import { EuiCodeBlock, EuiDescriptionList, EuiFieldSearch } from "@elastic/eui";
import React, { Fragment, useState } from "react";
import { ConfigSectionPanelTitle } from "../../../components/section";
import { getBigQueryDashboardUrl } from "../utils/bigquery";
import { SourceFeatures } from "./SourceFeatures";

const BQSourceConfig = ({ job }) => {
  const [searchFeature, setSearchFeature] = useState("");

  const items = [
    { title: "Source Type", description: "BigQuery" },
    {
      title: "Table Name",
      description: (
        <a
          href={getBigQueryDashboardUrl(
            job.config.job_config.bigquerySource.table,
          )}
          target="_blank"
          rel="noopener noreferrer"
        >
          {job.config.job_config.bigquerySource.table}
        </a>
      ),
    },
    {
      title: "Options",
      description: job.config.job_config.bigquerySource.options ? (
        <EuiCodeBlock paddingSize="s">
          {JSON.stringify(
            job.config.job_config.bigquerySource.options,
            undefined,
            2,
          )}
        </EuiCodeBlock>
      ) : (
        "-"
      ),
    },
    {
      title: "Features",
      description: (
        <EuiFieldSearch
          compressed
          isClearable={false}
          onChange={(e) => setSearchFeature(e.target.value)}
          placeholder="Search feature name"
        />
      ),
    },
  ];

  return (
    <Fragment>
      <ConfigSectionPanelTitle title="Source Config" />

      <EuiDescriptionList
        compressed
        columnWidths={[2, 4]}
        type="responsiveColumn"
        listItems={items}
      />

      <SourceFeatures
        features={job.config.job_config.bigquerySource.features}
        searchFeature={searchFeature}
      />
    </Fragment>
  );
};

const SourceConfig = ({ job }) => {
  // TODO: when GCS is introduced we need to make condition to return the correct component
  return <BQSourceConfig job={job} />;
};

export default SourceConfig;
