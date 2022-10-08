# coding: utf-8

"""
    Merlin

    API Guide for accessing Merlin's model management, deployment, and serving functionalities  # noqa: E501

    OpenAPI spec version: 0.14.0
    
    Generated by: https://github.com/swagger-api/swagger-codegen.git
"""

from __future__ import absolute_import

import re  # noqa: F401

# python 2 and python 3 compatibility library
import six

from swagger_client.api_client import ApiClient


class ModelsApi(object):
    """NOTE: This class is auto generated by the swagger code generator program.

    Do not edit the class manually.
    Ref: https://github.com/swagger-api/swagger-codegen
    """

    def __init__(self, api_client=None):
        if api_client is None:
            api_client = ApiClient()
        self.api_client = api_client

    def alerts_teams_get(self, **kwargs):  # noqa: E501
        """Lists teams for alert notification channel.  # noqa: E501

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please pass async_req=True
        >>> thread = api.alerts_teams_get(async_req=True)
        >>> result = thread.get()

        :param async_req bool
        :return: list[str]
                 If the method is called asynchronously,
                 returns the request thread.
        """
        kwargs['_return_http_data_only'] = True
        if kwargs.get('async_req'):
            return self.alerts_teams_get_with_http_info(**kwargs)  # noqa: E501
        else:
            (data) = self.alerts_teams_get_with_http_info(**kwargs)  # noqa: E501
            return data

    def alerts_teams_get_with_http_info(self, **kwargs):  # noqa: E501
        """Lists teams for alert notification channel.  # noqa: E501

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please pass async_req=True
        >>> thread = api.alerts_teams_get_with_http_info(async_req=True)
        >>> result = thread.get()

        :param async_req bool
        :return: list[str]
                 If the method is called asynchronously,
                 returns the request thread.
        """

        all_params = []  # noqa: E501
        all_params.append('async_req')
        all_params.append('_return_http_data_only')
        all_params.append('_preload_content')
        all_params.append('_request_timeout')

        params = locals()
        for key, val in six.iteritems(params['kwargs']):
            if key not in all_params:
                raise TypeError(
                    "Got an unexpected keyword argument '%s'"
                    " to method alerts_teams_get" % key
                )
            params[key] = val
        del params['kwargs']

        collection_formats = {}

        path_params = {}

        query_params = []

        header_params = {}

        form_params = []
        local_var_files = {}

        body_params = None
        # HTTP header `Accept`
        header_params['Accept'] = self.api_client.select_header_accept(
            ['*/*'])  # noqa: E501

        # Authentication setting
        auth_settings = ['Bearer']  # noqa: E501

        return self.api_client.call_api(
            '/alerts/teams', 'GET',
            path_params,
            query_params,
            header_params,
            body=body_params,
            post_params=form_params,
            files=local_var_files,
            response_type='list[str]',  # noqa: E501
            auth_settings=auth_settings,
            async_req=params.get('async_req'),
            _return_http_data_only=params.get('_return_http_data_only'),
            _preload_content=params.get('_preload_content', True),
            _request_timeout=params.get('_request_timeout'),
            collection_formats=collection_formats)

    def models_model_id_alerts_get(self, model_id, **kwargs):  # noqa: E501
        """Lists alerts for given model.  # noqa: E501

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please pass async_req=True
        >>> thread = api.models_model_id_alerts_get(model_id, async_req=True)
        >>> result = thread.get()

        :param async_req bool
        :param int model_id: (required)
        :return: list[ModelEndpointAlert]
                 If the method is called asynchronously,
                 returns the request thread.
        """
        kwargs['_return_http_data_only'] = True
        if kwargs.get('async_req'):
            return self.models_model_id_alerts_get_with_http_info(model_id, **kwargs)  # noqa: E501
        else:
            (data) = self.models_model_id_alerts_get_with_http_info(model_id, **kwargs)  # noqa: E501
            return data

    def models_model_id_alerts_get_with_http_info(self, model_id, **kwargs):  # noqa: E501
        """Lists alerts for given model.  # noqa: E501

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please pass async_req=True
        >>> thread = api.models_model_id_alerts_get_with_http_info(model_id, async_req=True)
        >>> result = thread.get()

        :param async_req bool
        :param int model_id: (required)
        :return: list[ModelEndpointAlert]
                 If the method is called asynchronously,
                 returns the request thread.
        """

        all_params = ['model_id']  # noqa: E501
        all_params.append('async_req')
        all_params.append('_return_http_data_only')
        all_params.append('_preload_content')
        all_params.append('_request_timeout')

        params = locals()
        for key, val in six.iteritems(params['kwargs']):
            if key not in all_params:
                raise TypeError(
                    "Got an unexpected keyword argument '%s'"
                    " to method models_model_id_alerts_get" % key
                )
            params[key] = val
        del params['kwargs']
        # verify the required parameter 'model_id' is set
        if ('model_id' not in params or
                params['model_id'] is None):
            raise ValueError("Missing the required parameter `model_id` when calling `models_model_id_alerts_get`")  # noqa: E501

        collection_formats = {}

        path_params = {}
        if 'model_id' in params:
            path_params['model_id'] = params['model_id']  # noqa: E501

        query_params = []

        header_params = {}

        form_params = []
        local_var_files = {}

        body_params = None
        # HTTP header `Accept`
        header_params['Accept'] = self.api_client.select_header_accept(
            ['*/*'])  # noqa: E501

        # Authentication setting
        auth_settings = ['Bearer']  # noqa: E501

        return self.api_client.call_api(
            '/models/{model_id}/alerts', 'GET',
            path_params,
            query_params,
            header_params,
            body=body_params,
            post_params=form_params,
            files=local_var_files,
            response_type='list[ModelEndpointAlert]',  # noqa: E501
            auth_settings=auth_settings,
            async_req=params.get('async_req'),
            _return_http_data_only=params.get('_return_http_data_only'),
            _preload_content=params.get('_preload_content', True),
            _request_timeout=params.get('_request_timeout'),
            collection_formats=collection_formats)

    def models_model_id_endpoints_model_endpoint_id_alert_get(self, model_id, model_endpoint_id, **kwargs):  # noqa: E501
        """Gets alert for given model endpoint.  # noqa: E501

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please pass async_req=True
        >>> thread = api.models_model_id_endpoints_model_endpoint_id_alert_get(model_id, model_endpoint_id, async_req=True)
        >>> result = thread.get()

        :param async_req bool
        :param int model_id: (required)
        :param str model_endpoint_id: (required)
        :return: ModelEndpointAlert
                 If the method is called asynchronously,
                 returns the request thread.
        """
        kwargs['_return_http_data_only'] = True
        if kwargs.get('async_req'):
            return self.models_model_id_endpoints_model_endpoint_id_alert_get_with_http_info(model_id, model_endpoint_id, **kwargs)  # noqa: E501
        else:
            (data) = self.models_model_id_endpoints_model_endpoint_id_alert_get_with_http_info(model_id, model_endpoint_id, **kwargs)  # noqa: E501
            return data

    def models_model_id_endpoints_model_endpoint_id_alert_get_with_http_info(self, model_id, model_endpoint_id, **kwargs):  # noqa: E501
        """Gets alert for given model endpoint.  # noqa: E501

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please pass async_req=True
        >>> thread = api.models_model_id_endpoints_model_endpoint_id_alert_get_with_http_info(model_id, model_endpoint_id, async_req=True)
        >>> result = thread.get()

        :param async_req bool
        :param int model_id: (required)
        :param str model_endpoint_id: (required)
        :return: ModelEndpointAlert
                 If the method is called asynchronously,
                 returns the request thread.
        """

        all_params = ['model_id', 'model_endpoint_id']  # noqa: E501
        all_params.append('async_req')
        all_params.append('_return_http_data_only')
        all_params.append('_preload_content')
        all_params.append('_request_timeout')

        params = locals()
        for key, val in six.iteritems(params['kwargs']):
            if key not in all_params:
                raise TypeError(
                    "Got an unexpected keyword argument '%s'"
                    " to method models_model_id_endpoints_model_endpoint_id_alert_get" % key
                )
            params[key] = val
        del params['kwargs']
        # verify the required parameter 'model_id' is set
        if ('model_id' not in params or
                params['model_id'] is None):
            raise ValueError("Missing the required parameter `model_id` when calling `models_model_id_endpoints_model_endpoint_id_alert_get`")  # noqa: E501
        # verify the required parameter 'model_endpoint_id' is set
        if ('model_endpoint_id' not in params or
                params['model_endpoint_id'] is None):
            raise ValueError("Missing the required parameter `model_endpoint_id` when calling `models_model_id_endpoints_model_endpoint_id_alert_get`")  # noqa: E501

        collection_formats = {}

        path_params = {}
        if 'model_id' in params:
            path_params['model_id'] = params['model_id']  # noqa: E501
        if 'model_endpoint_id' in params:
            path_params['model_endpoint_id'] = params['model_endpoint_id']  # noqa: E501

        query_params = []

        header_params = {}

        form_params = []
        local_var_files = {}

        body_params = None
        # HTTP header `Accept`
        header_params['Accept'] = self.api_client.select_header_accept(
            ['*/*'])  # noqa: E501

        # Authentication setting
        auth_settings = ['Bearer']  # noqa: E501

        return self.api_client.call_api(
            '/models/{model_id}/endpoints/{model_endpoint_id}/alert', 'GET',
            path_params,
            query_params,
            header_params,
            body=body_params,
            post_params=form_params,
            files=local_var_files,
            response_type='ModelEndpointAlert',  # noqa: E501
            auth_settings=auth_settings,
            async_req=params.get('async_req'),
            _return_http_data_only=params.get('_return_http_data_only'),
            _preload_content=params.get('_preload_content', True),
            _request_timeout=params.get('_request_timeout'),
            collection_formats=collection_formats)

    def models_model_id_endpoints_model_endpoint_id_alert_post(self, model_id, model_endpoint_id, **kwargs):  # noqa: E501
        """Creates alert for given model endpoint.  # noqa: E501

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please pass async_req=True
        >>> thread = api.models_model_id_endpoints_model_endpoint_id_alert_post(model_id, model_endpoint_id, async_req=True)
        >>> result = thread.get()

        :param async_req bool
        :param int model_id: (required)
        :param str model_endpoint_id: (required)
        :param ModelEndpointAlert body:
        :return: None
                 If the method is called asynchronously,
                 returns the request thread.
        """
        kwargs['_return_http_data_only'] = True
        if kwargs.get('async_req'):
            return self.models_model_id_endpoints_model_endpoint_id_alert_post_with_http_info(model_id, model_endpoint_id, **kwargs)  # noqa: E501
        else:
            (data) = self.models_model_id_endpoints_model_endpoint_id_alert_post_with_http_info(model_id, model_endpoint_id, **kwargs)  # noqa: E501
            return data

    def models_model_id_endpoints_model_endpoint_id_alert_post_with_http_info(self, model_id, model_endpoint_id, **kwargs):  # noqa: E501
        """Creates alert for given model endpoint.  # noqa: E501

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please pass async_req=True
        >>> thread = api.models_model_id_endpoints_model_endpoint_id_alert_post_with_http_info(model_id, model_endpoint_id, async_req=True)
        >>> result = thread.get()

        :param async_req bool
        :param int model_id: (required)
        :param str model_endpoint_id: (required)
        :param ModelEndpointAlert body:
        :return: None
                 If the method is called asynchronously,
                 returns the request thread.
        """

        all_params = ['model_id', 'model_endpoint_id', 'body']  # noqa: E501
        all_params.append('async_req')
        all_params.append('_return_http_data_only')
        all_params.append('_preload_content')
        all_params.append('_request_timeout')

        params = locals()
        for key, val in six.iteritems(params['kwargs']):
            if key not in all_params:
                raise TypeError(
                    "Got an unexpected keyword argument '%s'"
                    " to method models_model_id_endpoints_model_endpoint_id_alert_post" % key
                )
            params[key] = val
        del params['kwargs']
        # verify the required parameter 'model_id' is set
        if ('model_id' not in params or
                params['model_id'] is None):
            raise ValueError("Missing the required parameter `model_id` when calling `models_model_id_endpoints_model_endpoint_id_alert_post`")  # noqa: E501
        # verify the required parameter 'model_endpoint_id' is set
        if ('model_endpoint_id' not in params or
                params['model_endpoint_id'] is None):
            raise ValueError("Missing the required parameter `model_endpoint_id` when calling `models_model_id_endpoints_model_endpoint_id_alert_post`")  # noqa: E501

        collection_formats = {}

        path_params = {}
        if 'model_id' in params:
            path_params['model_id'] = params['model_id']  # noqa: E501
        if 'model_endpoint_id' in params:
            path_params['model_endpoint_id'] = params['model_endpoint_id']  # noqa: E501

        query_params = []

        header_params = {}

        form_params = []
        local_var_files = {}

        body_params = None
        if 'body' in params:
            body_params = params['body']
        # HTTP header `Content-Type`
        header_params['Content-Type'] = self.api_client.select_header_content_type(  # noqa: E501
            ['*/*'])  # noqa: E501

        # Authentication setting
        auth_settings = ['Bearer']  # noqa: E501

        return self.api_client.call_api(
            '/models/{model_id}/endpoints/{model_endpoint_id}/alert', 'POST',
            path_params,
            query_params,
            header_params,
            body=body_params,
            post_params=form_params,
            files=local_var_files,
            response_type=None,  # noqa: E501
            auth_settings=auth_settings,
            async_req=params.get('async_req'),
            _return_http_data_only=params.get('_return_http_data_only'),
            _preload_content=params.get('_preload_content', True),
            _request_timeout=params.get('_request_timeout'),
            collection_formats=collection_formats)

    def models_model_id_endpoints_model_endpoint_id_alert_put(self, model_id, model_endpoint_id, **kwargs):  # noqa: E501
        """Creates alert for given model endpoint.  # noqa: E501

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please pass async_req=True
        >>> thread = api.models_model_id_endpoints_model_endpoint_id_alert_put(model_id, model_endpoint_id, async_req=True)
        >>> result = thread.get()

        :param async_req bool
        :param int model_id: (required)
        :param str model_endpoint_id: (required)
        :param ModelEndpointAlert body:
        :return: None
                 If the method is called asynchronously,
                 returns the request thread.
        """
        kwargs['_return_http_data_only'] = True
        if kwargs.get('async_req'):
            return self.models_model_id_endpoints_model_endpoint_id_alert_put_with_http_info(model_id, model_endpoint_id, **kwargs)  # noqa: E501
        else:
            (data) = self.models_model_id_endpoints_model_endpoint_id_alert_put_with_http_info(model_id, model_endpoint_id, **kwargs)  # noqa: E501
            return data

    def models_model_id_endpoints_model_endpoint_id_alert_put_with_http_info(self, model_id, model_endpoint_id, **kwargs):  # noqa: E501
        """Creates alert for given model endpoint.  # noqa: E501

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please pass async_req=True
        >>> thread = api.models_model_id_endpoints_model_endpoint_id_alert_put_with_http_info(model_id, model_endpoint_id, async_req=True)
        >>> result = thread.get()

        :param async_req bool
        :param int model_id: (required)
        :param str model_endpoint_id: (required)
        :param ModelEndpointAlert body:
        :return: None
                 If the method is called asynchronously,
                 returns the request thread.
        """

        all_params = ['model_id', 'model_endpoint_id', 'body']  # noqa: E501
        all_params.append('async_req')
        all_params.append('_return_http_data_only')
        all_params.append('_preload_content')
        all_params.append('_request_timeout')

        params = locals()
        for key, val in six.iteritems(params['kwargs']):
            if key not in all_params:
                raise TypeError(
                    "Got an unexpected keyword argument '%s'"
                    " to method models_model_id_endpoints_model_endpoint_id_alert_put" % key
                )
            params[key] = val
        del params['kwargs']
        # verify the required parameter 'model_id' is set
        if ('model_id' not in params or
                params['model_id'] is None):
            raise ValueError("Missing the required parameter `model_id` when calling `models_model_id_endpoints_model_endpoint_id_alert_put`")  # noqa: E501
        # verify the required parameter 'model_endpoint_id' is set
        if ('model_endpoint_id' not in params or
                params['model_endpoint_id'] is None):
            raise ValueError("Missing the required parameter `model_endpoint_id` when calling `models_model_id_endpoints_model_endpoint_id_alert_put`")  # noqa: E501

        collection_formats = {}

        path_params = {}
        if 'model_id' in params:
            path_params['model_id'] = params['model_id']  # noqa: E501
        if 'model_endpoint_id' in params:
            path_params['model_endpoint_id'] = params['model_endpoint_id']  # noqa: E501

        query_params = []

        header_params = {}

        form_params = []
        local_var_files = {}

        body_params = None
        if 'body' in params:
            body_params = params['body']
        # HTTP header `Content-Type`
        header_params['Content-Type'] = self.api_client.select_header_content_type(  # noqa: E501
            ['*/*'])  # noqa: E501

        # Authentication setting
        auth_settings = ['Bearer']  # noqa: E501

        return self.api_client.call_api(
            '/models/{model_id}/endpoints/{model_endpoint_id}/alert', 'PUT',
            path_params,
            query_params,
            header_params,
            body=body_params,
            post_params=form_params,
            files=local_var_files,
            response_type=None,  # noqa: E501
            auth_settings=auth_settings,
            async_req=params.get('async_req'),
            _return_http_data_only=params.get('_return_http_data_only'),
            _preload_content=params.get('_preload_content', True),
            _request_timeout=params.get('_request_timeout'),
            collection_formats=collection_formats)

    def projects_project_id_models_get(self, project_id, **kwargs):  # noqa: E501
        """List existing models  # noqa: E501

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please pass async_req=True
        >>> thread = api.projects_project_id_models_get(project_id, async_req=True)
        >>> result = thread.get()

        :param async_req bool
        :param int project_id: Filter list of models by specific `project_id` (required)
        :param str name: Filter list of models by specific models `name`
        :return: list[Model]
                 If the method is called asynchronously,
                 returns the request thread.
        """
        kwargs['_return_http_data_only'] = True
        if kwargs.get('async_req'):
            return self.projects_project_id_models_get_with_http_info(project_id, **kwargs)  # noqa: E501
        else:
            (data) = self.projects_project_id_models_get_with_http_info(project_id, **kwargs)  # noqa: E501
            return data

    def projects_project_id_models_get_with_http_info(self, project_id, **kwargs):  # noqa: E501
        """List existing models  # noqa: E501

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please pass async_req=True
        >>> thread = api.projects_project_id_models_get_with_http_info(project_id, async_req=True)
        >>> result = thread.get()

        :param async_req bool
        :param int project_id: Filter list of models by specific `project_id` (required)
        :param str name: Filter list of models by specific models `name`
        :return: list[Model]
                 If the method is called asynchronously,
                 returns the request thread.
        """

        all_params = ['project_id', 'name']  # noqa: E501
        all_params.append('async_req')
        all_params.append('_return_http_data_only')
        all_params.append('_preload_content')
        all_params.append('_request_timeout')

        params = locals()
        for key, val in six.iteritems(params['kwargs']):
            if key not in all_params:
                raise TypeError(
                    "Got an unexpected keyword argument '%s'"
                    " to method projects_project_id_models_get" % key
                )
            params[key] = val
        del params['kwargs']
        # verify the required parameter 'project_id' is set
        if ('project_id' not in params or
                params['project_id'] is None):
            raise ValueError("Missing the required parameter `project_id` when calling `projects_project_id_models_get`")  # noqa: E501

        collection_formats = {}

        path_params = {}
        if 'project_id' in params:
            path_params['project_id'] = params['project_id']  # noqa: E501

        query_params = []
        if 'name' in params:
            query_params.append(('name', params['name']))  # noqa: E501

        header_params = {}

        form_params = []
        local_var_files = {}

        body_params = None
        # HTTP header `Accept`
        header_params['Accept'] = self.api_client.select_header_accept(
            ['*/*'])  # noqa: E501

        # Authentication setting
        auth_settings = ['Bearer']  # noqa: E501

        return self.api_client.call_api(
            '/projects/{project_id}/models', 'GET',
            path_params,
            query_params,
            header_params,
            body=body_params,
            post_params=form_params,
            files=local_var_files,
            response_type='list[Model]',  # noqa: E501
            auth_settings=auth_settings,
            async_req=params.get('async_req'),
            _return_http_data_only=params.get('_return_http_data_only'),
            _preload_content=params.get('_preload_content', True),
            _request_timeout=params.get('_request_timeout'),
            collection_formats=collection_formats)

    def projects_project_id_models_model_id_get(self, project_id, model_id, **kwargs):  # noqa: E501
        """Get model  # noqa: E501

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please pass async_req=True
        >>> thread = api.projects_project_id_models_model_id_get(project_id, model_id, async_req=True)
        >>> result = thread.get()

        :param async_req bool
        :param int project_id: project id of the project to be retrieved (required)
        :param int model_id: model id of the model to be retrieved (required)
        :return: list[Model]
                 If the method is called asynchronously,
                 returns the request thread.
        """
        kwargs['_return_http_data_only'] = True
        if kwargs.get('async_req'):
            return self.projects_project_id_models_model_id_get_with_http_info(project_id, model_id, **kwargs)  # noqa: E501
        else:
            (data) = self.projects_project_id_models_model_id_get_with_http_info(project_id, model_id, **kwargs)  # noqa: E501
            return data

    def projects_project_id_models_model_id_get_with_http_info(self, project_id, model_id, **kwargs):  # noqa: E501
        """Get model  # noqa: E501

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please pass async_req=True
        >>> thread = api.projects_project_id_models_model_id_get_with_http_info(project_id, model_id, async_req=True)
        >>> result = thread.get()

        :param async_req bool
        :param int project_id: project id of the project to be retrieved (required)
        :param int model_id: model id of the model to be retrieved (required)
        :return: list[Model]
                 If the method is called asynchronously,
                 returns the request thread.
        """

        all_params = ['project_id', 'model_id']  # noqa: E501
        all_params.append('async_req')
        all_params.append('_return_http_data_only')
        all_params.append('_preload_content')
        all_params.append('_request_timeout')

        params = locals()
        for key, val in six.iteritems(params['kwargs']):
            if key not in all_params:
                raise TypeError(
                    "Got an unexpected keyword argument '%s'"
                    " to method projects_project_id_models_model_id_get" % key
                )
            params[key] = val
        del params['kwargs']
        # verify the required parameter 'project_id' is set
        if ('project_id' not in params or
                params['project_id'] is None):
            raise ValueError("Missing the required parameter `project_id` when calling `projects_project_id_models_model_id_get`")  # noqa: E501
        # verify the required parameter 'model_id' is set
        if ('model_id' not in params or
                params['model_id'] is None):
            raise ValueError("Missing the required parameter `model_id` when calling `projects_project_id_models_model_id_get`")  # noqa: E501

        collection_formats = {}

        path_params = {}
        if 'project_id' in params:
            path_params['project_id'] = params['project_id']  # noqa: E501
        if 'model_id' in params:
            path_params['model_id'] = params['model_id']  # noqa: E501

        query_params = []

        header_params = {}

        form_params = []
        local_var_files = {}

        body_params = None
        # HTTP header `Accept`
        header_params['Accept'] = self.api_client.select_header_accept(
            ['*/*'])  # noqa: E501

        # Authentication setting
        auth_settings = ['Bearer']  # noqa: E501

        return self.api_client.call_api(
            '/projects/{project_id}/models/{model_id}', 'GET',
            path_params,
            query_params,
            header_params,
            body=body_params,
            post_params=form_params,
            files=local_var_files,
            response_type='list[Model]',  # noqa: E501
            auth_settings=auth_settings,
            async_req=params.get('async_req'),
            _return_http_data_only=params.get('_return_http_data_only'),
            _preload_content=params.get('_preload_content', True),
            _request_timeout=params.get('_request_timeout'),
            collection_formats=collection_formats)

    def projects_project_id_models_post(self, project_id, **kwargs):  # noqa: E501
        """Register a new models  # noqa: E501

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please pass async_req=True
        >>> thread = api.projects_project_id_models_post(project_id, async_req=True)
        >>> result = thread.get()

        :param async_req bool
        :param int project_id: Create new model in a specific `project_id` (required)
        :param Model body:
        :return: Model
                 If the method is called asynchronously,
                 returns the request thread.
        """
        kwargs['_return_http_data_only'] = True
        if kwargs.get('async_req'):
            return self.projects_project_id_models_post_with_http_info(project_id, **kwargs)  # noqa: E501
        else:
            (data) = self.projects_project_id_models_post_with_http_info(project_id, **kwargs)  # noqa: E501
            return data

    def projects_project_id_models_post_with_http_info(self, project_id, **kwargs):  # noqa: E501
        """Register a new models  # noqa: E501

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please pass async_req=True
        >>> thread = api.projects_project_id_models_post_with_http_info(project_id, async_req=True)
        >>> result = thread.get()

        :param async_req bool
        :param int project_id: Create new model in a specific `project_id` (required)
        :param Model body:
        :return: Model
                 If the method is called asynchronously,
                 returns the request thread.
        """

        all_params = ['project_id', 'body']  # noqa: E501
        all_params.append('async_req')
        all_params.append('_return_http_data_only')
        all_params.append('_preload_content')
        all_params.append('_request_timeout')

        params = locals()
        for key, val in six.iteritems(params['kwargs']):
            if key not in all_params:
                raise TypeError(
                    "Got an unexpected keyword argument '%s'"
                    " to method projects_project_id_models_post" % key
                )
            params[key] = val
        del params['kwargs']
        # verify the required parameter 'project_id' is set
        if ('project_id' not in params or
                params['project_id'] is None):
            raise ValueError("Missing the required parameter `project_id` when calling `projects_project_id_models_post`")  # noqa: E501

        collection_formats = {}

        path_params = {}
        if 'project_id' in params:
            path_params['project_id'] = params['project_id']  # noqa: E501

        query_params = []

        header_params = {}

        form_params = []
        local_var_files = {}

        body_params = None
        if 'body' in params:
            body_params = params['body']
        # HTTP header `Accept`
        header_params['Accept'] = self.api_client.select_header_accept(
            ['*/*'])  # noqa: E501

        # HTTP header `Content-Type`
        header_params['Content-Type'] = self.api_client.select_header_content_type(  # noqa: E501
            ['*/*'])  # noqa: E501

        # Authentication setting
        auth_settings = ['Bearer']  # noqa: E501

        return self.api_client.call_api(
            '/projects/{project_id}/models', 'POST',
            path_params,
            query_params,
            header_params,
            body=body_params,
            post_params=form_params,
            files=local_var_files,
            response_type='Model',  # noqa: E501
            auth_settings=auth_settings,
            async_req=params.get('async_req'),
            _return_http_data_only=params.get('_return_http_data_only'),
            _preload_content=params.get('_preload_content', True),
            _request_timeout=params.get('_request_timeout'),
            collection_formats=collection_formats)
