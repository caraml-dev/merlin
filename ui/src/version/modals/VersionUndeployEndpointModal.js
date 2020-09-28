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

import React, { useEffect } from "react";
import { EuiConfirmModal, EuiOverlayMask, EuiProgress } from "@elastic/eui";
import mocks from "../../mocks";
import { useMerlinApi } from "../../hooks/useMerlinApi";

const VersionUndeployEndpointModal = ({
  versionEndpoint,
  version,
  model,
  updateVersionsCallback,
  closeModal
}) => {
  const [{ isLoading, isLoaded }, undeployVersion] = useMerlinApi(
    `/models/${model.id}/versions/${version.id}/endpoint/${versionEndpoint.id}`,
    { method: "DELETE", addToast: true, mock: mocks.noBody },
    {},
    false
  );

  useEffect(() => {
    if (isLoaded) {
      closeModal();
      updateVersionsCallback();
    }
  }, [isLoaded, closeModal, updateVersionsCallback]);

  return (
    <EuiOverlayMask>
      <EuiConfirmModal
        title="Undeploy model version"
        onCancel={closeModal}
        onConfirm={undeployVersion}
        cancelButtonText="Cancel"
        confirmButtonText="Undeploy"
        buttonColor="danger">
        <p>
          You're about to undeploy the endpoint for model version{" "}
          <b>{version.id}</b> from <b>{versionEndpoint.environment_name}</b>{" "}
          environment.
        </p>
        <p>The following endpoint will not be accessible anymore.</p>
        {isLoading && (
          <EuiProgress
            size="xs"
            color="accent"
            className="euiProgress-beforePre"
          />
        )}
        <pre>
          <code>{versionEndpoint.url}</code>
        </pre>
      </EuiConfirmModal>
    </EuiOverlayMask>
  );
};

export default VersionUndeployEndpointModal;
