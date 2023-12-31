<!-- page-title: Merlin Limitations -->
# Limitations

This article is an aggregation of the limits imposed on various components of the Merlin platform.

## Project

### Project Name

A project name can only contain letters `a-z` (lowercase), numbers `0-9` and the dash `-` symbol. The maximum length of a project name is `50` characters.

An example of a valid project name would be `gojek-project-01`.

## Model

### Model Name

A model name can only contain letters `a-z` (lowercase), numbers `0-9` and the dash `-` symbol. The maximum length of a model name is `25` characters.

An example of a valid model name would be `gojek-model-01`.

### Model Deployment

The maximum number of model versions that can be deployed in an environment is `2` per model.

### Resources

The maximum amount of CPU cores that can be allocated to a model is `4`.

The maximum amount of memory that can be allocated to a model is `8GB`.

## Autoscaling Policy

### Autoscaling

Autoscaling is enabled for both staging and production environment. User can set minimum and maximum number of replica during deployment.

### Scale Down to Zero

"Scaling down to zero" is a feature in Merlin, which automatically reduces the number of model deployments to zero when they haven't received any traffic for 10 minutes. To make the model available again, it must receive HTTP traffic, which triggers a scale-up.

This feature is only applicable when your autoscaling policy is set to either `RPS` or `Concurrency`."

Note that, to utilise this feature, the minimum replicas for the deployment should be set to `0`.


### Logs

### Log History

Users can only view the logs that are still in the model’s container. Link to the associated stackdriver dashboard is provided in the log page to access past log.