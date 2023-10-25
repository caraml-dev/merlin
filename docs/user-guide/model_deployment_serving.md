# Model Deployment and Serving

Model deployment in Merlin is a process of creating a model service and it's [Model Version Endpoint](./model_version_endpoint.md). Internally, the deployment of the Model Version Endpoint is done via [kaniko](https://github.com/GoogleContainerTools/kaniko) and [KFServing](https://github.com/kubeflow/kfserving).

There are two types of Model Version deployment, standard and python function (PyFunc) deployment. The difference is PyFunc deployment includes Docker image building step by Kaniko.

Model serving is the next step of model deployment. After we have a running Model Version Endpoint, we can start serving the HTTP traffic by routing the Model Endpoint to it.

![Model Deployment and Serving](../diagrams/model_deployment_serving.drawio.svg)

# Model Versions and Deployments
Each model version itself can deployed with a different set of deployment configurations, such as the number of 
replicas, CPU/memory requests, autoscaling policy, environment variables, etc. Each set of these configurations that are
used to deploy a model version are called a *deployment*. 

While each model can have up to **2** model versions deployed at any point of time, each model version can only be 
deployed using **1** deployment at any point of time. 

Whenever a running model version is redeployed, a new *deployment* is created and the Merlin API server attempts to 
deploy the same model version with the new deployment configuration, all while keep the existing deployment running. 

If the deployment of the new configuration succeeds, the old deployment will be undeployed and the new *deployment* 
becomes the current *deployment* of the model version.

If the deployment of the new configuration fails, **the old deployment stays deployed** and remains as the current 
*deployment* of the model version. The new configuration will then show a 'Failed' status.
