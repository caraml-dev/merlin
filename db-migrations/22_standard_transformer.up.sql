ALTER TABLE transformers ADD COLUMN transformer_type VARCHAR(64);

UPDATE transformers
SET transformer_type = 'custom'
WHERE transformer_type is null;

ALTER TABLE environments ADD COLUMN default_transformer_resource_request jsonb;