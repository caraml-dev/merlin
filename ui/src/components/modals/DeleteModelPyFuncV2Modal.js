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

import React, { useEffect, useState } from "react";
import { EuiConfirmModal, EuiFieldText, EuiOverlayMask, EuiProgress } from "@elastic/eui";
import { useMerlinApi } from "../../hooks/useMerlinApi";
import mocks from "../../mocks";

const DeleteModelPyFuncV2Modal = ({
  model,
  callback,
  closeModal
}) => {
  const [activeJob, setActiveJob] = useState([])
  const [inactiveJob, setInactiveJob] = useState([])
  const [deleteConfirmation, setDeleteConfirmation] = useState('')

  const [{ isLoading, isLoaded }, undeployVersion] = useMerlinApi(
    `projects/${model.project_id}/models/${model.id}`,
    { method: "DELETE", addToast: true, mock: mocks.noBody },
    {},
    false
  );

  const [jobs] = useMerlinApi(
    `/projects/${model.project_id}/jobs?model_id=${model.id}`,
    { mock: mocks.jobList },
    []
  );

  useEffect(() => {
    setActiveJob(jobs.data.filter(item => item.status === "pending" || item.status === "running"))
    setInactiveJob(jobs.data.filter(item => item.status !== "pending" && item.status !== "running"))
  }, [jobs])

  useEffect(() => {
    if (isLoaded) {
      closeModal();
      callback();
    }
  }, [isLoaded, closeModal, callback]);

  return (
    <EuiOverlayMask>
      <EuiConfirmModal
        title="Delete Model"
        onCancel={closeModal}
        onConfirm={undeployVersion}
        cancelButtonText="Cancel"
        confirmButtonText="Delete"
        buttonColor="danger"
        confirmButtonDisabled={deleteConfirmation !== model.name || activeJob.length > 0}>
        {isLoading && (
          <EuiProgress
            size="xs"
            color="accent"
            className="euiProgress-beforePre"
          />
        )}
        {jobs.isLoading ? (
          <EuiProgress
          size="xs"
          color="accent"
          className="euiProgress-beforePre"
          />
        ) : (
          <div>
            {activeJob.length > 0 ? (
              <div>
                You cannot delete this Model because there are <b> {activeJob.length} Active Prediction Jobs</b> using this model. 
                <br/> <br/> If you still wish to delete this model, please <b>Terminate</b> Prediction Jobs that use this model or wait until the job completes.         
              </div>
              
            ) : (
              <p>
                You are about to delete model <b>{model.name}</b>. This action cannot be undone. 
                
                <br/> <br/> To confirm, please type "<b>{model.name}</b>" in the box below
                  <EuiFieldText     
                    fullWidth            
                    placeholder={model.name}
                    value={deleteConfirmation}
                    onChange={(e) => setDeleteConfirmation(e.target.value)}
                    isInvalid={deleteConfirmation !== model.name} />  
              </p>
            )}
            {activeJob.length === 0 && inactiveJob.length > 0 && (
                <span>Deleting this Model will also delete {inactiveJob.length} <b>Inactive</b> Prediction Jobs using this model. <br/> <br/></span>
            )}
            {/* <JobsTableModelVersion jobs={activeJob.length > 0 ? activeJob : inactiveJob} isLoaded={jobs.isLoaded} error={jobs.error} /> */}
          </div>
        )}
        
      </EuiConfirmModal>
    </EuiOverlayMask>
  );
};

export default DeleteModelPyFuncV2Modal;
