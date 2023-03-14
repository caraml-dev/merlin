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


class ModelEndpoint(object):
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
        'id': 'int',
        'model_id': 'int',
        'model': 'Model',
        'status': 'EndpointStatus',
        'url': 'str',
        'rule': 'ModelEndpointRule',
        'environment_name': 'str',
        'environment': 'Environment',
        'protocol': 'Protocol',
        'created_at': 'datetime',
        'updated_at': 'datetime'
    }

    attribute_map = {
        'id': 'id',
        'model_id': 'model_id',
        'model': 'model',
        'status': 'status',
        'url': 'url',
        'rule': 'rule',
        'environment_name': 'environment_name',
        'environment': 'environment',
        'protocol': 'protocol',
        'created_at': 'created_at',
        'updated_at': 'updated_at'
    }

    def __init__(self, id=None, model_id=None, model=None, status=None, url=None, rule=None, environment_name=None, environment=None, protocol=None, created_at=None, updated_at=None):  # noqa: E501
        """ModelEndpoint - a model defined in Swagger"""  # noqa: E501

        self._id = None
        self._model_id = None
        self._model = None
        self._status = None
        self._url = None
        self._rule = None
        self._environment_name = None
        self._environment = None
        self._protocol = None
        self._created_at = None
        self._updated_at = None
        self.discriminator = None

        if id is not None:
            self.id = id
        if model_id is not None:
            self.model_id = model_id
        if model is not None:
            self.model = model
        if status is not None:
            self.status = status
        if url is not None:
            self.url = url
        if rule is not None:
            self.rule = rule
        if environment_name is not None:
            self.environment_name = environment_name
        if environment is not None:
            self.environment = environment
        if protocol is not None:
            self.protocol = protocol
        if created_at is not None:
            self.created_at = created_at
        if updated_at is not None:
            self.updated_at = updated_at

    @property
    def id(self):
        """Gets the id of this ModelEndpoint.  # noqa: E501


        :return: The id of this ModelEndpoint.  # noqa: E501
        :rtype: int
        """
        return self._id

    @id.setter
    def id(self, id):
        """Sets the id of this ModelEndpoint.


        :param id: The id of this ModelEndpoint.  # noqa: E501
        :type: int
        """

        self._id = id

    @property
    def model_id(self):
        """Gets the model_id of this ModelEndpoint.  # noqa: E501


        :return: The model_id of this ModelEndpoint.  # noqa: E501
        :rtype: int
        """
        return self._model_id

    @model_id.setter
    def model_id(self, model_id):
        """Sets the model_id of this ModelEndpoint.


        :param model_id: The model_id of this ModelEndpoint.  # noqa: E501
        :type: int
        """

        self._model_id = model_id

    @property
    def model(self):
        """Gets the model of this ModelEndpoint.  # noqa: E501


        :return: The model of this ModelEndpoint.  # noqa: E501
        :rtype: Model
        """
        return self._model

    @model.setter
    def model(self, model):
        """Sets the model of this ModelEndpoint.


        :param model: The model of this ModelEndpoint.  # noqa: E501
        :type: Model
        """

        self._model = model

    @property
    def status(self):
        """Gets the status of this ModelEndpoint.  # noqa: E501


        :return: The status of this ModelEndpoint.  # noqa: E501
        :rtype: EndpointStatus
        """
        return self._status

    @status.setter
    def status(self, status):
        """Sets the status of this ModelEndpoint.


        :param status: The status of this ModelEndpoint.  # noqa: E501
        :type: EndpointStatus
        """

        self._status = status

    @property
    def url(self):
        """Gets the url of this ModelEndpoint.  # noqa: E501


        :return: The url of this ModelEndpoint.  # noqa: E501
        :rtype: str
        """
        return self._url

    @url.setter
    def url(self, url):
        """Sets the url of this ModelEndpoint.


        :param url: The url of this ModelEndpoint.  # noqa: E501
        :type: str
        """

        self._url = url

    @property
    def rule(self):
        """Gets the rule of this ModelEndpoint.  # noqa: E501


        :return: The rule of this ModelEndpoint.  # noqa: E501
        :rtype: ModelEndpointRule
        """
        return self._rule

    @rule.setter
    def rule(self, rule):
        """Sets the rule of this ModelEndpoint.


        :param rule: The rule of this ModelEndpoint.  # noqa: E501
        :type: ModelEndpointRule
        """

        self._rule = rule

    @property
    def environment_name(self):
        """Gets the environment_name of this ModelEndpoint.  # noqa: E501


        :return: The environment_name of this ModelEndpoint.  # noqa: E501
        :rtype: str
        """
        return self._environment_name

    @environment_name.setter
    def environment_name(self, environment_name):
        """Sets the environment_name of this ModelEndpoint.


        :param environment_name: The environment_name of this ModelEndpoint.  # noqa: E501
        :type: str
        """

        self._environment_name = environment_name

    @property
    def environment(self):
        """Gets the environment of this ModelEndpoint.  # noqa: E501


        :return: The environment of this ModelEndpoint.  # noqa: E501
        :rtype: Environment
        """
        return self._environment

    @environment.setter
    def environment(self, environment):
        """Sets the environment of this ModelEndpoint.


        :param environment: The environment of this ModelEndpoint.  # noqa: E501
        :type: Environment
        """

        self._environment = environment

    @property
    def protocol(self):
        """Gets the protocol of this ModelEndpoint.  # noqa: E501


        :return: The protocol of this ModelEndpoint.  # noqa: E501
        :rtype: Protocol
        """
        return self._protocol

    @protocol.setter
    def protocol(self, protocol):
        """Sets the protocol of this ModelEndpoint.


        :param protocol: The protocol of this ModelEndpoint.  # noqa: E501
        :type: Protocol
        """

        self._protocol = protocol

    @property
    def created_at(self):
        """Gets the created_at of this ModelEndpoint.  # noqa: E501


        :return: The created_at of this ModelEndpoint.  # noqa: E501
        :rtype: datetime
        """
        return self._created_at

    @created_at.setter
    def created_at(self, created_at):
        """Sets the created_at of this ModelEndpoint.


        :param created_at: The created_at of this ModelEndpoint.  # noqa: E501
        :type: datetime
        """

        self._created_at = created_at

    @property
    def updated_at(self):
        """Gets the updated_at of this ModelEndpoint.  # noqa: E501


        :return: The updated_at of this ModelEndpoint.  # noqa: E501
        :rtype: datetime
        """
        return self._updated_at

    @updated_at.setter
    def updated_at(self, updated_at):
        """Sets the updated_at of this ModelEndpoint.


        :param updated_at: The updated_at of this ModelEndpoint.  # noqa: E501
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
        if issubclass(ModelEndpoint, dict):
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
        if not isinstance(other, ModelEndpoint):
            return False

        return self.__dict__ == other.__dict__

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        return not self == other
