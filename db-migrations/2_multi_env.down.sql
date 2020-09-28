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

DROP TYPE IF EXISTS environments CASCADE;


ALTER TABLE versions ADD COLUMN endpoint_id integer REFERENCES version_endpoints (id);

UPDATE versions
SET versions.endpoint_id = version_endpoints.id
FROM version_endpoints
WHERE version_endpoints.version_id = versions.id;

ALTER TABLE version_endpoints DROP COLUMN version_id;
ALTER TABLE version_endpoints DROP COLUMN environment_name;


ALTER TABLE models ADD COLUMN endpoint_id integer REFERENCES model_endpoints (id);

UPDATE models
SET models.endpoint_id = model_endpoints.id
FROM model_endpoints
WHERE model_endpoints.model_id = models.id;

ALTER TABLE model_endpoints DROP COLUMN model_id;
ALTER TABLE model_endpoints DROP COLUMN environment_name;


ALTER TABLE versions ADD COLUMN is_serving boolean;