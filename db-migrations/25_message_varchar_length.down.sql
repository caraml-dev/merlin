ALTER TABLE version_endpoints ALTER COLUMN message varchar(64) USING SUBSTR(message, 1, 64);
