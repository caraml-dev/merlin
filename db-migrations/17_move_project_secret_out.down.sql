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

ALTER TABLE models ADD CONSTRAINT models_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects (id);

ALTER TABLE secrets ADD CONSTRAINT secrets_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects (id);

ALTER TABLE prediction_jobs ADD CONSTRAINT prediction_jobs_project_fkey FOREIGN KEY (project_id) REFERENCES projects (id);

ALTER TABLE deployments ADD CONSTRAINT deployments_project_fkey FOREIGN KEY (project_id) REFERENCES projects (id);
