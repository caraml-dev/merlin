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
import EndpointsTableModelVersion from "../../pages/version/components/modal/EndpointsTableModelVersion";

const DeleteModelVersionModal = ({
  version,
  model,
  callback,
  closeModal
}) => {
  const [activeEndpoint, setActiveEndpoint] = useState([])
  const [inactiveEndpoint, setInactiveEndpoint] = useState([])
  const [deleteConfirmation, setDeleteConfirmation] = useState('')

  const [{ isLoading, isLoaded }, undeployVersion] = useMerlinApi(
    `/models/${model.id}/versions/${version.id}`,
    { method: "DELETE", addToast: true, mock: mocks.noBody },
    {},
    false
  );

  useEffect(() => {
    if (isLoaded) {
      setDeleteConfirmation("")
      closeModal();
      callback();
    }
  }, [isLoaded, closeModal, callback]);

  useEffect(() => {
    setActiveEndpoint(version.endpoints.filter(item => item.status === "pending" || item.status ==="running" || item.status ==="serving"))
    setInactiveEndpoint(version.endpoints.filter(item => item.status !== "terminated"))
  }, [version])

  return (
    <EuiOverlayMask>
      <EuiConfirmModal
        title="Delete Model Version"
        onCancel={closeModal}
        onConfirm={undeployVersion}
        cancelButtonText="Cancel"
        confirmButtonText="Delete"
        buttonColor="danger"
        confirmButtonDisabled={deleteConfirmation !== `${model.name}-version-${version.id}` || activeEndpoint.length > 0}>
        {isLoading && (
          <EuiProgress
            size="xs"
            color="accent"
            className="euiProgress-beforePre"
          />
        )}
        {activeEndpoint.length > 0 ? (
          <span>
            You cannot delete this Model Version because there are <b> {activeEndpoint.length} Endpoints</b> using this version. 
            <br/> <br/> If you still wish to delete this model version, please <b>Undeploy</b> Endpoints that use this version. <br/>
          </span>
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
        {activeEndpoint.length === 0 && inactiveEndpoint.length > 0 && (
            <p>Deleting this Model Version will also delete {inactiveEndpoint.length} <b>Failed</b> Endpoints using this version. </p>
        )}

        <EndpointsTableModelVersion
          endpoints={activeEndpoint.length > 0 ? activeEndpoint : inactiveEndpoint}
        />
        
      </EuiConfirmModal>
    </EuiOverlayMask>
  );
};

export default DeleteModelVersionModal;
