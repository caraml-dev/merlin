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

CREATE TABLE IF NOT EXISTS environments
(
    id                  serial PRIMARY KEY,
    name                varchar(50)  NOT NULL,
    cluster             varchar(100) NOT NULL,
    is_default          boolean      DEFAULT false,
    created_at          timestamp    NOT NULL default current_timestamp,
    updated_at          timestamp    NOT NULL default current_timestamp,
    UNIQUE (name),

    CONSTRAINT name_not_empty CHECK (name <> ''),
    CONSTRAINT cluster_not_empty CHECK (cluster <> ''),
    -- constraint so that at most 1 environment is default
    -- however, it doesn't check if there is at least 1 is default
    CONSTRAINT is_default_true_or_null CHECK (is_default),
    CONSTRAINT is_default_only_1_true UNIQUE (is_default) DEFERRABLE
);

-- migrate version_endpoints and version table
-- so model version can have one to many relationship with version endpoint
ALTER TABLE version_endpoints ADD COLUMN version_id integer;
ALTER TABLE version_endpoints ADD COLUMN version_model_id integer;

ALTER TABLE version_endpoints
    ADD CONSTRAINT version_endpoints_model_version_fkey FOREIGN KEY (version_model_id, version_id) REFERENCES versions ( model_id , id);

UPDATE version_endpoints
SET version_id = versions.id, version_model_id = versions.model_id
FROM versions
WHERE version_endpoints.id = versions.endpoint_id;

ALTER TABLE version_endpoints ADD COLUMN environment_name varchar(50) REFERENCES environments (name);

ALTER TABLE model_endpoints ADD COLUMN environment_name varchar(50)  REFERENCES environments (name);

ALTER TABLE versions DROP COLUMN endpoint_id;
ALTER TABLE versions DROP COLUMN is_serving;