# coding: utf-8

"""
    Merlin

    API Guide for accessing Merlin's model management, deployment, and serving functionalities  # noqa: E501

    OpenAPI spec version: 0.14.0
    
    Generated by: https://github.com/swagger-api/swagger-codegen.git
"""

import pprint
import re  # noqa: F401

import six

class VersionEndpoint(object):
    """NOTE: This class is auto generated by the swagger code generator program.

    Do not edit the class manually.
    """
    """
    Attributes:
      swagger_types (dict): The key is attribute name
                            and the value is attribute type.
      attribute_map (dict): The key is attribute name
                            and the value is json key in definition.
    """
    swagger_types = {
        'id': 'str',
        'version_id': 'int',
        'status': 'EndpointStatus',
        'url': 'str',
        'service_name': 'str',
        'environment_name': 'str',
        'environment': 'Environment',
        'monitoring_url': 'str',
        'message': 'str',
        'resource_request': 'ResourceRequest',
        'image_builder_resource_request': 'ResourceRequest',
        'env_vars': 'list[EnvVar]',
        'transformer': 'Transformer',
        'logger': 'Logger',
        'deployment_mode': 'DeploymentMode',
        'autoscaling_policy': 'AutoscalingPolicy',
        'protocol': 'Protocol',
        'created_at': 'datetime',
        'updated_at': 'datetime'
    }

    attribute_map = {
        'id': 'id',
        'version_id': 'version_id',
        'status': 'status',
        'url': 'url',
        'service_name': 'service_name',
        'environment_name': 'environment_name',
        'environment': 'environment',
        'monitoring_url': 'monitoring_url',
        'message': 'message',
        'resource_request': 'resource_request',
        'image_builder_resource_request': 'image_builder_resource_request',
        'env_vars': 'env_vars',
        'transformer': 'transformer',
        'logger': 'logger',
        'deployment_mode': 'deployment_mode',
        'autoscaling_policy': 'autoscaling_policy',
        'protocol': 'protocol',
        'created_at': 'created_at',
        'updated_at': 'updated_at'
    }

    def __init__(self, id=None, version_id=None, status=None, url=None, service_name=None, environment_name=None, environment=None, monitoring_url=None, message=None, resource_request=None, image_builder_resource_request=None, env_vars=None, transformer=None, logger=None, deployment_mode=None, autoscaling_policy=None, protocol=None, created_at=None, updated_at=None):  # noqa: E501
        """VersionEndpoint - a model defined in Swagger"""  # noqa: E501
        self._id = None
        self._version_id = None
        self._status = None
        self._url = None
        self._service_name = None
        self._environment_name = None
        self._environment = None
        self._monitoring_url = None
        self._message = None
        self._resource_request = None
        self._image_builder_resource_request = None
        self._env_vars = None
        self._transformer = None
        self._logger = None
        self._deployment_mode = None
        self._autoscaling_policy = None
        self._protocol = None
        self._created_at = None
        self._updated_at = None
        self.discriminator = None
        if id is not None:
            self.id = id
        if version_id is not None:
            self.version_id = version_id
        if status is not None:
            self.status = status
        if url is not None:
            self.url = url
        if service_name is not None:
            self.service_name = service_name
        if environment_name is not None:
            self.environment_name = environment_name
        if environment is not None:
            self.environment = environment
        if monitoring_url is not None:
            self.monitoring_url = monitoring_url
        if message is not None:
            self.message = message
        if resource_request is not None:
            self.resource_request = resource_request
        if image_builder_resource_request is not None:
            self.image_builder_resource_request = image_builder_resource_request
        if env_vars is not None:
            self.env_vars = env_vars
        if transformer is not None:
            self.transformer = transformer
        if logger is not None:
            self.logger = logger
        if deployment_mode is not None:
            self.deployment_mode = deployment_mode
        if autoscaling_policy is not None:
            self.autoscaling_policy = autoscaling_policy
        if protocol is not None:
            self.protocol = protocol
        if created_at is not None:
            self.created_at = created_at
        if updated_at is not None:
            self.updated_at = updated_at

    @property
    def id(self):
        """Gets the id of this VersionEndpoint.  # noqa: E501


        :return: The id of this VersionEndpoint.  # noqa: E501
        :rtype: str
        """
        return self._id

    @id.setter
    def id(self, id):
        """Sets the id of this VersionEndpoint.


        :param id: The id of this VersionEndpoint.  # noqa: E501
        :type: str
        """

        self._id = id

    @property
    def version_id(self):
        """Gets the version_id of this VersionEndpoint.  # noqa: E501


        :return: The version_id of this VersionEndpoint.  # noqa: E501
        :rtype: int
        """
        return self._version_id

    @version_id.setter
    def version_id(self, version_id):
        """Sets the version_id of this VersionEndpoint.


        :param version_id: The version_id of this VersionEndpoint.  # noqa: E501
        :type: int
        """

        self._version_id = version_id

    @property
    def status(self):
        """Gets the status of this VersionEndpoint.  # noqa: E501


        :return: The status of this VersionEndpoint.  # noqa: E501
        :rtype: EndpointStatus
        """
        return self._status

    @status.setter
    def status(self, status):
        """Sets the status of this VersionEndpoint.


        :param status: The status of this VersionEndpoint.  # noqa: E501
        :type: EndpointStatus
        """

        self._status = status

    @property
    def url(self):
        """Gets the url of this VersionEndpoint.  # noqa: E501


        :return: The url of this VersionEndpoint.  # noqa: E501
        :rtype: str
        """
        return self._url

    @url.setter
    def url(self, url):
        """Sets the url of this VersionEndpoint.


        :param url: The url of this VersionEndpoint.  # noqa: E501
        :type: str
        """

        self._url = url

    @property
    def service_name(self):
        """Gets the service_name of this VersionEndpoint.  # noqa: E501


        :return: The service_name of this VersionEndpoint.  # noqa: E501
        :rtype: str
        """
        return self._service_name

    @service_name.setter
    def service_name(self, service_name):
        """Sets the service_name of this VersionEndpoint.


        :param service_name: The service_name of this VersionEndpoint.  # noqa: E501
        :type: str
        """

        self._service_name = service_name

    @property
    def environment_name(self):
        """Gets the environment_name of this VersionEndpoint.  # noqa: E501


        :return: The environment_name of this VersionEndpoint.  # noqa: E501
        :rtype: str
        """
        return self._environment_name

    @environment_name.setter
    def environment_name(self, environment_name):
        """Sets the environment_name of this VersionEndpoint.


        :param environment_name: The environment_name of this VersionEndpoint.  # noqa: E501
        :type: str
        """

        self._environment_name = environment_name

    @property
    def environment(self):
        """Gets the environment of this VersionEndpoint.  # noqa: E501


        :return: The environment of this VersionEndpoint.  # noqa: E501
        :rtype: Environment
        """
        return self._environment

    @environment.setter
    def environment(self, environment):
        """Sets the environment of this VersionEndpoint.


        :param environment: The environment of this VersionEndpoint.  # noqa: E501
        :type: Environment
        """

        self._environment = environment

    @property
    def monitoring_url(self):
        """Gets the monitoring_url of this VersionEndpoint.  # noqa: E501


        :return: The monitoring_url of this VersionEndpoint.  # noqa: E501
        :rtype: str
        """
        return self._monitoring_url

    @monitoring_url.setter
    def monitoring_url(self, monitoring_url):
        """Sets the monitoring_url of this VersionEndpoint.


        :param monitoring_url: The monitoring_url of this VersionEndpoint.  # noqa: E501
        :type: str
        """

        self._monitoring_url = monitoring_url

    @property
    def message(self):
        """Gets the message of this VersionEndpoint.  # noqa: E501


        :return: The message of this VersionEndpoint.  # noqa: E501
        :rtype: str
        """
        return self._message

    @message.setter
    def message(self, message):
        """Sets the message of this VersionEndpoint.


        :param message: The message of this VersionEndpoint.  # noqa: E501
        :type: str
        """

        self._message = message

    @property
    def resource_request(self):
        """Gets the resource_request of this VersionEndpoint.  # noqa: E501


        :return: The resource_request of this VersionEndpoint.  # noqa: E501
        :rtype: ResourceRequest
        """
        return self._resource_request

    @resource_request.setter
    def resource_request(self, resource_request):
        """Sets the resource_request of this VersionEndpoint.


        :param resource_request: The resource_request of this VersionEndpoint.  # noqa: E501
        :type: ResourceRequest
        """

        self._resource_request = resource_request

    @property
    def image_builder_resource_request(self):
        """Gets the image_builder_resource_request of this VersionEndpoint.  # noqa: E501


        :return: The image_builder_resource_request of this VersionEndpoint.  # noqa: E501
        :rtype: ResourceRequest
        """
        return self._image_builder_resource_request

    @image_builder_resource_request.setter
    def image_builder_resource_request(self, image_builder_resource_request):
        """Sets the image_builder_resource_request of this VersionEndpoint.


        :param image_builder_resource_request: The image_builder_resource_request of this VersionEndpoint.  # noqa: E501
        :type: ResourceRequest
        """

        self._image_builder_resource_request = image_builder_resource_request

    @property
    def env_vars(self):
        """Gets the env_vars of this VersionEndpoint.  # noqa: E501


        :return: The env_vars of this VersionEndpoint.  # noqa: E501
        :rtype: list[EnvVar]
        """
        return self._env_vars

    @env_vars.setter
    def env_vars(self, env_vars):
        """Sets the env_vars of this VersionEndpoint.


        :param env_vars: The env_vars of this VersionEndpoint.  # noqa: E501
        :type: list[EnvVar]
        """

        self._env_vars = env_vars

    @property
    def transformer(self):
        """Gets the transformer of this VersionEndpoint.  # noqa: E501


        :return: The transformer of this VersionEndpoint.  # noqa: E501
        :rtype: Transformer
        """
        return self._transformer

    @transformer.setter
    def transformer(self, transformer):
        """Sets the transformer of this VersionEndpoint.


        :param transformer: The transformer of this VersionEndpoint.  # noqa: E501
        :type: Transformer
        """

        self._transformer = transformer

    @property
    def logger(self):
        """Gets the logger of this VersionEndpoint.  # noqa: E501


        :return: The logger of this VersionEndpoint.  # noqa: E501
        :rtype: Logger
        """
        return self._logger

    @logger.setter
    def logger(self, logger):
        """Sets the logger of this VersionEndpoint.


        :param logger: The logger of this VersionEndpoint.  # noqa: E501
        :type: Logger
        """

        self._logger = logger

    @property
    def deployment_mode(self):
        """Gets the deployment_mode of this VersionEndpoint.  # noqa: E501


        :return: The deployment_mode of this VersionEndpoint.  # noqa: E501
        :rtype: DeploymentMode
        """
        return self._deployment_mode

    @deployment_mode.setter
    def deployment_mode(self, deployment_mode):
        """Sets the deployment_mode of this VersionEndpoint.


        :param deployment_mode: The deployment_mode of this VersionEndpoint.  # noqa: E501
        :type: DeploymentMode
        """

        self._deployment_mode = deployment_mode

    @property
    def autoscaling_policy(self):
        """Gets the autoscaling_policy of this VersionEndpoint.  # noqa: E501


        :return: The autoscaling_policy of this VersionEndpoint.  # noqa: E501
        :rtype: AutoscalingPolicy
        """
        return self._autoscaling_policy

    @autoscaling_policy.setter
    def autoscaling_policy(self, autoscaling_policy):
        """Sets the autoscaling_policy of this VersionEndpoint.


        :param autoscaling_policy: The autoscaling_policy of this VersionEndpoint.  # noqa: E501
        :type: AutoscalingPolicy
        """

        self._autoscaling_policy = autoscaling_policy

    @property
    def protocol(self):
        """Gets the protocol of this VersionEndpoint.  # noqa: E501


        :return: The protocol of this VersionEndpoint.  # noqa: E501
        :rtype: Protocol
        """
        return self._protocol

    @protocol.setter
    def protocol(self, protocol):
        """Sets the protocol of this VersionEndpoint.


        :param protocol: The protocol of this VersionEndpoint.  # noqa: E501
        :type: Protocol
        """

        self._protocol = protocol

    @property
    def created_at(self):
        """Gets the created_at of this VersionEndpoint.  # noqa: E501


        :return: The created_at of this VersionEndpoint.  # noqa: E501
        :rtype: datetime
        """
        return self._created_at

    @created_at.setter
    def created_at(self, created_at):
        """Sets the created_at of this VersionEndpoint.


        :param created_at: The created_at of this VersionEndpoint.  # noqa: E501
        :type: datetime
        """

        self._created_at = created_at

    @property
    def updated_at(self):
        """Gets the updated_at of this VersionEndpoint.  # noqa: E501


        :return: The updated_at of this VersionEndpoint.  # noqa: E501
        :rtype: datetime
        """
        return self._updated_at

    @updated_at.setter
    def updated_at(self, updated_at):
        """Sets the updated_at of this VersionEndpoint.


        :param updated_at: The updated_at of this VersionEndpoint.  # noqa: E501
        :type: datetime
        """

        self._updated_at = updated_at

    def to_dict(self):
        """Returns the model properties as a dict"""
        result = {}

        for attr, _ in six.iteritems(self.swagger_types):
            value = getattr(self, attr)
            if isinstance(value, list):
                result[attr] = list(map(
                    lambda x: x.to_dict() if hasattr(x, "to_dict") else x,
                    value
                ))
            elif hasattr(value, "to_dict"):
                result[attr] = value.to_dict()
            elif isinstance(value, dict):
                result[attr] = dict(map(
                    lambda item: (item[0], item[1].to_dict())
                    if hasattr(item[1], "to_dict") else item,
                    value.items()
                ))
            else:
                result[attr] = value
        if issubclass(VersionEndpoint, dict):
            for key, value in self.items():
                result[key] = value

        return result

    def to_str(self):
        """Returns the string representation of the model"""
        return pprint.pformat(self.to_dict())

    def __repr__(self):
        """For `print` and `pprint`"""
        return self.to_str()

    def __eq__(self, other):
        """Returns true if both objects are equal"""
        if not isinstance(other, VersionEndpoint):
            return False

        return self.__dict__ == other.__dict__

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        return not self == other
