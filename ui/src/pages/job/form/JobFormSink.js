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
  EuiText,
} from "@elastic/eui";
import React, { Fragment, useContext, useState } from "react";
import {
  validateBigqueryColumn,
  validateBigqueryTable,
} from "../../../validation/validateBigquery";
import { ClusteredFieldsComboBox } from "./components/ClusteredFieldsComboBox";
import { ResultDataTypeSelect } from "./components/ResultDataTypeSelect";
import { JobFormContext } from "./context";

export const JobFormSink = () => {
  const { job, setBigquerySink, setBigquerySinkOptions, setModelResult } =
    useContext(JobFormContext);

  const [isValidTable, setValidTable] = useState(
    job.config.job_config.bigquerySink.table
      ? validateBigqueryTable(job.config.job_config.bigquerySink.table)
      : true
  );
  const [isValidColumn, setValidColumn] = useState(
    job.config.job_config.bigquerySink.resultColumn
      ? validateBigqueryColumn(job.config.job_config.bigquerySink.resultColumn)
      : true
  );

  const sinkTypeOptions = [
    {
      value: "bigquery",
      inputDisplay: "Google BigQuery",
      dropdownDisplay: <strong>Google BigQuery</strong>,
    },
  ];

  const saveModeOptions = [
    {
      value: "ERRORIFEXISTS",
      inputDisplay: "ERRORIFEXISTS",
      dropdownDisplay: (
        <Fragment>
          <strong>ERRORIFEXISTS</strong>
          <EuiText size="xs">
            <p className="euiTextColor--subdued">
              Throw error if the destination table already exists.
            </p>
          </EuiText>
        </Fragment>
      ),
    },
    {
      value: "OVERWRITE",
      inputDisplay: "OVERWRITE",
      dropdownDisplay: (
        <Fragment>
          <strong>OVERWRITE</strong>
          <EuiText size="xs">
            <p className="euiTextColor--subdued">
              Overwrite the destination table if it exists.
            </p>
          </EuiText>
        </Fragment>
      ),
    },
    {
      value: "APPEND",
      inputDisplay: "APPEND",
      dropdownDisplay: (
        <Fragment>
          <strong>APPEND</strong>
          <EuiText size="xs">
            <p className="euiTextColor--subdued">
              Append the new result into destination table if it exists.
            </p>
          </EuiText>
        </Fragment>
      ),
    },
    {
      value: "IGNORE",
      inputDisplay: "IGNORE",
      dropdownDisplay: (
        <Fragment>
          <strong>IGNORE</strong>
          <EuiText size="xs">
            <p className="euiTextColor--subdued">
              Do not write the prediction result if the destination table
              exists.
            </p>
          </EuiText>
        </Fragment>
      ),
    },
  ];

  const createTableOptions = [
    {
      value: "CREATE_IF_NEEDED",
      inputDisplay: "CREATE_IF_NEEDED",
      dropdownDisplay: (
        <Fragment>
          <strong>CREATE_IF_NEEDED</strong>
          <EuiText size="xs">
            <p className="euiTextColor--subdued">
              Configures the job to create the table if it does not exist.
            </p>
          </EuiText>
        </Fragment>
      ),
    },
    {
      value: "CREATE_NEVER",
      inputDisplay: "CREATE_NEVER",
      dropdownDisplay: (
        <Fragment>
          <strong>CREATE_NEVER</strong>
          <EuiText size="xs">
            <p className="euiTextColor--subdued">
              Configures the job to fail if the table does not exist.
            </p>
          </EuiText>
        </Fragment>
      ),
    },
  ];

  const intermediateFormatOptions = [
    {
      value: "parquet",
      inputDisplay: "Parquet",
      dropdownDisplay: (
        <Fragment>
          <strong>Parquet</strong>
        </Fragment>
      ),
    },
    {
      value: "orc",
      inputDisplay: "ORC",
      dropdownDisplay: (
        <Fragment>
          <strong>ORC</strong>
        </Fragment>
      ),
    },
  ];

  return (
    <Fragment>
      <EuiForm>
        <EuiFlexGroup>
          <EuiFlexItem>
            <EuiFormRow
              fullWidth
              label="Select Data Sink Type *"
              helpText="Currently only support Google BigQuery."
            >
              <EuiSuperSelect
                fullWidth
                options={sinkTypeOptions}
                valueOfSelected="bigquery"
                onChange={() => {}}
                hasDividers
              />
            </EuiFormRow>
          </EuiFlexItem>

          <EuiFlexItem>
            <EuiFormRow
              fullWidth
              label="Data Sink Table *"
              helpText="Specify the location of BigQuery sink table."
            >
              <EuiFieldText
                fullWidth
                placeholder="Format: project_name.dataset_name.table_name"
                value={job.config.job_config.bigquerySink.table}
                onChange={(e) => {
                  const val = e.target.value;
                  setValidTable(validateBigqueryTable(val));
                  setBigquerySink("table", val);
                }}
                isInvalid={!isValidTable}
                aria-label="Data Sink Table"
              />
            </EuiFormRow>
          </EuiFlexItem>
        </EuiFlexGroup>

        <EuiSpacer size="l" />

        <EuiFlexGroup>
          <EuiFlexItem>
            <ResultDataTypeSelect
              resultType={job.config.job_config.model.result.type}
              resultItemType={job.config.job_config.model.result.item_type}
              setResultType={(selected) => setModelResult("type", selected)}
              setResultItemType={(selected) =>
                setModelResult("item_type", selected)
              }
            />
          </EuiFlexItem>

          <EuiFlexItem>
            <EuiFormRow
              fullWidth
              label="Result Column *"
              helpText="Column name in the destination table that will be used to store the prediction result."
            >
              <EuiFieldText
                fullWidth
                value={job.config.job_config.bigquerySink.resultColumn}
                onChange={(e) => {
                  const val = e.target.value;
                  setValidColumn(validateBigqueryColumn(val));
                  setBigquerySink("resultColumn", val);
                }}
                isInvalid={!isValidColumn}
                placeholder="e.g., prediction_score"
                aria-label="Data Sink Result Column"
              />
            </EuiFormRow>
          </EuiFlexItem>
        </EuiFlexGroup>

        <EuiSpacer size="l" />

        <EuiFlexGroup>
          <EuiFlexItem>
            <EuiFormRow
              fullWidth
              label="Select Save Mode *"
              helpText="Specify the expected behavior of saving the prediction result into destination table."
            >
              <EuiSuperSelect
                fullWidth
                options={saveModeOptions}
                valueOfSelected={job.config.job_config.bigquerySink.saveMode}
                onChange={(value) => {
                  setBigquerySink("saveMode", value);
                }}
                hasDividers
              />
            </EuiFormRow>
          </EuiFlexItem>

          <EuiFlexItem>
            <EuiFormRow
              fullWidth
              label="GCS Staging Bucket *"
              helpText="Google Cloud Storage bucket that will be used as temporary storage for storing prediction result before loading it to the destination table."
            >
              <EuiFieldText
                fullWidth
                value={job.config.job_config.bigquerySink.stagingBucket}
                onChange={(e) =>
                  setBigquerySink("stagingBucket", e.target.value)
                }
                placeholder="GCS bucket name"
                aria-label="Data Sink GCS Staging Bucket"
              />
            </EuiFormRow>
          </EuiFlexItem>
        </EuiFlexGroup>

        <EuiSpacer size="l" />

        <EuiAccordion
          id="sinkOptions"
          buttonContent={<EuiText size="s">Advanced configurations</EuiText>}
          paddingSize="m"
        >
          <EuiFlexGroup>
            <EuiFlexItem>
              <EuiFormRow
                label="Create Table"
                helpText="Specifies whether the job is allowed to create new tables."
              >
                <EuiSuperSelect
                  fullWidth
                  options={createTableOptions}
                  valueOfSelected={
                    job.config.job_config.bigquerySink.options.createDisposition
                  }
                  onChange={(value) => {
                    setBigquerySinkOptions("createDisposition", value);
                  }}
                  hasDividers
                />
              </EuiFormRow>
            </EuiFlexItem>

            <EuiFlexItem>
              <EuiFormRow
                label="Intermediate Format"
                helpText="The format of the data before it is loaded to BigQuery."
              >
                <EuiSuperSelect
                  fullWidth
                  options={intermediateFormatOptions}
                  valueOfSelected={
                    job.config.job_config.bigquerySink.options
                      .intermediateFormat
                  }
                  onChange={(value) => {
                    setBigquerySinkOptions("intermediateFormat", value);
                  }}
                  hasDividers
                />
              </EuiFormRow>
            </EuiFlexItem>
          </EuiFlexGroup>

          <EuiSpacer size="l" />

          <EuiFlexGroup>
            <EuiFlexItem>
              <EuiFormRow
                label="Partition Field"
                helpText="Partition the table by this field."
              >
                <EuiFieldText
                  value={
                    job.config.job_config.bigquerySink.options.partitionField
                      ? job.config.job_config.bigquerySink.options
                          .partitionField
                      : ""
                  }
                  onChange={(e) =>
                    setBigquerySinkOptions("partitionField", e.target.value)
                  }
                  aria-label="Data Sink Partition Field"
                />
              </EuiFormRow>
            </EuiFlexItem>

            <EuiFlexItem>
              <EuiFormRow
                label="Partition Expiration"
                helpText="Number of milliseconds for which to keep the storage for partitions in the table."
              >
                <EuiFieldNumber
                  value={
                    job.config.job_config.bigquerySink.options
                      .partitionExpirationMs
                      ? job.config.job_config.bigquerySink.options
                          .partitionExpirationMs
                      : ""
                  }
                  onChange={(e) => {
                    const sanitizedValue = parseInt(e.target.value);
                    !isNaN(sanitizedValue) &&
                      setBigquerySinkOptions(
                        "partitionExpirationMs",
                        sanitizedValue
                      );
                  }}
                  aria-label="Data Sink Partition Expiration"
                />
              </EuiFormRow>
            </EuiFlexItem>
          </EuiFlexGroup>

          <EuiSpacer size="l" />

          <EuiFlexGroup>
            <EuiFlexItem>
              <EuiFormRow
                label="Allow Field Addition"
                helpText={
                  <p>
                    Adds the <EuiCode>ALLOW_FIELD_ADDITION</EuiCode>{" "}
                    SchemaUpdateOption to the BigQuery LoadJob.
                  </p>
                }
              >
                <EuiSwitch
                  label="Add ALLOW_FIELD_ADDITION?"
                  checked={
                    job.config.job_config.bigquerySink.options
                      .allowFieldAddition === "true"
                  }
                  onChange={(e) =>
                    setBigquerySinkOptions(
                      "allowFieldAddition",
                      e.target.checked
                    )
                  }
                />
              </EuiFormRow>
            </EuiFlexItem>

            <EuiFlexItem>
              <EuiFormRow
                label="Allow Field Relaxation"
                helpText={
                  <p>
                    Adds the <EuiCode>ALLOW_FIELD_RELAXATION</EuiCode>{" "}
                    SchemaUpdateOption to the BigQuery LoadJob.
                  </p>
                }
              >
                <EuiSwitch
                  label="Add ALLOW_FIELD_RELAXATION?"
                  checked={
                    job.config.job_config.bigquerySink.options
                      .allowFieldRelaxation === "true"
                  }
                  onChange={(e) =>
                    setBigquerySinkOptions(
                      "allowFieldRelaxation",
                      e.target.checked
                    )
                  }
                />
              </EuiFormRow>
            </EuiFlexItem>
          </EuiFlexGroup>

          <EuiSpacer size="l" />

          <ClusteredFieldsComboBox
            options={
              job.config.job_config.bigquerySink.options.clusteredFields
                ? job.config.job_config.bigquerySink.options.clusteredFields
                    .split(",")
                    .map((field) => {
                      return { label: field };
                    })
                : []
            }
            onChange={(selected) => {
              setBigquerySinkOptions("clusteredFields", selected);
            }}
          />
        </EuiAccordion>
      </EuiForm>
    </Fragment>
  );
};
