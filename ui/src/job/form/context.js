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

import React, { useCallback, useEffect, useState } from "react";
import PropTypes from "prop-types";

export const JobFormContext = React.createContext({});

export const JobFormContextProvider = ({ job: initJob, ...props }) => {
  const [job, setJob] = useState(initJob);

  useEffect(() => {
    console.log("Debug Job:", job);
  }, [job]);

  const setVersionId = useCallback(
    versionId => {
      setJob(j => ({
        ...j,
        version_id: versionId
      }));
    },
    [setJob]
  );

  const setServiceAccountName = useCallback(
    serviceAccountName => {
      setJob(j => ({
        ...j,
        config: {
          ...j.config,
          service_account_name: serviceAccountName
        }
      }));
    },
    [setJob]
  );

  const setResourceRequest = useCallback(
    (field, value) => {
      setJob(j => ({
        ...j,
        config: {
          ...j.config,
          resource_request: {
            ...j.config.resource_request,
            [field]: value
          }
        }
      }));
    },
    [setJob]
  );

  const setModel = useCallback(
    (field, value) => {
      setJob(j => ({
        ...j,
        config: {
          ...j.config,
          job_config: {
            ...j.config.job_config,
            model: {
              ...j.config.job_config.model,
              [field]: value
            }
          }
        }
      }));
    },
    [setJob]
  );

  const setModelResult = useCallback(
    (field, value) => {
      setJob(j => ({
        ...j,
        config: {
          ...j.config,
          job_config: {
            ...j.config.job_config,
            model: {
              ...j.config.job_config.model,
              result: {
                ...j.config.job_config.model.result,
                [field]: value
              }
            }
          }
        }
      }));
    },
    [setJob]
  );

  const setBigquerySource = useCallback(
    (field, value) => {
      setJob(j => ({
        ...j,
        config: {
          ...j.config,
          job_config: {
            ...j.config.job_config,
            bigquerySource: {
              ...j.config.job_config.bigquerySource,
              [field]: value
            }
          }
        }
      }));
    },
    [setJob]
  );

  const setBigquerySourceOptions = useCallback(
    (field, value) => {
      setJob(j => ({
        ...j,
        config: {
          ...j.config,
          job_config: {
            ...j.config.job_config,
            bigquerySource: {
              ...j.config.job_config.bigquerySource,
              options: {
                ...j.config.job_config.bigquerySource.options,
                [field]: String(value)
              }
            }
          }
        }
      }));
    },
    [setJob]
  );

  const unsetBigquerySourceOptions = useCallback(
    field => {
      setJob(j => {
        delete j.config.job_config.bigquerySource.options[field];
        return j;
      });
    },
    [setJob]
  );

  const setBigquerySink = useCallback(
    (field, value) => {
      setJob(j => ({
        ...j,
        config: {
          ...j.config,
          job_config: {
            ...j.config.job_config,
            bigquerySink: {
              ...j.config.job_config.bigquerySink,
              [field]: value
            }
          }
        }
      }));
    },
    [setJob]
  );

  const setBigquerySinkOptions = useCallback(
    (field, value) => {
      setJob(j => ({
        ...j,
        config: {
          ...j.config,
          job_config: {
            ...j.config.job_config,
            bigquerySink: {
              ...j.config.job_config.bigquerySink,
              options: {
                ...j.config.job_config.bigquerySink.options,
                [field]: String(value)
              }
            }
          }
        }
      }));
    },
    [setJob]
  );

  const setEnvVars = useCallback(
    newEnvVars => {
      setJob(j => ({
        ...j,
        config: {
          ...j.config,
          env_vars:
            JSON.stringify(newEnvVars) !== JSON.stringify(j.config.env_vars)
              ? newEnvVars
              : j.config.env_vars
        }
      }));
    },
    [setJob]
  );

  return (
    <JobFormContext.Provider
      value={{
        job,
        setVersionId,
        setServiceAccountName,
        setResourceRequest,
        setModel,
        setModelResult,
        setBigquerySource,
        setBigquerySourceOptions,
        unsetBigquerySourceOptions,
        setBigquerySink,
        setBigquerySinkOptions,
        setEnvVars
      }}>
      {props.children}
    </JobFormContext.Provider>
  );
};

JobFormContextProvider.propTypes = {
  initJob: PropTypes.object
};
