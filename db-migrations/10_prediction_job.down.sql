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

DROP TABLE IF EXISTS prediction_jobs CASCADE;

ALTER TABLE environments
    DROP COLUMN is_default_prediction_job;
ALTER TABLE environments
    DROP COLUMN is_prediction_job_enabled;
ALTER TABLE environments
    DROP COLUMN default_prediction_job_resource_request;