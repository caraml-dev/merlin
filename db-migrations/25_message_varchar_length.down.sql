ALTER TABLE version_endpoints DROP COLUMN message;
ALTER TABLE version_endpoints ADD COLUMN message varchar(64);
