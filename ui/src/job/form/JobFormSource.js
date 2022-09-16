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

import React, { Fragment, useContext, useState } from "react";
import {
  EuiAccordion,
  EuiCode,
  EuiFieldNumber,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
  EuiSpacer,
  EuiSuperSelect,
  EuiSwitch,
  EuiText
} from "@elastic/eui";
import Editor from "@monaco-editor/react"
import { validateBigqueryTable } from "../../validation/validateBigquery";
import { JobFormContext } from "./context";
import { FeatureComboBox } from "./components/FeaturesComboBox";

export const JobFormSource = () => {
  const {
    job,
    setBigquerySource,
    setBigquerySourceOptions,
    unsetBigquerySourceOptions
  } = useContext(JobFormContext);

  const [isValidTable, setValidTable] = useState(
    job.config.job_config.bigquerySource.table
      ? validateBigqueryTable(job.config.job_config.bigquerySource.table)
      : true
  );

  const sourceTypeOptions = [
    {
      value: "bigquery",
      inputDisplay: "Google BigQuery",
      dropdownDisplay: <strong>Google BigQuery</strong>
    },
    {
      value: "gcs",
      inputDisplay: "Google Cloud Storage",
      dropdownDisplay: <strong>Google Cloud Storage (coming soon)</strong>,
      disabled: true
    }
  ];

  const readDataFormatOptions = [
    {
      value: "AVRO",
      inputDisplay: "AVRO",
      dropdownDisplay: <strong>AVRO</strong>
    },
    {
      value: "ARROW",
      inputDisplay: "ARROW",
      dropdownDisplay: (
        <Fragment>
          <strong>ARROW</strong>
          <EuiText size="xs">
            <p className="euiTextColor--subdued">
              Unsupported Arrow filters are not pushed down and results are
              filtered later by Spark. Currently Arrow does not suport
              disjunction across columns.
            </p>
          </EuiText>
        </Fragment>
      )
    }
  ];

  return (
    <Fragment>
      <EuiForm>
        <EuiFlexGroup>
          <EuiFlexItem>
            <EuiFormRow
              fullWidth
              label="Select Data Source Type *"
              helpText="Currently only support Google BigQuery.">
              <EuiSuperSelect
                fullWidth
                options={sourceTypeOptions}
                valueOfSelected="bigquery"
                onChange={() => {}}
                hasDividers
              />
            </EuiFormRow>
          </EuiFlexItem>

          <EuiFlexItem>
            <EuiFormRow
              fullWidth
              label="Data Source Table *"
              helpText="Specify the location of BigQuery source table.">
              <EuiFieldText
                fullWidth
                placeholder="Format: project_name.dataset_name.table_name"
                value={job.config.job_config.bigquerySource.table}
                onChange={e => {
                  const val = e.target.value;
                  setValidTable(validateBigqueryTable(val));
                  setBigquerySource("table", val);
                }}
                isInvalid={!isValidTable}
                aria-label="Data Source Table"
              />
            </EuiFormRow>

            <EuiFormRow fullWidth>
              <EuiSwitch
                label="The data source table is a BigQuery view."
                checked={
                  job.config.job_config.bigquerySource.options.viewsEnabled ===
                  "true"
                }
                onChange={e =>
                  setBigquerySourceOptions("viewsEnabled", e.target.checked)
                }
              />
            </EuiFormRow>
          </EuiFlexItem>
        </EuiFlexGroup>

        <EuiSpacer size="l" />

        <FeatureComboBox
          options={
            job.config.job_config.bigquerySource.features
              ? job.config.job_config.bigquerySource.features.map(feature => {
                  return { label: feature };
                })
              : []
          }
          onChange={selected => {
            setBigquerySource("features", selected);
          }}
        />

        <EuiSpacer size="l" />

        <EuiFlexGroup>
          <EuiFlexItem>
            <EuiFormRow
              fullWidth
              label="Data Source Table Filter"
              helpText={
                <p>
                  Specified the <EuiCode>WHERE</EuiCode> conditions to filter
                  the data source table. Example:{" "}
                  <EuiCode>_PARTITIONDATE > '2020-01-01'</EuiCode>.
                </p>
              }>
              <Editor
                width="100%"
                height="100px"
                aria-label="Data Source Table Filter"
                setOptions={{
                  fontSize: "14px"
                }}
                value={
                  job.config.job_config.bigquerySource.options.filter
                    ? job.config.job_config.bigquerySource.options.filter
                    : ""
                }
                onChange={value => {
                  setBigquerySourceOptions("filter", value);

                  const sanitizedValue = value.trim();
                  if (sanitizedValue === "") {
                    unsetBigquerySourceOptions("filter");
                  }
                }}
              />
            </EuiFormRow>
          </EuiFlexItem>
        </EuiFlexGroup>

        <EuiSpacer size="l" />

        <EuiAccordion
          id="sourceOptions"
          buttonContent={<EuiText size="s">Advanced configurations</EuiText>}
          paddingSize="m">
          <EuiFlexGroup>
            <EuiFlexItem>
              <EuiFormRow
                label="Google Cloud Parent Project ID"
                helpText="The Google Cloud Project ID of the table to bill for the export. Leave it blank to use the Service Account's default project.">
                <EuiFieldText
                  value={
                    job.config.job_config.bigquerySource.options.parentProject
                      ? job.config.job_config.bigquerySource.options
                          .parentProject
                      : ""
                  }
                  onChange={e =>
                    setBigquerySourceOptions("parentProject", e.target.value)
                  }
                  aria-label="Data Source Google Cloud Parent Project ID"
                />
              </EuiFormRow>
            </EuiFlexItem>

            <EuiFlexItem>
              <EuiFormRow
                label="Max Parallelism"
                helpText="The maximal number of partitions to split the data into.">
                <EuiFieldNumber
                  value={
                    job.config.job_config.bigquerySource.options.maxParallelism
                      ? job.config.job_config.bigquerySource.options
                          .maxParallelism
                      : ""
                  }
                  onChange={e => {
                    const sanitizedValue = parseInt(e.target.value);
                    !isNaN(sanitizedValue) &&
                      setBigquerySourceOptions(
                        "maxParallelism",
                        sanitizedValue
                      );
                  }}
                  aria-label="Data Source Max Parallelism"
                />
              </EuiFormRow>
            </EuiFlexItem>
          </EuiFlexGroup>

          <EuiSpacer size="l" />

          <EuiFlexGroup>
            <EuiFlexItem>
              <EuiFormRow
                label="Read Data Format"
                helpText="Data Format for reading from BigQuery.">
                <EuiSuperSelect
                  fullWidth
                  options={readDataFormatOptions}
                  valueOfSelected={
                    job.config.job_config.bigquerySource.options.readDataFormat
                  }
                  onChange={value => {
                    setBigquerySourceOptions("readDataFormat", value);
                  }}
                  hasDividers
                />
              </EuiFormRow>
            </EuiFlexItem>

            <EuiFlexItem>
              <EuiFormRow
                label="Optimized Empty Projection"
                helpText={
                  <p>
                    By enabling this, the connector uses an optimized empty
                    projection approach for <EuiCode>count()</EuiCode>{" "}
                    execution.
                  </p>
                }>
                <EuiSwitch
                  label="Use optimized empty projection?"
                  checked={
                    job.config.job_config.bigquerySource.options
                      .optimizedEmptyProjection === "true"
                  }
                  onChange={e =>
                    setBigquerySourceOptions(
                      "optimizedEmptyProjection",
                      e.target.checked
                    )
                  }
                />
              </EuiFormRow>
            </EuiFlexItem>
          </EuiFlexGroup>

          {job.config.job_config.bigquerySource.options.viewsEnabled ===
            "true" && (
            <Fragment>
              <EuiSpacer size="l" />

              <EuiFlexGroup>
                <EuiFlexItem>
                  <EuiFormRow
                    label="View Materalization Project ID"
                    helpText="The Google Cloud Project ID where the materialized view is going to be created. Defaults to view's Project ID.">
                    <EuiFieldText
                      value={
                        job.config.job_config.bigquerySource.options
                          .viewMaterializationProject
                          ? job.config.job_config.bigquerySource.options
                              .viewMaterializationProject
                          : ""
                      }
                      onChange={e =>
                        setBigquerySourceOptions(
                          "viewMaterializationProject",
                          e.target.value
                        )
                      }
                      aria-label="Data Source View Materalization Project ID"
                    />
                  </EuiFormRow>
                </EuiFlexItem>

                <EuiFlexItem>
                  <EuiFormRow
                    label="View Materalization Dataset"
                    helpText="The dataset where the materialized view is going to be created. Defaults to view's dataset.">
                    <EuiFieldText
                      value={
                        job.config.job_config.bigquerySource.options
                          .viewMaterializationDataset
                          ? job.config.job_config.bigquerySource.options
                              .viewMaterializationDataset
                          : ""
                      }
                      onChange={e =>
                        setBigquerySourceOptions(
                          "viewMaterializationDataset",
                          e.target.value
                        )
                      }
                      aria-label="Data Source View Materalization Dataset"
                    />
                  </EuiFormRow>
                </EuiFlexItem>
              </EuiFlexGroup>
            </Fragment>
          )}
        </EuiAccordion>
      </EuiForm>
    </Fragment>
  );
};
