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
  const [activeJobs, setActiveJobs] = useState([])
  const [inactiveJobs, setInactiveJobs] = useState([])
  const [deleteConfirmation, setDeleteConfirmation] = useState('')

  const [{ isLoading, isLoaded }, deleteModel] = useMerlinApi(
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

  const isActiveJob = function(status) {
    return ["pending", "running"].includes(status)
  }

  useEffect(() => {
    setActiveJobs(jobs.data.filter(item => isActiveJob(item.status)))
    setInactiveJobs(jobs.data.filter(item => !isActiveJob(item.status)))
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
        onConfirm={deleteModel}
        cancelButtonText="Cancel"
        confirmButtonText="Delete"
        buttonColor="danger"
        confirmButtonDisabled={deleteConfirmation !== model.name || activeJobs.length > 0}>
        {jobs.isLoading ? (
          <EuiProgress
          size="xs"
          color="accent"
          className="euiProgress-beforePre"
          />
        ) : (
          <div>
            {activeJobs.length > 0 ? (
              <div>
                You cannot delete this Model because there are <b> {activeJobs.length} Active Prediction Jobs</b> using this model. 
                <br/> <br/> If you still wish to delete this model, please <b>Terminate</b> Prediction Jobs that use this model or wait until the job completes.         
              </div>
              
            ) : (
              <div>
                You are about to delete model <b>{model.name}</b>. This action cannot be undone. 
                
                <br/> <br/> To confirm, please type "<b>{model.name}</b>" in the box below
                  <EuiFieldText     
                    fullWidth            
                    placeholder={model.name}
                    value={deleteConfirmation}
                    onChange={(e) => setDeleteConfirmation(e.target.value)}
                    isInvalid={deleteConfirmation !== model.name} />  
              </div>
            )}
            {activeJobs.length === 0 && inactiveJobs.length > 0 && (
                <span>Deleting this Model will also delete {inactiveJobs.length} <b>Inactive</b> Prediction Jobs using this model. <br/> <br/></span>
            )}
          </div>
        )}
        {isLoading && (
          <span>
            <EuiProgress
              size="xs"
              color="accent"
              className="euiProgress-beforePre"
            />
          </span>
        )}
      </EuiConfirmModal>
    </EuiOverlayMask>
  );
};

export default DeleteModelPyFuncV2Modal;
