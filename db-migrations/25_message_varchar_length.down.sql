ALTER TABLE version_endpoints ALTER COLUMN message TYPE varchar(64) USING SUBSTR(message, 1, 64);
