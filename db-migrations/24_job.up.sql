
CREATE TABLE IF NOT EXISTS jobs (
    id          bigserial PRIMARY KEY,
    name        varchar(255),
    arguments   jsonb,
    completed   boolean NOT NULL default false,
    created_at  timestamp NOT NULL default current_timestamp,
    updated_at  timestamp NOT NULL default current_timestamp
);

CREATE INDEX jobs_name_idx ON jobs(name) WHERE completed is not true;