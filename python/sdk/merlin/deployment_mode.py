from enum import Enum


class DeploymentMode(Enum):
    """
    Deployment mode for deploying a model version
    """
    SERVERLESS = "serverless"
    RAW_DEPLOYMENT = "raw_deployment"
