ALTER TABLE versions ADD COLUMN labels jsonb;

CREATE INDEX version_labels_idx ON versions USING GIN (labels jsonb_path_ops);
