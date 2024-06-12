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
  EuiConfirmModal,
  EuiFieldText,
  EuiOverlayMask,
  EuiProgress,
} from "@elastic/eui";
import PropTypes from "prop-types";
import React, { useEffect, useState } from "react";
import { useMerlinApi } from "../../../hooks/useMerlinApi";
import mocks from "../../../mocks";

const StopPredictionJobModal = ({ job, closeModal }) => {
  const [{ isLoading, isLoaded }, stopRunningJob] = useMerlinApi(
    `/models/${job.model_id}/versions/${job.version_id}/jobs/${job.id}/stop`,
    { method: "PUT", addToast: true, mock: mocks.noBody },
    {},
    false,
  );

  const [deleteConfirmation, setDeleteConfirmation] = useState("");

  useEffect(() => {
    if (isLoaded) {
      closeModal();
    }
  }, [isLoaded, closeModal]);

  return (
    <EuiOverlayMask>
      <EuiConfirmModal
        title={
          job.status === "pending" || job.status === "running"
            ? "Stop Job"
            : "Delete Job"
        }
        onCancel={closeModal}
        onConfirm={stopRunningJob}
        cancelButtonText="Cancel"
        confirmButtonText={
          job.status === "pending" || job.status === "running"
            ? "Stop Job"
            : "Delete Job"
        }
        buttonColor="danger"
        confirmButtonDisabled={
          deleteConfirmation !==
          `id-${job.id}-model-${job.model_id}-version-${job.version_id}`
        }
      >
        <div>
          You're about to{" "}
          {job.status === "pending" || job.status === "running"
            ? "stop"
            : "delete"}{" "}
          prediction job id <b>{job.id}</b> in version <b>{job.version_id}</b>{" "}
          of model <b>{job.model_id}</b>.
          <br /> <br /> To confirm, please type "
          <b>
            id-{job.id}-model-{job.model_id}-version-{job.version_id}
          </b>
          " in the box below
          <EuiFieldText
            fullWidth
            placeholder={`id-${job.id}-model-${job.model_id}-version-${job.version_id}`}
            value={deleteConfirmation}
            onChange={(e) => setDeleteConfirmation(e.target.value)}
            isInvalid={
              deleteConfirmation !==
              `id-${job.id}-model-${job.model_id}-version-${job.version_id}`
            }
          />
        </div>
        {isLoading && (
          <EuiProgress
            size="xs"
            color="accent"
            className="euiProgress-beforePre"
          />
        )}
      </EuiConfirmModal>
    </EuiOverlayMask>
  );
};

StopPredictionJobModal.propTypes = {
  job: PropTypes.object,
  closeModal: PropTypes.func,
};

export default StopPredictionJobModal;
