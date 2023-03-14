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


class ModelEndpointAlertCondition(object):
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
        'enabled': 'bool',
        'metric_type': 'AlertConditionMetricType',
        'severity': 'AlertConditionSeverity',
        'target': 'float',
        'percentile': 'float',
        'unit': 'str'
    }

    attribute_map = {
        'enabled': 'enabled',
        'metric_type': 'metric_type',
        'severity': 'severity',
        'target': 'target',
        'percentile': 'percentile',
        'unit': 'unit'
    }

    def __init__(self, enabled=None, metric_type=None, severity=None, target=None, percentile=None, unit=None):  # noqa: E501
        """ModelEndpointAlertCondition - a model defined in Swagger"""  # noqa: E501

        self._enabled = None
        self._metric_type = None
        self._severity = None
        self._target = None
        self._percentile = None
        self._unit = None
        self.discriminator = None

        if enabled is not None:
            self.enabled = enabled
        if metric_type is not None:
            self.metric_type = metric_type
        if severity is not None:
            self.severity = severity
        if target is not None:
            self.target = target
        if percentile is not None:
            self.percentile = percentile
        if unit is not None:
            self.unit = unit

    @property
    def enabled(self):
        """Gets the enabled of this ModelEndpointAlertCondition.  # noqa: E501


        :return: The enabled of this ModelEndpointAlertCondition.  # noqa: E501
        :rtype: bool
        """
        return self._enabled

    @enabled.setter
    def enabled(self, enabled):
        """Sets the enabled of this ModelEndpointAlertCondition.


        :param enabled: The enabled of this ModelEndpointAlertCondition.  # noqa: E501
        :type: bool
        """

        self._enabled = enabled

    @property
    def metric_type(self):
        """Gets the metric_type of this ModelEndpointAlertCondition.  # noqa: E501


        :return: The metric_type of this ModelEndpointAlertCondition.  # noqa: E501
        :rtype: AlertConditionMetricType
        """
        return self._metric_type

    @metric_type.setter
    def metric_type(self, metric_type):
        """Sets the metric_type of this ModelEndpointAlertCondition.


        :param metric_type: The metric_type of this ModelEndpointAlertCondition.  # noqa: E501
        :type: AlertConditionMetricType
        """

        self._metric_type = metric_type

    @property
    def severity(self):
        """Gets the severity of this ModelEndpointAlertCondition.  # noqa: E501


        :return: The severity of this ModelEndpointAlertCondition.  # noqa: E501
        :rtype: AlertConditionSeverity
        """
        return self._severity

    @severity.setter
    def severity(self, severity):
        """Sets the severity of this ModelEndpointAlertCondition.


        :param severity: The severity of this ModelEndpointAlertCondition.  # noqa: E501
        :type: AlertConditionSeverity
        """

        self._severity = severity

    @property
    def target(self):
        """Gets the target of this ModelEndpointAlertCondition.  # noqa: E501


        :return: The target of this ModelEndpointAlertCondition.  # noqa: E501
        :rtype: float
        """
        return self._target

    @target.setter
    def target(self, target):
        """Sets the target of this ModelEndpointAlertCondition.


        :param target: The target of this ModelEndpointAlertCondition.  # noqa: E501
        :type: float
        """

        self._target = target

    @property
    def percentile(self):
        """Gets the percentile of this ModelEndpointAlertCondition.  # noqa: E501


        :return: The percentile of this ModelEndpointAlertCondition.  # noqa: E501
        :rtype: float
        """
        return self._percentile

    @percentile.setter
    def percentile(self, percentile):
        """Sets the percentile of this ModelEndpointAlertCondition.


        :param percentile: The percentile of this ModelEndpointAlertCondition.  # noqa: E501
        :type: float
        """

        self._percentile = percentile

    @property
    def unit(self):
        """Gets the unit of this ModelEndpointAlertCondition.  # noqa: E501


        :return: The unit of this ModelEndpointAlertCondition.  # noqa: E501
        :rtype: str
        """
        return self._unit

    @unit.setter
    def unit(self, unit):
        """Sets the unit of this ModelEndpointAlertCondition.


        :param unit: The unit of this ModelEndpointAlertCondition.  # noqa: E501
        :type: str
        """

        self._unit = unit

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
        if issubclass(ModelEndpointAlertCondition, dict):
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
        if not isinstance(other, ModelEndpointAlertCondition):
            return False

        return self.__dict__ == other.__dict__

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        return not self == other
