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
import JobsTableModelVersion from "../../pages/version/components/modal/JobsTableModelVersion";

const DeleteModelVersionPyFuncV2Modal = ({
  version,
  model,
  callback,
  closeModal
}) => {
  const [activeJob, setActiveJob] = useState([])
  const [inactiveJob, setInactiveJob] = useState([])
  const [deleteConfirmation, setDeleteConfirmation] = useState('')

  const [{ isLoading, isLoaded }, undeployVersion] = useMerlinApi(
    `/models/${model.id}/versions/${version.id}`,
    { method: "DELETE", addToast: true, mock: mocks.noBody },
    {},
    false
  );

  const [jobs] = useMerlinApi(
    `/models/${model.id}/versions/${version.id}/jobs`,
    { mock: mocks.jobList },
    []
  );

  const isActiveJob = function(status) {
    return ["pending", "running"].includes(status)
  }

  useEffect(() => {
    setActiveJob(jobs.data.filter(item => isActiveJob(item.status)))
    setInactiveJob(jobs.data.filter(item => !isActiveJob(item.status)))
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
        title="Delete Model Version"
        onCancel={closeModal}
        onConfirm={undeployVersion}
        cancelButtonText="Cancel"
        confirmButtonText="Delete"
        buttonColor="danger"
        confirmButtonDisabled={deleteConfirmation !== `${model.name}-version-${version.id}` || activeJob.length > 0}>
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
                You cannot delete this Model Version because there are <b> {activeJob.length} Active Prediction Jobs</b> using this version. 
                <br/> <br/> If you still wish to delete this model version, please <b>Terminate</b> Prediction Jobs that use this version or wait until the job completes.         
              </div>
              
            ) : (
              <p>
                You are about to delete model <b>{model.name}</b> version <b>{version.id}</b>. This action cannot be undone. 
                
                <br/> <br/> To confirm, please type "<b>{model.name}-version-{version.id}</b>" in the box below
                  <EuiFieldText     
                    fullWidth            
                    placeholder={`${model.name}-version-${version.id}`}
                    value={deleteConfirmation}
                    onChange={(e) => setDeleteConfirmation(e.target.value)}
                    isInvalid={deleteConfirmation !== `${model.name}-version-${version.id}`} />  
              </p>
            )}
            {activeJob.length === 0 && inactiveJob.length > 0 && (
                <span>Deleting this Model Version will also delete {inactiveJob.length} <b>Inactive</b> Prediction Jobs using this version. <br/> <br/></span>
            )}
            <JobsTableModelVersion jobs={activeJob.length > 0 ? activeJob : inactiveJob} isLoaded={jobs.isLoaded} error={jobs.error} />
          </div>
        )}
        
      </EuiConfirmModal>
    </EuiOverlayMask>
  );
};

export default DeleteModelVersionPyFuncV2Modal;
