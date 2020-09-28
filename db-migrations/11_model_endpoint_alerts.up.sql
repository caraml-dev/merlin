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

CREATE TABLE IF NOT EXISTS model_endpoint_alerts (
    id                  serial      PRIMARY KEY,
    model_id            integer     REFERENCES models (id) NOT NULL,
    model_endpoint_id   integer     REFERENCES model_endpoints (id) NOT NULL,
    environment_name    varchar(50) REFERENCES environments (name) NOT NULL,
    team_name           varchar(25),
    alert_conditions    jsonb,
    created_at          timestamp   NOT NULL default current_timestamp,
    updated_at          timestamp   NOT NULL default current_timestamp
);

CREATE INDEX model_endpoint_alerts_idx_1 ON model_endpoint_alerts (
    model_id, model_endpoint_id
);
