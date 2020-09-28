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

CREATE TABLE
IF NOT EXISTS secrets
(
  id serial PRIMARY KEY,
  project_id integer REFERENCES projects
(id) NOT NULL,
  name varchar
(100) NOT NULL,
  data text NOT NULL,
  created_at timestamp NOT NULL default current_timestamp,
  updated_at timestamp NOT NULL default current_timestamp,
  UNIQUE
(project_id, name)
);