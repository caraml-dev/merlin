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

class Environment(object):
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
        'name': 'str',
        'cluster': 'str',
        'is_default': 'bool',
        'region': 'str',
        'gcp_project': 'str',
        'default_resource_request': 'ResourceRequest',
        'default_transformer_resource_request': 'ResourceRequest',
        'default_prediction_job_resource_request': 'PredictionJobResourceRequest',
        'gpus': 'list[GPU]',
        'created_at': 'datetime',
        'updated_at': 'datetime'
    }

    attribute_map = {
        'id': 'id',
        'name': 'name',
        'cluster': 'cluster',
        'is_default': 'is_default',
        'region': 'region',
        'gcp_project': 'gcp_project',
        'default_resource_request': 'default_resource_request',
        'default_transformer_resource_request': 'default_transformer_resource_request',
        'default_prediction_job_resource_request': 'default_prediction_job_resource_request',
        'gpus': 'gpus',
        'created_at': 'created_at',
        'updated_at': 'updated_at'
    }

    def __init__(self, id=None, name=None, cluster=None, is_default=None, region=None, gcp_project=None, default_resource_request=None, default_transformer_resource_request=None, default_prediction_job_resource_request=None, gpus=None, created_at=None, updated_at=None):  # noqa: E501
        """Environment - a model defined in Swagger"""  # noqa: E501
        self._id = None
        self._name = None
        self._cluster = None
        self._is_default = None
        self._region = None
        self._gcp_project = None
        self._default_resource_request = None
        self._default_transformer_resource_request = None
        self._default_prediction_job_resource_request = None
        self._gpus = None
        self._created_at = None
        self._updated_at = None
        self.discriminator = None
        if id is not None:
            self.id = id
        self.name = name
        if cluster is not None:
            self.cluster = cluster
        if is_default is not None:
            self.is_default = is_default
        if region is not None:
            self.region = region
        if gcp_project is not None:
            self.gcp_project = gcp_project
        if default_resource_request is not None:
            self.default_resource_request = default_resource_request
        if default_transformer_resource_request is not None:
            self.default_transformer_resource_request = default_transformer_resource_request
        if default_prediction_job_resource_request is not None:
            self.default_prediction_job_resource_request = default_prediction_job_resource_request
        if gpus is not None:
            self.gpus = gpus
        if created_at is not None:
            self.created_at = created_at
        if updated_at is not None:
            self.updated_at = updated_at

    @property
    def id(self):
        """Gets the id of this Environment.  # noqa: E501


        :return: The id of this Environment.  # noqa: E501
        :rtype: int
        """
        return self._id

    @id.setter
    def id(self, id):
        """Sets the id of this Environment.


        :param id: The id of this Environment.  # noqa: E501
        :type: int
        """

        self._id = id

    @property
    def name(self):
        """Gets the name of this Environment.  # noqa: E501


        :return: The name of this Environment.  # noqa: E501
        :rtype: str
        """
        return self._name

    @name.setter
    def name(self, name):
        """Sets the name of this Environment.


        :param name: The name of this Environment.  # noqa: E501
        :type: str
        """
        if name is None:
            raise ValueError("Invalid value for `name`, must not be `None`")  # noqa: E501

        self._name = name

    @property
    def cluster(self):
        """Gets the cluster of this Environment.  # noqa: E501


        :return: The cluster of this Environment.  # noqa: E501
        :rtype: str
        """
        return self._cluster

    @cluster.setter
    def cluster(self, cluster):
        """Sets the cluster of this Environment.


        :param cluster: The cluster of this Environment.  # noqa: E501
        :type: str
        """

        self._cluster = cluster

    @property
    def is_default(self):
        """Gets the is_default of this Environment.  # noqa: E501


        :return: The is_default of this Environment.  # noqa: E501
        :rtype: bool
        """
        return self._is_default

    @is_default.setter
    def is_default(self, is_default):
        """Sets the is_default of this Environment.


        :param is_default: The is_default of this Environment.  # noqa: E501
        :type: bool
        """

        self._is_default = is_default

    @property
    def region(self):
        """Gets the region of this Environment.  # noqa: E501


        :return: The region of this Environment.  # noqa: E501
        :rtype: str
        """
        return self._region

    @region.setter
    def region(self, region):
        """Sets the region of this Environment.


        :param region: The region of this Environment.  # noqa: E501
        :type: str
        """

        self._region = region

    @property
    def gcp_project(self):
        """Gets the gcp_project of this Environment.  # noqa: E501


        :return: The gcp_project of this Environment.  # noqa: E501
        :rtype: str
        """
        return self._gcp_project

    @gcp_project.setter
    def gcp_project(self, gcp_project):
        """Sets the gcp_project of this Environment.


        :param gcp_project: The gcp_project of this Environment.  # noqa: E501
        :type: str
        """

        self._gcp_project = gcp_project

    @property
    def default_resource_request(self):
        """Gets the default_resource_request of this Environment.  # noqa: E501


        :return: The default_resource_request of this Environment.  # noqa: E501
        :rtype: ResourceRequest
        """
        return self._default_resource_request

    @default_resource_request.setter
    def default_resource_request(self, default_resource_request):
        """Sets the default_resource_request of this Environment.


        :param default_resource_request: The default_resource_request of this Environment.  # noqa: E501
        :type: ResourceRequest
        """

        self._default_resource_request = default_resource_request

    @property
    def default_transformer_resource_request(self):
        """Gets the default_transformer_resource_request of this Environment.  # noqa: E501


        :return: The default_transformer_resource_request of this Environment.  # noqa: E501
        :rtype: ResourceRequest
        """
        return self._default_transformer_resource_request

    @default_transformer_resource_request.setter
    def default_transformer_resource_request(self, default_transformer_resource_request):
        """Sets the default_transformer_resource_request of this Environment.


        :param default_transformer_resource_request: The default_transformer_resource_request of this Environment.  # noqa: E501
        :type: ResourceRequest
        """

        self._default_transformer_resource_request = default_transformer_resource_request

    @property
    def default_prediction_job_resource_request(self):
        """Gets the default_prediction_job_resource_request of this Environment.  # noqa: E501


        :return: The default_prediction_job_resource_request of this Environment.  # noqa: E501
        :rtype: PredictionJobResourceRequest
        """
        return self._default_prediction_job_resource_request

    @default_prediction_job_resource_request.setter
    def default_prediction_job_resource_request(self, default_prediction_job_resource_request):
        """Sets the default_prediction_job_resource_request of this Environment.


        :param default_prediction_job_resource_request: The default_prediction_job_resource_request of this Environment.  # noqa: E501
        :type: PredictionJobResourceRequest
        """

        self._default_prediction_job_resource_request = default_prediction_job_resource_request

    @property
    def gpus(self):
        """Gets the gpus of this Environment.  # noqa: E501


        :return: The gpus of this Environment.  # noqa: E501
        :rtype: list[GPU]
        """
        return self._gpus

    @gpus.setter
    def gpus(self, gpus):
        """Sets the gpus of this Environment.


        :param gpus: The gpus of this Environment.  # noqa: E501
        :type: list[GPU]
        """

        self._gpus = gpus

    @property
    def created_at(self):
        """Gets the created_at of this Environment.  # noqa: E501


        :return: The created_at of this Environment.  # noqa: E501
        :rtype: datetime
        """
        return self._created_at

    @created_at.setter
    def created_at(self, created_at):
        """Sets the created_at of this Environment.


        :param created_at: The created_at of this Environment.  # noqa: E501
        :type: datetime
        """

        self._created_at = created_at

    @property
    def updated_at(self):
        """Gets the updated_at of this Environment.  # noqa: E501


        :return: The updated_at of this Environment.  # noqa: E501
        :rtype: datetime
        """
        return self._updated_at

    @updated_at.setter
    def updated_at(self, updated_at):
        """Sets the updated_at of this Environment.


        :param updated_at: The updated_at of this Environment.  # noqa: E501
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
        if issubclass(Environment, dict):
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
        if not isinstance(other, Environment):
            return False

        return self.__dict__ == other.__dict__

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        return not self == other
