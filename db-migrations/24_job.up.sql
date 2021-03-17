CREATE TYPE job_status as ENUM ('pending', 'running', 'completed');

CREATE TABLE IF NOT EXISTS jobs (
    id bigserial PRIMARY KEY,
    name varchar(255),
    arguments jsonb,
    status          job_status NOT NULL default 'pending',
    created_at          timestamp NOT NULL default current_timestamp,
    updated_at          timestamp NOT NULL default current_timestamp, 
);

CREATE INDEX jobs_name_idx ON jobs(name);
CREATE INDEX jobs_status_idx ON jobs(status);