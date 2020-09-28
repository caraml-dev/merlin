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

CREATE TABLE IF NOT EXISTS deployments
(
    id                  serial PRIMARY KEY,
    version_id          integer,
    version_model_id    integer,
    project_id          integer,
    version_endpoint_id uuid,
    status              endpoint_status,
    error               text,
    created_at          timestamp NOT NULL default current_timestamp,
    updated_at          timestamp NOT NULL default current_timestamp,
    CONSTRAINT deployments_model_version_fkey
        FOREIGN KEY (version_model_id, version_id) REFERENCES versions (model_id, id),
    CONSTRAINT deployments_project_fkey
        FOREIGN KEY (project_id) REFERENCES projects (id),
    CONSTRAINT deployments_version_endpoint_fkey
        FOREIGN KEY (version_endpoint_id) REFERENCES version_endpoints (id)
);