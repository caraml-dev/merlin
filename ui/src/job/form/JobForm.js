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

import React, { useContext, useEffect, useState } from "react";
import { navigate } from "@reach/router";
import {
  EuiIcon,
  EuiPage,
  EuiPageBody,
  EuiPageSection,
  EuiPageHeader,
  EuiPageHeaderSection,
  EuiSteps,
  EuiTitle
} from "@elastic/eui";
import { addToast, replaceBreadcrumbs } from "@gojek/mlp-ui";
import { useMerlinApi } from "../../hooks/useMerlinApi";
import {
  validateBigqueryTable,
  validateBigqueryColumn
} from "../../validation/validateBigquery";
import { JobFormContext } from "./context";
import { JobFormOthers } from "./JobFormOthers";
import { JobFormSink } from "./JobFormSink";
import { JobFormSource } from "./JobFormSource";
import { JobFormStep } from "./JobFormStep";
import PropTypes from "prop-types";

const isSourceConfigured = job => {
  return (
    !!job.config.job_config.bigquerySource.table &&
    validateBigqueryTable(job.config.job_config.bigquerySource.table) &&
    job.config.job_config.bigquerySource.features.length > 0
  );
};

const isSinkConfigured = job => {
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

const isOthersConfigured = job => {
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

export const JobForm = ({ projectId, modelId, versionId, isNewJob = true }) => {
  const [model, setModel] = useState();
  const [{ data: models, isLoaded: modelsLoaded }] = useMerlinApi(
    `/projects/${projectId}/models`,
    {},
    [],
    !model
  );
  useEffect(() => {
    modelsLoaded && setModel(models.find(m => m.id.toString() === modelId));
  }, [models, modelsLoaded, modelId, setModel]);

  const [{ data: versions }] = useMerlinApi(
    `/models/${modelId}/versions`,
    {},
    [],
    true
  );

  const { job, setModel: setJobModel } = useContext(JobFormContext);

  useEffect(() => {
    const breadcrumbs = [];
    projectId &&
      breadcrumbs.push({
        text: "Models",
        href: `/merlin/projects/${projectId}/models`
      });
    projectId &&
      model &&
      breadcrumbs.push({
        text: model.name,
        href: `/merlin/projects/${projectId}/models/${model.id}`
      });
    isNewJob &&
      projectId &&
      model &&
      versionId &&
      breadcrumbs.push({
        text: `Model Version ${versionId}`,
        href: `/merlin/projects/${projectId}/models/${model.id}/versions/${versionId}`
      });
    !isNewJob &&
      breadcrumbs.push({
        text: "Jobs",
        href: `/merlin/projects/${projectId}/models/${modelId}/versions/${versionId}/jobs`
      });
    !isNewJob &&
      breadcrumbs.push({
        text: `${job.id}`,
        href: `/merlin/projects/${projectId}/models/${modelId}/versions/${versionId}/jobs/${job.id}`
      });
    breadcrumbs.push({
      text: isNewJob ? "Start Batch Job" : "Recreate Batch Job"
    });
    replaceBreadcrumbs(breadcrumbs);
  }, [projectId, model, versionId, isNewJob, modelId, job.id]);

  // Job related hooks
  useEffect(() => {
    model && setJobModel("type", model.type.toUpperCase());
  }, [model, setJobModel]);
  useEffect(() => {
    if (job.version_id && versions) {
      const version = versions.find(
        v => v.id.toString() === job.version_id.toString()
      );
      version && setJobModel("uri", version.artifact_uri + "/model");
    }
  }, [job.version_id, versions, setJobModel]);

  const [submissionResponse, submitForm] = useMerlinApi(
    `/models/${modelId}/versions/${job.version_id}/jobs`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" }
    },
    {},
    false
  );

  useEffect(() => {
    if (submissionResponse.isLoaded && !submissionResponse.error) {
      addToast({
        id: "create-job-success-toast",
        title: "Batch Job Created",
        color: "success",
        iconType: "check"
      });

      navigate(
        `/merlin/projects/${projectId}/models/${modelId}/versions/all/jobs`
      );
    }
  }, [submissionResponse, projectId, modelId, isNewJob]);

  const [currentStep, setCurrentStep] = useState(1);

  const onPrev = () => setCurrentStep(i => i - 1);
  const onNext = () => setCurrentStep(i => i + 1);
  const onSubmit = () => submitForm({ body: JSON.stringify(job) });

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
      isStepCompleted: isSourceConfigured(job)
    },
    {
      title: "Configure Data Sink",
      children: <JobFormSink />,
      isStepCompleted: isSinkConfigured(job)
    },
    {
      title: "Other Configurations",
      children: (
        <JobFormOthers
          versions={versions}
          isSelectVersionDisabled={!!versionId}
        />
      ),
      isStepCompleted: isOthersConfigured(job)
    }
  ];

  return (
    <EuiPage>
      <EuiPageBody>
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <EuiTitle size="l">
              <h1>
                <EuiIcon type="storage" size="xl" />
                &nbsp; {isNewJob ? "Start Batch Job" : "Recreate Batch Job"}
              </h1>
            </EuiTitle>
          </EuiPageHeaderSection>
        </EuiPageHeader>

        <EuiPageSection>
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
                  onSubmit={onSubmit}>
                  {step.children}
                </JobFormStep>
              );
              delete step["isStepCompleted"];
              return step;
            })}
          />
        </EuiPageSection>
      </EuiPageBody>
    </EuiPage>
  );
};

JobForm.propTypes = {
  projectId: PropTypes.string,
  modelId: PropTypes.string,
  versionId: PropTypes.string,
  isNewJob: PropTypes.bool
};
