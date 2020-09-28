# Model Deployment and Serving

Model deployment in Merlin is a process of creating a model service and it's [Model Version Endpoint](./model_version_endpoint.md). Internally, the deployment of the Model Version Endpoint is done via [kaniko](https://github.com/GoogleContainerTools/kaniko) and [KFServing](https://github.com/kubeflow/kfserving).

There are two types of Model Version deployment, standard and python function (PyFunc) deployment. The difference is PyFunc deployment includes Docker image building step by Kaniko.

Model serving is the next step of model deployment. After we have a running Model Version Endpoint, we can start serving the HTTP traffic by routing the Model Endpoint to it.

![Model Deployment and Serving](../diagrams/model_deployment_serving.drawio.svg)
