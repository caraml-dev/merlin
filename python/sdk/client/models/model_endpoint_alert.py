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


class ModelEndpointAlert(object):
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
        'model_id': 'int',
        'model_endpoint_id': 'int',
        'environment_name': 'str',
        'team_name': 'str',
        'alert_conditions': 'list[ModelEndpointAlertCondition]'
    }

    attribute_map = {
        'model_id': 'model_id',
        'model_endpoint_id': 'model_endpoint_id',
        'environment_name': 'environment_name',
        'team_name': 'team_name',
        'alert_conditions': 'alert_conditions'
    }

    def __init__(self, model_id=None, model_endpoint_id=None, environment_name=None, team_name=None, alert_conditions=None):  # noqa: E501
        """ModelEndpointAlert - a model defined in Swagger"""  # noqa: E501

        self._model_id = None
        self._model_endpoint_id = None
        self._environment_name = None
        self._team_name = None
        self._alert_conditions = None
        self.discriminator = None

        if model_id is not None:
            self.model_id = model_id
        if model_endpoint_id is not None:
            self.model_endpoint_id = model_endpoint_id
        if environment_name is not None:
            self.environment_name = environment_name
        if team_name is not None:
            self.team_name = team_name
        if alert_conditions is not None:
            self.alert_conditions = alert_conditions

    @property
    def model_id(self):
        """Gets the model_id of this ModelEndpointAlert.  # noqa: E501


        :return: The model_id of this ModelEndpointAlert.  # noqa: E501
        :rtype: int
        """
        return self._model_id

    @model_id.setter
    def model_id(self, model_id):
        """Sets the model_id of this ModelEndpointAlert.


        :param model_id: The model_id of this ModelEndpointAlert.  # noqa: E501
        :type: int
        """

        self._model_id = model_id

    @property
    def model_endpoint_id(self):
        """Gets the model_endpoint_id of this ModelEndpointAlert.  # noqa: E501


        :return: The model_endpoint_id of this ModelEndpointAlert.  # noqa: E501
        :rtype: int
        """
        return self._model_endpoint_id

    @model_endpoint_id.setter
    def model_endpoint_id(self, model_endpoint_id):
        """Sets the model_endpoint_id of this ModelEndpointAlert.


        :param model_endpoint_id: The model_endpoint_id of this ModelEndpointAlert.  # noqa: E501
        :type: int
        """

        self._model_endpoint_id = model_endpoint_id

    @property
    def environment_name(self):
        """Gets the environment_name of this ModelEndpointAlert.  # noqa: E501


        :return: The environment_name of this ModelEndpointAlert.  # noqa: E501
        :rtype: str
        """
        return self._environment_name

    @environment_name.setter
    def environment_name(self, environment_name):
        """Sets the environment_name of this ModelEndpointAlert.


        :param environment_name: The environment_name of this ModelEndpointAlert.  # noqa: E501
        :type: str
        """

        self._environment_name = environment_name

    @property
    def team_name(self):
        """Gets the team_name of this ModelEndpointAlert.  # noqa: E501


        :return: The team_name of this ModelEndpointAlert.  # noqa: E501
        :rtype: str
        """
        return self._team_name

    @team_name.setter
    def team_name(self, team_name):
        """Sets the team_name of this ModelEndpointAlert.


        :param team_name: The team_name of this ModelEndpointAlert.  # noqa: E501
        :type: str
        """

        self._team_name = team_name

    @property
    def alert_conditions(self):
        """Gets the alert_conditions of this ModelEndpointAlert.  # noqa: E501


        :return: The alert_conditions of this ModelEndpointAlert.  # noqa: E501
        :rtype: list[ModelEndpointAlertCondition]
        """
        return self._alert_conditions

    @alert_conditions.setter
    def alert_conditions(self, alert_conditions):
        """Sets the alert_conditions of this ModelEndpointAlert.


        :param alert_conditions: The alert_conditions of this ModelEndpointAlert.  # noqa: E501
        :type: list[ModelEndpointAlertCondition]
        """

        self._alert_conditions = alert_conditions

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
        if issubclass(ModelEndpointAlert, dict):
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
        if not isinstance(other, ModelEndpointAlert):
            return False

        return self.__dict__ == other.__dict__

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        return not self == other
