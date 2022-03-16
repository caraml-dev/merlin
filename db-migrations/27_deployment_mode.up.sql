CREATE TYPE deployment_mode as ENUM ('raw_deployment', 'serverless');
ALTER TABLE version_endpoints ADD COLUMN deployment_mode deployment_mode NOT NULL default 'serverless';
