ALTER TABLE version_endpoints
ADD COLUMN enable_model_observability BOOLEAN NOT NULL DEFAULT false;
