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

import { addToast, replaceBreadcrumbs } from "@caraml-dev/ui-lib";
import { EuiPageTemplate, EuiPanel, EuiSpacer, EuiSteps } from "@elastic/eui";
import PropTypes from "prop-types";
import React, { Fragment, useContext, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useMerlinApi } from "../../../hooks/useMerlinApi";
import {
  validateBigqueryColumn,
  validateBigqueryTable,
} from "../../../validation/validateBigquery";
import { generateBreadcrumbs } from "../utils/breadcrumbs";
import { JobFormOthers } from "./JobFormOthers";
import { JobFormSink } from "./JobFormSink";
import { JobFormSource } from "./JobFormSource";
import { JobFormStep } from "./JobFormStep";
import { JobFormContext } from "./context";

const _ = require("lodash");
const isSourceConfigured = (job) => {
  return (
    !!job.config.job_config.bigquerySource.table &&
    validateBigqueryTable(job.config.job_config.bigquerySource.table) &&
    job.config.job_config.bigquerySource.features.length > 0
  );
};

const isSinkConfigured = (job) => {
  return (
    !!job.config.job_config.model.result.type &&
    !!job.config.job_config.model.result.item_type &&
    !!job.config.job_config.bigquerySink.table &&
    validateBigqueryTable(job.config.job_config.bigquerySink.table) &&
    !!job.config.job_config.bigquerySink.stagingBucket &&
    !!job.config.job_config.bigquerySink.resultColumn &&
    validateBigqueryColumn(job.config.job_config.bigquerySink.resultColumn) &&
    !!job.config.job_config.bigquerySink.saveMode
  );
};

const isOthersConfigured = (job) => {
  return (
    !!job.version_id &&
    !!job.config.service_account_name &&
    !!job.config.resource_request.driver_cpu_request &&
    !!job.config.resource_request.driver_memory_request &&
    !!job.config.resource_request.executor_cpu_request &&
    !!job.config.resource_request.executor_memory_request &&
    !!job.config.resource_request.executor_replica
  );
};

export const JobForm = ({ isNewJob = true }) => {
  const { projectId, modelId, versionId } = useParams();
  const navigate = useNavigate();
  const [model, setModel] = useState();
  const [{ data: models, isLoaded: modelsLoaded }] = useMerlinApi(
    `/projects/${projectId}/models`,
    {},
    [],
    !model,
  );
  useEffect(() => {
    modelsLoaded && setModel(models.find((m) => m.id.toString() === modelId));
  }, [models, modelsLoaded, modelId, setModel]);

  const [{ data: versions }] = useMerlinApi(
    `/models/${modelId}/versions`,
    {},
    [],
    true,
  );

  const { job, setModel: setJobModel } = useContext(JobFormContext);

  useEffect(() => {
    replaceBreadcrumbs([
      ...generateBreadcrumbs(projectId, modelId, model, versionId, job.id),
      {
        text: isNewJob ? "Start Batch Job" : "Recreate Batch Job",
      },
    ]);
  }, [projectId, model, versionId, modelId, job.id, isNewJob]);

  // Job related hooks
  useEffect(() => {
    model && setJobModel("type", model.type.toUpperCase());
  }, [model, setJobModel]);

  useEffect(() => {
    if (job.version_id && versions) {
      const version = versions.find(
        (v) => v.id.toString() === job.version_id.toString(),
      );
      version && setJobModel("uri", version.artifact_uri + "/model");
    }
  }, [job.version_id, versions, setJobModel]);

  const [submissionResponse, submitForm] = useMerlinApi(
    `/models/${modelId}/versions/${job.version_id}/jobs`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false,
  );

  useEffect(() => {
    if (submissionResponse.isLoaded && !submissionResponse.error) {
      addToast({
        id: "create-job-success-toast",
        title: "Batch Job Created",
        color: "success",
        iconType: "check",
      });

      navigate(
        `/merlin/projects/${projectId}/models/${modelId}/versions/all/jobs`,
      );
    }
  }, [submissionResponse, projectId, modelId, isNewJob, navigate]);

  const [currentStep, setCurrentStep] = useState(1);

  const onPrev = () => setCurrentStep((i) => i - 1);
  const onNext = () => setCurrentStep((i) => i + 1);
  const onSubmit = () => {
    if (job?.config?.image_builder_resource_request?.cpu_request === "") {
      delete job.config.image_builder_resource_request.cpu_request;
    }
    if (job?.config?.image_builder_resource_request?.memory_request === "") {
      delete job.config.image_builder_resource_request.memory_request;
    }
    if (_.isEmpty(job?.config?.image_builder_resource_request)) {
      delete job.config.image_builder_resource_request;
    }
    submitForm({ body: JSON.stringify(job) });
  };

  const stepStatus = (step, idx) => {
    switch (true) {
      case idx > currentStep:
        return "disabled";
      case step.isStepCompleted:
        return "complete";
      default:
        return null;
    }
  };

  const steps = [
    {
      title: "Configure Data Source",
      children: <JobFormSource />,
      isStepCompleted: isSourceConfigured(job),
    },
    {
      title: "Configure Data Sink",
      children: <JobFormSink />,
      isStepCompleted: isSinkConfigured(job),
    },
    {
      title: "Other Configurations",
      children: (
        <JobFormOthers
          versions={versions}
          isSelectVersionDisabled={!!versionId}
        />
      ),
      isStepCompleted: isOthersConfigured(job),
    },
  ];

  return (
    <EuiPageTemplate restrictWidth="90%" paddingSize="none">
      <EuiPageTemplate.Header
        bottomBorder={false}
        iconType={"storage"}
        pageTitle={
          <Fragment>
            &nbsp; {isNewJob ? "Start Batch Job" : "Recreate Batch Job"}
          </Fragment>
        }
      />

      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
        <EuiPanel>
          <EuiSteps
            headingElement="div"
            steps={steps.map((step, index) => {
              step.status = stepStatus(step, index + 1);
              step.children = (
                <JobFormStep
                  step={index + 1}
                  currentStep={currentStep}
                  isStepCompleted={step.isStepCompleted}
                  isLastStep={index === steps.length - 1}
                  onPrev={onPrev}
                  onNext={onNext}
                  onSubmit={onSubmit}
                >
                  {step.children}
                </JobFormStep>
              );
              delete step["isStepCompleted"];
              return step;
            })}
          />
        </EuiPanel>
      </EuiPageTemplate.Section>
      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};

JobForm.propTypes = {
  isNewJob: PropTypes.bool,
};
