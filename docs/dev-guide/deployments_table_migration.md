# Deployments Table Migration

This doc outline the data migration steps required for [PR #483 - Refactor deployments storage on redeploy and undeployment](https://github.com/caraml-dev/merlin/pull/483).

## Migration Scripts

There's two distinct scenario of deployments data to be migrated:

1. For version endpoints in terminated status, and
2. For version endpoints in pending/running/serving status

### For terminated version endpoints

If endpoint status = terminated, update all succesful deployments status to terminated

```
UPDATE
	deployments
SET
	status = 'terminated'
WHERE
	id IN (
	SELECT
		d.id
	FROM
		deployments d
	JOIN version_endpoints ve ON
		d.version_endpoint_id = ve.id
	WHERE
		ve.status = 'terminated'
		AND d.status IN ('pending', 'running', 'serving')
)
```

### For running/serving version endpoint
