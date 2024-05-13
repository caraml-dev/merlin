/*
Merlin

API Guide for accessing Merlin's model management, deployment, and serving functionalities

API version: 0.14.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package client

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ModelsAPIService ModelsAPI service
type ModelsAPIService service

type ApiAlertsTeamsGetRequest struct {
	ctx        context.Context
	ApiService *ModelsAPIService
}

func (r ApiAlertsTeamsGetRequest) Execute() ([]string, *http.Response, error) {
	return r.ApiService.AlertsTeamsGetExecute(r)
}

/*
AlertsTeamsGet Lists teams for alert notification channel.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiAlertsTeamsGetRequest
*/
func (a *ModelsAPIService) AlertsTeamsGet(ctx context.Context) ApiAlertsTeamsGetRequest {
	return ApiAlertsTeamsGetRequest{
		ApiService: a,
		ctx:        ctx,
	}
}

// Execute executes the request
//  @return []string
func (a *ModelsAPIService) AlertsTeamsGetExecute(r ApiAlertsTeamsGetRequest) ([]string, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodGet
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue []string
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ModelsAPIService.AlertsTeamsGet")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/alerts/teams"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"*/*"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["Bearer"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiModelsModelIdAlertsGetRequest struct {
	ctx        context.Context
	ApiService *ModelsAPIService
	modelId    int32
}

func (r ApiModelsModelIdAlertsGetRequest) Execute() ([]ModelEndpointAlert, *http.Response, error) {
	return r.ApiService.ModelsModelIdAlertsGetExecute(r)
}

/*
ModelsModelIdAlertsGet Lists alerts for given model.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param modelId
 @return ApiModelsModelIdAlertsGetRequest
*/
func (a *ModelsAPIService) ModelsModelIdAlertsGet(ctx context.Context, modelId int32) ApiModelsModelIdAlertsGetRequest {
	return ApiModelsModelIdAlertsGetRequest{
		ApiService: a,
		ctx:        ctx,
		modelId:    modelId,
	}
}

// Execute executes the request
//  @return []ModelEndpointAlert
func (a *ModelsAPIService) ModelsModelIdAlertsGetExecute(r ApiModelsModelIdAlertsGetRequest) ([]ModelEndpointAlert, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodGet
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue []ModelEndpointAlert
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ModelsAPIService.ModelsModelIdAlertsGet")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/models/{model_id}/alerts"
	localVarPath = strings.Replace(localVarPath, "{"+"model_id"+"}", url.PathEscape(parameterValueToString(r.modelId, "modelId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"*/*"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["Bearer"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiModelsModelIdEndpointsModelEndpointIdAlertGetRequest struct {
	ctx             context.Context
	ApiService      *ModelsAPIService
	modelId         int32
	modelEndpointId string
}

func (r ApiModelsModelIdEndpointsModelEndpointIdAlertGetRequest) Execute() (*ModelEndpointAlert, *http.Response, error) {
	return r.ApiService.ModelsModelIdEndpointsModelEndpointIdAlertGetExecute(r)
}

/*
ModelsModelIdEndpointsModelEndpointIdAlertGet Gets alert for given model endpoint.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param modelId
 @param modelEndpointId
 @return ApiModelsModelIdEndpointsModelEndpointIdAlertGetRequest
*/
func (a *ModelsAPIService) ModelsModelIdEndpointsModelEndpointIdAlertGet(ctx context.Context, modelId int32, modelEndpointId string) ApiModelsModelIdEndpointsModelEndpointIdAlertGetRequest {
	return ApiModelsModelIdEndpointsModelEndpointIdAlertGetRequest{
		ApiService:      a,
		ctx:             ctx,
		modelId:         modelId,
		modelEndpointId: modelEndpointId,
	}
}

// Execute executes the request
//  @return ModelEndpointAlert
func (a *ModelsAPIService) ModelsModelIdEndpointsModelEndpointIdAlertGetExecute(r ApiModelsModelIdEndpointsModelEndpointIdAlertGetRequest) (*ModelEndpointAlert, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodGet
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue *ModelEndpointAlert
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ModelsAPIService.ModelsModelIdEndpointsModelEndpointIdAlertGet")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/models/{model_id}/endpoints/{model_endpoint_id}/alert"
	localVarPath = strings.Replace(localVarPath, "{"+"model_id"+"}", url.PathEscape(parameterValueToString(r.modelId, "modelId")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"model_endpoint_id"+"}", url.PathEscape(parameterValueToString(r.modelEndpointId, "modelEndpointId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"*/*"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["Bearer"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiModelsModelIdEndpointsModelEndpointIdAlertPostRequest struct {
	ctx             context.Context
	ApiService      *ModelsAPIService
	modelId         int32
	modelEndpointId string
	body            *ModelEndpointAlert
}

func (r ApiModelsModelIdEndpointsModelEndpointIdAlertPostRequest) Body(body ModelEndpointAlert) ApiModelsModelIdEndpointsModelEndpointIdAlertPostRequest {
	r.body = &body
	return r
}

func (r ApiModelsModelIdEndpointsModelEndpointIdAlertPostRequest) Execute() (*http.Response, error) {
	return r.ApiService.ModelsModelIdEndpointsModelEndpointIdAlertPostExecute(r)
}

/*
ModelsModelIdEndpointsModelEndpointIdAlertPost Creates alert for given model endpoint.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param modelId
 @param modelEndpointId
 @return ApiModelsModelIdEndpointsModelEndpointIdAlertPostRequest
*/
func (a *ModelsAPIService) ModelsModelIdEndpointsModelEndpointIdAlertPost(ctx context.Context, modelId int32, modelEndpointId string) ApiModelsModelIdEndpointsModelEndpointIdAlertPostRequest {
	return ApiModelsModelIdEndpointsModelEndpointIdAlertPostRequest{
		ApiService:      a,
		ctx:             ctx,
		modelId:         modelId,
		modelEndpointId: modelEndpointId,
	}
}

// Execute executes the request
func (a *ModelsAPIService) ModelsModelIdEndpointsModelEndpointIdAlertPostExecute(r ApiModelsModelIdEndpointsModelEndpointIdAlertPostRequest) (*http.Response, error) {
	var (
		localVarHTTPMethod = http.MethodPost
		localVarPostBody   interface{}
		formFiles          []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ModelsAPIService.ModelsModelIdEndpointsModelEndpointIdAlertPost")
	if err != nil {
		return nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/models/{model_id}/endpoints/{model_endpoint_id}/alert"
	localVarPath = strings.Replace(localVarPath, "{"+"model_id"+"}", url.PathEscape(parameterValueToString(r.modelId, "modelId")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"model_endpoint_id"+"}", url.PathEscape(parameterValueToString(r.modelEndpointId, "modelEndpointId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.body
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["Bearer"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}

type ApiModelsModelIdEndpointsModelEndpointIdAlertPutRequest struct {
	ctx             context.Context
	ApiService      *ModelsAPIService
	modelId         int32
	modelEndpointId string
	body            *ModelEndpointAlert
}

func (r ApiModelsModelIdEndpointsModelEndpointIdAlertPutRequest) Body(body ModelEndpointAlert) ApiModelsModelIdEndpointsModelEndpointIdAlertPutRequest {
	r.body = &body
	return r
}

func (r ApiModelsModelIdEndpointsModelEndpointIdAlertPutRequest) Execute() (*http.Response, error) {
	return r.ApiService.ModelsModelIdEndpointsModelEndpointIdAlertPutExecute(r)
}

/*
ModelsModelIdEndpointsModelEndpointIdAlertPut Creates alert for given model endpoint.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param modelId
 @param modelEndpointId
 @return ApiModelsModelIdEndpointsModelEndpointIdAlertPutRequest
*/
func (a *ModelsAPIService) ModelsModelIdEndpointsModelEndpointIdAlertPut(ctx context.Context, modelId int32, modelEndpointId string) ApiModelsModelIdEndpointsModelEndpointIdAlertPutRequest {
	return ApiModelsModelIdEndpointsModelEndpointIdAlertPutRequest{
		ApiService:      a,
		ctx:             ctx,
		modelId:         modelId,
		modelEndpointId: modelEndpointId,
	}
}

// Execute executes the request
func (a *ModelsAPIService) ModelsModelIdEndpointsModelEndpointIdAlertPutExecute(r ApiModelsModelIdEndpointsModelEndpointIdAlertPutRequest) (*http.Response, error) {
	var (
		localVarHTTPMethod = http.MethodPut
		localVarPostBody   interface{}
		formFiles          []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ModelsAPIService.ModelsModelIdEndpointsModelEndpointIdAlertPut")
	if err != nil {
		return nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/models/{model_id}/endpoints/{model_endpoint_id}/alert"
	localVarPath = strings.Replace(localVarPath, "{"+"model_id"+"}", url.PathEscape(parameterValueToString(r.modelId, "modelId")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"model_endpoint_id"+"}", url.PathEscape(parameterValueToString(r.modelEndpointId, "modelEndpointId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.body
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["Bearer"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}

type ApiProjectsProjectIdModelsGetRequest struct {
	ctx        context.Context
	ApiService *ModelsAPIService
	projectId  int32
	name       *string
}

// Filter list of models by specific models &#x60;name&#x60;
func (r ApiProjectsProjectIdModelsGetRequest) Name(name string) ApiProjectsProjectIdModelsGetRequest {
	r.name = &name
	return r
}

func (r ApiProjectsProjectIdModelsGetRequest) Execute() ([]Model, *http.Response, error) {
	return r.ApiService.ProjectsProjectIdModelsGetExecute(r)
}

/*
ProjectsProjectIdModelsGet List existing models

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param projectId Filter list of models by specific `project_id`
 @return ApiProjectsProjectIdModelsGetRequest
*/
func (a *ModelsAPIService) ProjectsProjectIdModelsGet(ctx context.Context, projectId int32) ApiProjectsProjectIdModelsGetRequest {
	return ApiProjectsProjectIdModelsGetRequest{
		ApiService: a,
		ctx:        ctx,
		projectId:  projectId,
	}
}

// Execute executes the request
//  @return []Model
func (a *ModelsAPIService) ProjectsProjectIdModelsGetExecute(r ApiProjectsProjectIdModelsGetRequest) ([]Model, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodGet
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue []Model
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ModelsAPIService.ProjectsProjectIdModelsGet")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/projects/{project_id}/models"
	localVarPath = strings.Replace(localVarPath, "{"+"project_id"+"}", url.PathEscape(parameterValueToString(r.projectId, "projectId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.name != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "name", r.name, "")
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"*/*"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["Bearer"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiProjectsProjectIdModelsModelIdDeleteRequest struct {
	ctx        context.Context
	ApiService *ModelsAPIService
	projectId  int32
	modelId    int32
}

func (r ApiProjectsProjectIdModelsModelIdDeleteRequest) Execute() (int32, *http.Response, error) {
	return r.ApiService.ProjectsProjectIdModelsModelIdDeleteExecute(r)
}

/*
ProjectsProjectIdModelsModelIdDelete Delete model

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param projectId project id of the project to be deleted
 @param modelId model id of the model to be deleted
 @return ApiProjectsProjectIdModelsModelIdDeleteRequest
*/
func (a *ModelsAPIService) ProjectsProjectIdModelsModelIdDelete(ctx context.Context, projectId int32, modelId int32) ApiProjectsProjectIdModelsModelIdDeleteRequest {
	return ApiProjectsProjectIdModelsModelIdDeleteRequest{
		ApiService: a,
		ctx:        ctx,
		projectId:  projectId,
		modelId:    modelId,
	}
}

// Execute executes the request
//  @return int32
func (a *ModelsAPIService) ProjectsProjectIdModelsModelIdDeleteExecute(r ApiProjectsProjectIdModelsModelIdDeleteRequest) (int32, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodDelete
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue int32
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ModelsAPIService.ProjectsProjectIdModelsModelIdDelete")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/projects/{project_id}/models/{model_id}"
	localVarPath = strings.Replace(localVarPath, "{"+"project_id"+"}", url.PathEscape(parameterValueToString(r.projectId, "projectId")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"model_id"+"}", url.PathEscape(parameterValueToString(r.modelId, "modelId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"*/*"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["Bearer"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiProjectsProjectIdModelsModelIdGetRequest struct {
	ctx        context.Context
	ApiService *ModelsAPIService
	projectId  int32
	modelId    int32
}

func (r ApiProjectsProjectIdModelsModelIdGetRequest) Execute() ([]Model, *http.Response, error) {
	return r.ApiService.ProjectsProjectIdModelsModelIdGetExecute(r)
}

/*
ProjectsProjectIdModelsModelIdGet Get model

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param projectId project id of the project to be retrieved
 @param modelId model id of the model to be retrieved
 @return ApiProjectsProjectIdModelsModelIdGetRequest
*/
func (a *ModelsAPIService) ProjectsProjectIdModelsModelIdGet(ctx context.Context, projectId int32, modelId int32) ApiProjectsProjectIdModelsModelIdGetRequest {
	return ApiProjectsProjectIdModelsModelIdGetRequest{
		ApiService: a,
		ctx:        ctx,
		projectId:  projectId,
		modelId:    modelId,
	}
}

// Execute executes the request
//  @return []Model
func (a *ModelsAPIService) ProjectsProjectIdModelsModelIdGetExecute(r ApiProjectsProjectIdModelsModelIdGetRequest) ([]Model, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodGet
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue []Model
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ModelsAPIService.ProjectsProjectIdModelsModelIdGet")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/projects/{project_id}/models/{model_id}"
	localVarPath = strings.Replace(localVarPath, "{"+"project_id"+"}", url.PathEscape(parameterValueToString(r.projectId, "projectId")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"model_id"+"}", url.PathEscape(parameterValueToString(r.modelId, "modelId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"*/*"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["Bearer"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiProjectsProjectIdModelsPostRequest struct {
	ctx        context.Context
	ApiService *ModelsAPIService
	projectId  int32
	body       *Model
}

func (r ApiProjectsProjectIdModelsPostRequest) Body(body Model) ApiProjectsProjectIdModelsPostRequest {
	r.body = &body
	return r
}

func (r ApiProjectsProjectIdModelsPostRequest) Execute() (*Model, *http.Response, error) {
	return r.ApiService.ProjectsProjectIdModelsPostExecute(r)
}

/*
ProjectsProjectIdModelsPost Register a new models

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param projectId Create new model in a specific `project_id`
 @return ApiProjectsProjectIdModelsPostRequest
*/
func (a *ModelsAPIService) ProjectsProjectIdModelsPost(ctx context.Context, projectId int32) ApiProjectsProjectIdModelsPostRequest {
	return ApiProjectsProjectIdModelsPostRequest{
		ApiService: a,
		ctx:        ctx,
		projectId:  projectId,
	}
}

// Execute executes the request
//  @return Model
func (a *ModelsAPIService) ProjectsProjectIdModelsPostExecute(r ApiProjectsProjectIdModelsPostRequest) (*Model, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodPost
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue *Model
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "ModelsAPIService.ProjectsProjectIdModelsPost")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/projects/{project_id}/models"
	localVarPath = strings.Replace(localVarPath, "{"+"project_id"+"}", url.PathEscape(parameterValueToString(r.projectId, "projectId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"*/*"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.body
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["Bearer"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}
