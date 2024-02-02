-- Updating default protocol for version endpoints
update 
    version_endpoints
set protocol = 'HTTP_JSON'
where protocol is null or protocol = '';


-- Updating default protocol for model endpoints
update 
    model_endpoints
set protocol = 'HTTP_JSON'
where protocol is null or protocol = '';
