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


class StandardTransformerSimulationResponse(object):
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
        'response': 'FreeFormObject',
        'operation_tracing': 'OperationTracing'
    }

    attribute_map = {
        'response': 'response',
        'operation_tracing': 'operation_tracing'
    }

    def __init__(self, response=None, operation_tracing=None):  # noqa: E501
        """StandardTransformerSimulationResponse - a model defined in Swagger"""  # noqa: E501

        self._response = None
        self._operation_tracing = None
        self.discriminator = None

        if response is not None:
            self.response = response
        if operation_tracing is not None:
            self.operation_tracing = operation_tracing

    @property
    def response(self):
        """Gets the response of this StandardTransformerSimulationResponse.  # noqa: E501


        :return: The response of this StandardTransformerSimulationResponse.  # noqa: E501
        :rtype: FreeFormObject
        """
        return self._response

    @response.setter
    def response(self, response):
        """Sets the response of this StandardTransformerSimulationResponse.


        :param response: The response of this StandardTransformerSimulationResponse.  # noqa: E501
        :type: FreeFormObject
        """

        self._response = response

    @property
    def operation_tracing(self):
        """Gets the operation_tracing of this StandardTransformerSimulationResponse.  # noqa: E501


        :return: The operation_tracing of this StandardTransformerSimulationResponse.  # noqa: E501
        :rtype: OperationTracing
        """
        return self._operation_tracing

    @operation_tracing.setter
    def operation_tracing(self, operation_tracing):
        """Sets the operation_tracing of this StandardTransformerSimulationResponse.


        :param operation_tracing: The operation_tracing of this StandardTransformerSimulationResponse.  # noqa: E501
        :type: OperationTracing
        """

        self._operation_tracing = operation_tracing

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
        if issubclass(StandardTransformerSimulationResponse, dict):
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
        if not isinstance(other, StandardTransformerSimulationResponse):
            return False

        return self.__dict__ == other.__dict__

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        return not self == other
