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

CREATE TABLE IF NOT EXISTS transformers
(
    id                      SERIAL      PRIMARY KEY,
    version_endpoint_id     UUID        NOT NULL,
    image                   VARCHAR     NOT NULL,
    command                 VARCHAR,
    args                    VARCHAR,
    resource_request        JSONB,
    env_vars                JSONB,
    enabled                 BOOLEAN     NOT NULL DEFAULT false,
    created_at              TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(version_endpoint_id)
);
