-- Copyright 2020 The Merlin Authors
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--      http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.

CREATE TYPE prediction_job_status as ENUM ('pending', 'running', 'terminating', 'completed', 'failed', 'terminated', 'failed_submission');

CREATE TABLE IF NOT EXISTS prediction_jobs
(
    id               serial PRIMARY KEY,
    name             varchar(512),
    version_id       integer,
    version_model_id integer,
    project_id       integer,
    config           jsonb,
    status           prediction_job_status NOT NULL default 'pending',
    error            text,
    created_at       timestamp             NOT NULL default current_timestamp,
    updated_at       timestamp             NOT NULL default current_timestamp,
    CONSTRAINT prediction_jobs_model_version_fkey
        FOREIGN KEY (version_model_id, version_id) REFERENCES versions (model_id, id),
    CONSTRAINT prediction_jobs_project_fkey
        FOREIGN KEY (project_id) REFERENCES projects (id)
);

ALTER TABLE environments
    ADD COLUMN is_default_prediction_job boolean;
ALTER TABLE environments
    ADD COLUMN is_prediction_job_enabled boolean;
ALTER TABLE environments
    ADD COLUMN default_prediction_job_resource_request jsonb;
-- constraint so that at most 1 environment is default
-- however, it doesn't check if there is at least 1 is default
ALTER TABLE environments
    ADD CONSTRAINT is_default_prediction_job_true_or_null CHECK (is_default_prediction_job);
ALTER TABLE environments
    ADD CONSTRAINT is_default_prediction_job_only_1_true UNIQUE (is_default_prediction_job) DEFERRABLE