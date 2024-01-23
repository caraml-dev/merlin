CREATE TABLE IF NOT EXISTS model_schemas
(
    id                  serial PRIMARY KEY,
    model_id integer,
    spec JSONB NOT NULL,
    created_at              TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT model_schemas_model_fkey
        FOREIGN KEY (model_id) REFERENCES models (id)
);

CREATE INDEX model_schemas_model_id_idx ON model_schemas(model_id);
ALTER TABLE versions ADD COLUMN model_schema_id integer REFERENCES model_schemas (id);