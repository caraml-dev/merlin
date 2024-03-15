CREATE TYPE publisher_status as ENUM ('pending', 'running', 'failed', 'terminated');

CREATE TABLE IF NOT EXISTS observability_publishers
(
    id                  serial PRIMARY KEY,
    version_model_id    integer,
    version_id          integer,
    revision            integer,
    status              publisher_status,
    model_schema_spec   JSONB, 
    created_at          TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(version_model_id)
);