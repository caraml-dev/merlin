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

Scale-down-to-zero is a feature in Merlin where model deployment that doesn’t receive traffic within 10 minutes windows will be scaled down and have 0 replicas. The model needs to receive HTTP traffic to scale up and become available again.

This feature is enabled for both production and staging, however in production the default minimum number of replica is set to 2, so user need to explicitly set the minimum replica to 0 during deployment in order to opt-in for this feature.

### Logs

### Log History

Users can only view the logs that are still in the model’s container. Link to the associated stackdriver dashboard is provided in the log page to access past log.
