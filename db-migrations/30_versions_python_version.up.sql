-- Introduce Python version column, with default as version 3.7 which is the
-- only supported major version as of the introduction of the column.
ALTER TABLE versions ADD COLUMN python_version varchar(16) NOT NULL DEFAULT '3.7.*';
-- Drop default after existing rows are updated
ALTER TABLE versions ALTER COLUMN python_version DROP DEFAULT;