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

const DeleteModelModal = ({
  model,
  callback,
  closeModal
}) => {
  const [activeVersionEndpoints, setActiveVersionEndpoints] = useState([])
  const [activeModelEndpoints, setActiveModelEndpoints] = useState([])

  const [deleteConfirmation, setDeleteConfirmation] = useState('')

  const [{ isLoading, isLoaded }, deleteModel] = useMerlinApi(
    `projects/${model.project_id}/models/${model.id}`,
    { method: "DELETE", addToast: true, mock: mocks.noBody },
    {},
    false
  );

  const[versions] = useMerlinApi(
    `/models/${model.id}/versions`,
    {}
  )

  useEffect(() => {
    if (isLoaded) {
      setDeleteConfirmation("")
      closeModal();
      callback();
    }
  }, [isLoaded, closeModal, callback]);

  useEffect(() => {
    const tempActiveVersionEndpoint = [];
    if (versions.isLoaded){
      for (const item of versions.data){
        const activeEndpoint = item.endpoints.filter(
          (endpoint) => ["pending", "running", "serving"].includes(endpoint.status)
        );
        if (activeEndpoint.length > 0){
          tempActiveVersionEndpoint.push(activeEndpoint)
        } 
      }
    }
    setActiveVersionEndpoints(tempActiveVersionEndpoint)
  }, [versions])

  useEffect(() => {
    setActiveModelEndpoints(model.endpoints.filter(endpoint => endpoint.status !== "terminated"));
  }, [model])

  return (
    <EuiOverlayMask>
      <EuiConfirmModal
        title="Delete Model"
        onCancel={closeModal}
        onConfirm={deleteModel}
        cancelButtonText="Cancel"
        confirmButtonText="Delete"
        buttonColor="danger"
        confirmButtonDisabled={deleteConfirmation !== model.name || activeVersionEndpoints.length > 0}>

        {versions.isLoading ? (
            <EuiProgress
            size="xs"
            color="accent"
            className="euiProgress-beforePre"
          />
        ) : (
        <div>
          {activeModelEndpoints.length > 0 ? (
            <span>
              You cannot delete this Model because there are <b> {activeModelEndpoints.length} Endpoints</b> using this model. 
              <br/> <br/> If you still wish to delete this model, please <b>Stop Serving</b> Endpoints that use this model. <br/>
            </span>        
          ) : (
            <div>
            {activeVersionEndpoints.length > 0 ? (
              <span>
                You cannot delete this Model because there are <b> {activeVersionEndpoints.length} Endpoints</b> currently using the model version associated with it. 
                <br/> <br/> If you still wish to delete this model, please <b>Undeploy</b> Endpoints that use this version. <br/>
              </span>
            ) : (
              <div>
                  You are about to delete model <b>{model.name}</b>. This action cannot be undone. 
                  <br/> <br/> Please note that all the related entities of this model (<b>Model Version</b> and <b>Model Version Endpoint)</b> will be <b>deleted</b>.
                  <br/> <br/> To confirm, please type "<b>{model.name}</b>" in the box below
                  <EuiFieldText     
                    fullWidth            
                    placeholder={model.name}
                    value={deleteConfirmation}
                    onChange={(e) => setDeleteConfirmation(e.target.value)}
                    isInvalid={deleteConfirmation !== model.name} />  
              </div>
            )}
            </div>
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

export default DeleteModelModal;
