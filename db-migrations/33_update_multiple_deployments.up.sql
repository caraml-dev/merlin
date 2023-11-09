-- For terminated endpoint, update the status of all successful deployments to 'terminated'
update
	deployments
set
	status = 'terminated'
where
	id in (
	select
		d.id
	from
		deployments d
	join version_endpoints ve on
		d.version_endpoint_id = ve.id
	where
		ve.status = 'terminated'
		and d.status in ('pending', 'running', 'serving'));

-- For running/serving endpoint, update the status for all but the last successful deployment to 'terminated'
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
		and d2.status in ('pending', 'running', 'serving'));
