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

CREATE TYPE endpoint_status as ENUM ('pending', 'running', 'serving', 'failed', 'terminated');

CREATE TABLE IF NOT EXISTS projects
(
    id                  bigserial PRIMARY KEY,
    name                varchar(50)  NOT NULL,
    mlflow_tracking_url varchar(100) NOT NULL,
    k8s_cluster_name    varchar(100) NOT NULL default '',
    created_at          timestamp    NOT NULL default current_timestamp,
    updated_at          timestamp    NOT NULL default current_timestamp,
    UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS models
(
    id                   serial PRIMARY KEY,
    project_id           integer REFERENCES projects (id)           NOT NULL,
    name                 varchar(50)                                NOT NULL,
    type                 varchar(20)                                NOT NULL default 'other',
    mlflow_experiment_id integer                                    NOT NULL,
    endpoint_id          integer,
    created_at           timestamp                                  NOT NULL default current_timestamp,
    updated_at           timestamp                                  NOT NULL default current_timestamp,
    UNIQUE (project_id, name)
);

CREATE TABLE IF NOT EXISTS model_endpoints
(
    id              serial PRIMARY KEY,
    model_id        integer REFERENCES models (id) NOT NULL,
    status          endpoint_status         NOT NULL default 'pending',
    url             varchar(128)            NOT NULL,
    rule            jsonb,
    created_at      timestamp               NOT NULL default current_timestamp,
    updated_at      timestamp               NOT NULL default current_timestamp
);

CREATE TABLE IF NOT EXISTS versions
(
    id            integer                        NOT NULL,
    model_id      integer REFERENCES models (id) NOT NULL,
    is_serving    boolean                        NOT NULL,
    mlflow_run_id varchar(32)                    NOT NULL,
    artifact_uri  varchar(100)                   NOT NULL,
    endpoint_id   uuid,
    properties    jsonb,
    created_at    timestamp                      NOT NULL default current_timestamp,
    updated_at    timestamp                      NOT NULL default current_timestamp,
    CONSTRAINT pk_version_id PRIMARY KEY (id, model_id),
    UNIQUE (endpoint_id)
);

CREATE TABLE IF NOT EXISTS version_endpoints
(
    id            uuid                                PRIMARY KEY,
    service_name varchar(256)                         NOT NULL default '',
    inference_service_name varchar(256)               NOT NULL default '',
    namespace     varchar(64)                         NOT NULL default '',
    url           varchar(256)                        default '',
    status        endpoint_status                     NOT NULL default 'pending',
    created_at    timestamp                           NOT NULL default current_timestamp,
    updated_at    timestamp                           NOT NULL default current_timestamp
);

ALTER TABLE versions
    ADD CONSTRAINT fk_endpoints FOREIGN KEY (endpoint_id) REFERENCES version_endpoints (id);
