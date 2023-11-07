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

To get/verify the list of deployments to be updated, we can run this query:

```
select
	d.id,
	d.version_endpoint_id,
	d.status
from
	deployments d
where
	d.status in ('pending', 'running', 'serving')
	and exists (
	select
		1
	from
		deployments d2
	where
		d.version_endpoint_id = d2.version_endpoint_id
		and d.id < d2.id
		and d2.status in ('pending', 'running', 'serving')
)
order by
	1 desc;
```

Then we can take some version_endpoint_id as samples to veify the given version endpoint has multiple deployments and the query above returns previous successful deployments:

```
select * from deployments d where d.version_endpoint_id = '<version_endpoint_id from query above>';
```

Finally, using WHERE condition above, update the deployments table to set the status to 'terminated' for all but the last deployment for each version endpoints.

```
update deployments as d
set status = 'terminated'
where
	d.status in ('pending', 'running', 'serving')
	and exists (
	select
		1
	from
		deployments d2
	where
		d.version_endpoint_id = d2.version_endpoint_id
		and d.id < d2.id
		and d2.status in ('pending', 'running', 'serving')
);
```
