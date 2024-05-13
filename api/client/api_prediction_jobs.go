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

// PredictionJobsAPIService PredictionJobsAPI service
type PredictionJobsAPIService service

type ApiModelsModelIdVersionsVersionIdJobsGetRequest struct {
	ctx        context.Context
	ApiService *PredictionJobsAPIService
	modelId    int32
	versionId  int32
}

func (r ApiModelsModelIdVersionsVersionIdJobsGetRequest) Execute() ([]PredictionJob, *http.Response, error) {
	return r.ApiService.ModelsModelIdVersionsVersionIdJobsGetExecute(r)
}

/*
ModelsModelIdVersionsVersionIdJobsGet List all prediction jobs of a model version

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param modelId
 @param versionId
 @return ApiModelsModelIdVersionsVersionIdJobsGetRequest
*/
func (a *PredictionJobsAPIService) ModelsModelIdVersionsVersionIdJobsGet(ctx context.Context, modelId int32, versionId int32) ApiModelsModelIdVersionsVersionIdJobsGetRequest {
	return ApiModelsModelIdVersionsVersionIdJobsGetRequest{
		ApiService: a,
		ctx:        ctx,
		modelId:    modelId,
		versionId:  versionId,
	}
}

// Execute executes the request
//  @return []PredictionJob
func (a *PredictionJobsAPIService) ModelsModelIdVersionsVersionIdJobsGetExecute(r ApiModelsModelIdVersionsVersionIdJobsGetRequest) ([]PredictionJob, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodGet
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue []PredictionJob
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "PredictionJobsAPIService.ModelsModelIdVersionsVersionIdJobsGet")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/models/{model_id}/versions/{version_id}/jobs"
	localVarPath = strings.Replace(localVarPath, "{"+"model_id"+"}", url.PathEscape(parameterValueToString(r.modelId, "modelId")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"version_id"+"}", url.PathEscape(parameterValueToString(r.versionId, "versionId")), -1)

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

type ApiModelsModelIdVersionsVersionIdJobsJobIdContainersGetRequest struct {
	ctx        context.Context
	ApiService *PredictionJobsAPIService
	modelId    int32
	versionId  int32
	jobId      string
}

func (r ApiModelsModelIdVersionsVersionIdJobsJobIdContainersGetRequest) Execute() (*Container, *http.Response, error) {
	return r.ApiService.ModelsModelIdVersionsVersionIdJobsJobIdContainersGetExecute(r)
}

/*
ModelsModelIdVersionsVersionIdJobsJobIdContainersGet Get all container belong to a prediction job

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param modelId
 @param versionId
 @param jobId
 @return ApiModelsModelIdVersionsVersionIdJobsJobIdContainersGetRequest
*/
func (a *PredictionJobsAPIService) ModelsModelIdVersionsVersionIdJobsJobIdContainersGet(ctx context.Context, modelId int32, versionId int32, jobId string) ApiModelsModelIdVersionsVersionIdJobsJobIdContainersGetRequest {
	return ApiModelsModelIdVersionsVersionIdJobsJobIdContainersGetRequest{
		ApiService: a,
		ctx:        ctx,
		modelId:    modelId,
		versionId:  versionId,
		jobId:      jobId,
	}
}

// Execute executes the request
//  @return Container
func (a *PredictionJobsAPIService) ModelsModelIdVersionsVersionIdJobsJobIdContainersGetExecute(r ApiModelsModelIdVersionsVersionIdJobsJobIdContainersGetRequest) (*Container, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodGet
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue *Container
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "PredictionJobsAPIService.ModelsModelIdVersionsVersionIdJobsJobIdContainersGet")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/models/{model_id}/versions/{version_id}/jobs/{job_id}/containers"
	localVarPath = strings.Replace(localVarPath, "{"+"model_id"+"}", url.PathEscape(parameterValueToString(r.modelId, "modelId")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"version_id"+"}", url.PathEscape(parameterValueToString(r.versionId, "versionId")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"job_id"+"}", url.PathEscape(parameterValueToString(r.jobId, "jobId")), -1)

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

type ApiModelsModelIdVersionsVersionIdJobsJobIdGetRequest struct {
	ctx        context.Context
	ApiService *PredictionJobsAPIService
	modelId    int32
	versionId  int32
	jobId      int32
}

func (r ApiModelsModelIdVersionsVersionIdJobsJobIdGetRequest) Execute() (*PredictionJob, *http.Response, error) {
	return r.ApiService.ModelsModelIdVersionsVersionIdJobsJobIdGetExecute(r)
}

/*
ModelsModelIdVersionsVersionIdJobsJobIdGet Get prediction jobs with given id

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param modelId
 @param versionId
 @param jobId
 @return ApiModelsModelIdVersionsVersionIdJobsJobIdGetRequest
*/
func (a *PredictionJobsAPIService) ModelsModelIdVersionsVersionIdJobsJobIdGet(ctx context.Context, modelId int32, versionId int32, jobId int32) ApiModelsModelIdVersionsVersionIdJobsJobIdGetRequest {
	return ApiModelsModelIdVersionsVersionIdJobsJobIdGetRequest{
		ApiService: a,
		ctx:        ctx,
		modelId:    modelId,
		versionId:  versionId,
		jobId:      jobId,
	}
}

// Execute executes the request
//  @return PredictionJob
func (a *PredictionJobsAPIService) ModelsModelIdVersionsVersionIdJobsJobIdGetExecute(r ApiModelsModelIdVersionsVersionIdJobsJobIdGetRequest) (*PredictionJob, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodGet
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue *PredictionJob
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "PredictionJobsAPIService.ModelsModelIdVersionsVersionIdJobsJobIdGet")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/models/{model_id}/versions/{version_id}/jobs/{job_id}"
	localVarPath = strings.Replace(localVarPath, "{"+"model_id"+"}", url.PathEscape(parameterValueToString(r.modelId, "modelId")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"version_id"+"}", url.PathEscape(parameterValueToString(r.versionId, "versionId")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"job_id"+"}", url.PathEscape(parameterValueToString(r.jobId, "jobId")), -1)

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

type ApiModelsModelIdVersionsVersionIdJobsJobIdStopPutRequest struct {
	ctx        context.Context
	ApiService *PredictionJobsAPIService
	modelId    int32
	versionId  int32
	jobId      int32
}

func (r ApiModelsModelIdVersionsVersionIdJobsJobIdStopPutRequest) Execute() (*http.Response, error) {
	return r.ApiService.ModelsModelIdVersionsVersionIdJobsJobIdStopPutExecute(r)
}

/*
ModelsModelIdVersionsVersionIdJobsJobIdStopPut Stop prediction jobs with given id

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param modelId
 @param versionId
 @param jobId
 @return ApiModelsModelIdVersionsVersionIdJobsJobIdStopPutRequest
*/
func (a *PredictionJobsAPIService) ModelsModelIdVersionsVersionIdJobsJobIdStopPut(ctx context.Context, modelId int32, versionId int32, jobId int32) ApiModelsModelIdVersionsVersionIdJobsJobIdStopPutRequest {
	return ApiModelsModelIdVersionsVersionIdJobsJobIdStopPutRequest{
		ApiService: a,
		ctx:        ctx,
		modelId:    modelId,
		versionId:  versionId,
		jobId:      jobId,
	}
}

// Execute executes the request
func (a *PredictionJobsAPIService) ModelsModelIdVersionsVersionIdJobsJobIdStopPutExecute(r ApiModelsModelIdVersionsVersionIdJobsJobIdStopPutRequest) (*http.Response, error) {
	var (
		localVarHTTPMethod = http.MethodPut
		localVarPostBody   interface{}
		formFiles          []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "PredictionJobsAPIService.ModelsModelIdVersionsVersionIdJobsJobIdStopPut")
	if err != nil {
		return nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/models/{model_id}/versions/{version_id}/jobs/{job_id}/stop"
	localVarPath = strings.Replace(localVarPath, "{"+"model_id"+"}", url.PathEscape(parameterValueToString(r.modelId, "modelId")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"version_id"+"}", url.PathEscape(parameterValueToString(r.versionId, "versionId")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"job_id"+"}", url.PathEscape(parameterValueToString(r.jobId, "jobId")), -1)

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

type ApiModelsModelIdVersionsVersionIdJobsPagePageGetRequest struct {
	ctx        context.Context
	ApiService *PredictionJobsAPIService
	modelId    int32
	versionId  int32
	projectId  int32
	page       int32
	pageSize   *int32
}

// Number of items on each page. It defaults to 50.
func (r ApiModelsModelIdVersionsVersionIdJobsPagePageGetRequest) PageSize(pageSize int32) ApiModelsModelIdVersionsVersionIdJobsPagePageGetRequest {
	r.pageSize = &pageSize
	return r
}

func (r ApiModelsModelIdVersionsVersionIdJobsPagePageGetRequest) Execute() (*ListJobsPaginatedResponse, *http.Response, error) {
	return r.ApiService.ModelsModelIdVersionsVersionIdJobsPagePageGetExecute(r)
}

/*
ModelsModelIdVersionsVersionIdJobsPagePageGet List all prediction jobs of a model version

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param modelId
 @param versionId
 @param projectId
 @param page
 @return ApiModelsModelIdVersionsVersionIdJobsPagePageGetRequest
*/
func (a *PredictionJobsAPIService) ModelsModelIdVersionsVersionIdJobsPagePageGet(ctx context.Context, modelId int32, versionId int32, projectId int32, page int32) ApiModelsModelIdVersionsVersionIdJobsPagePageGetRequest {
	return ApiModelsModelIdVersionsVersionIdJobsPagePageGetRequest{
		ApiService: a,
		ctx:        ctx,
		modelId:    modelId,
		versionId:  versionId,
		projectId:  projectId,
		page:       page,
	}
}

// Execute executes the request
//  @return ListJobsPaginatedResponse
func (a *PredictionJobsAPIService) ModelsModelIdVersionsVersionIdJobsPagePageGetExecute(r ApiModelsModelIdVersionsVersionIdJobsPagePageGetRequest) (*ListJobsPaginatedResponse, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodGet
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue *ListJobsPaginatedResponse
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "PredictionJobsAPIService.ModelsModelIdVersionsVersionIdJobsPagePageGet")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/models/{model_id}/versions/{version_id}/jobs/page/{page}"
	localVarPath = strings.Replace(localVarPath, "{"+"model_id"+"}", url.PathEscape(parameterValueToString(r.modelId, "modelId")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"version_id"+"}", url.PathEscape(parameterValueToString(r.versionId, "versionId")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"project_id"+"}", url.PathEscape(parameterValueToString(r.projectId, "projectId")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"page"+"}", url.PathEscape(parameterValueToString(r.page, "page")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.pageSize != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "page_size", r.pageSize, "")
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

type ApiModelsModelIdVersionsVersionIdJobsPostRequest struct {
	ctx        context.Context
	ApiService *PredictionJobsAPIService
	modelId    int32
	versionId  int32
	body       *PredictionJob
}

func (r ApiModelsModelIdVersionsVersionIdJobsPostRequest) Body(body PredictionJob) ApiModelsModelIdVersionsVersionIdJobsPostRequest {
	r.body = &body
	return r
}

func (r ApiModelsModelIdVersionsVersionIdJobsPostRequest) Execute() (*PredictionJob, *http.Response, error) {
	return r.ApiService.ModelsModelIdVersionsVersionIdJobsPostExecute(r)
}

/*
ModelsModelIdVersionsVersionIdJobsPost Create a prediction job from the given model version

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param modelId
 @param versionId
 @return ApiModelsModelIdVersionsVersionIdJobsPostRequest
*/
func (a *PredictionJobsAPIService) ModelsModelIdVersionsVersionIdJobsPost(ctx context.Context, modelId int32, versionId int32) ApiModelsModelIdVersionsVersionIdJobsPostRequest {
	return ApiModelsModelIdVersionsVersionIdJobsPostRequest{
		ApiService: a,
		ctx:        ctx,
		modelId:    modelId,
		versionId:  versionId,
	}
}

// Execute executes the request
//  @return PredictionJob
func (a *PredictionJobsAPIService) ModelsModelIdVersionsVersionIdJobsPostExecute(r ApiModelsModelIdVersionsVersionIdJobsPostRequest) (*PredictionJob, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodPost
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue *PredictionJob
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "PredictionJobsAPIService.ModelsModelIdVersionsVersionIdJobsPost")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/models/{model_id}/versions/{version_id}/jobs"
	localVarPath = strings.Replace(localVarPath, "{"+"model_id"+"}", url.PathEscape(parameterValueToString(r.modelId, "modelId")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"version_id"+"}", url.PathEscape(parameterValueToString(r.versionId, "versionId")), -1)

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

type ApiProjectsProjectIdJobsGetRequest struct {
	ctx        context.Context
	ApiService *PredictionJobsAPIService
	projectId  int32
	id         *int32
	name       *string
	modelId    *int32
	versionId  *int32
	status     *string
	error_     *string
}

func (r ApiProjectsProjectIdJobsGetRequest) Id(id int32) ApiProjectsProjectIdJobsGetRequest {
	r.id = &id
	return r
}

func (r ApiProjectsProjectIdJobsGetRequest) Name(name string) ApiProjectsProjectIdJobsGetRequest {
	r.name = &name
	return r
}

func (r ApiProjectsProjectIdJobsGetRequest) ModelId(modelId int32) ApiProjectsProjectIdJobsGetRequest {
	r.modelId = &modelId
	return r
}

func (r ApiProjectsProjectIdJobsGetRequest) VersionId(versionId int32) ApiProjectsProjectIdJobsGetRequest {
	r.versionId = &versionId
	return r
}

func (r ApiProjectsProjectIdJobsGetRequest) Status(status string) ApiProjectsProjectIdJobsGetRequest {
	r.status = &status
	return r
}

func (r ApiProjectsProjectIdJobsGetRequest) Error_(error_ string) ApiProjectsProjectIdJobsGetRequest {
	r.error_ = &error_
	return r
}

func (r ApiProjectsProjectIdJobsGetRequest) Execute() ([]PredictionJob, *http.Response, error) {
	return r.ApiService.ProjectsProjectIdJobsGetExecute(r)
}

/*
ProjectsProjectIdJobsGet List all prediction jobs created using the model

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param projectId
 @return ApiProjectsProjectIdJobsGetRequest
*/
func (a *PredictionJobsAPIService) ProjectsProjectIdJobsGet(ctx context.Context, projectId int32) ApiProjectsProjectIdJobsGetRequest {
	return ApiProjectsProjectIdJobsGetRequest{
		ApiService: a,
		ctx:        ctx,
		projectId:  projectId,
	}
}

// Execute executes the request
//  @return []PredictionJob
func (a *PredictionJobsAPIService) ProjectsProjectIdJobsGetExecute(r ApiProjectsProjectIdJobsGetRequest) ([]PredictionJob, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodGet
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue []PredictionJob
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "PredictionJobsAPIService.ProjectsProjectIdJobsGet")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/projects/{project_id}/jobs"
	localVarPath = strings.Replace(localVarPath, "{"+"project_id"+"}", url.PathEscape(parameterValueToString(r.projectId, "projectId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.id != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "id", r.id, "")
	}
	if r.name != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "name", r.name, "")
	}
	if r.modelId != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "model_id", r.modelId, "")
	}
	if r.versionId != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "version_id", r.versionId, "")
	}
	if r.status != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "status", r.status, "")
	}
	if r.error_ != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "error", r.error_, "")
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

type ApiProjectsProjectIdJobsPagePageGetRequest struct {
	ctx        context.Context
	ApiService *PredictionJobsAPIService
	projectId  int32
	page       int32
	id         *int32
	name       *string
	modelId    *int32
	versionId  *int32
	status     *string
	error_     *string
	pageSize   *int32
}

func (r ApiProjectsProjectIdJobsPagePageGetRequest) Id(id int32) ApiProjectsProjectIdJobsPagePageGetRequest {
	r.id = &id
	return r
}

func (r ApiProjectsProjectIdJobsPagePageGetRequest) Name(name string) ApiProjectsProjectIdJobsPagePageGetRequest {
	r.name = &name
	return r
}

func (r ApiProjectsProjectIdJobsPagePageGetRequest) ModelId(modelId int32) ApiProjectsProjectIdJobsPagePageGetRequest {
	r.modelId = &modelId
	return r
}

func (r ApiProjectsProjectIdJobsPagePageGetRequest) VersionId(versionId int32) ApiProjectsProjectIdJobsPagePageGetRequest {
	r.versionId = &versionId
	return r
}

func (r ApiProjectsProjectIdJobsPagePageGetRequest) Status(status string) ApiProjectsProjectIdJobsPagePageGetRequest {
	r.status = &status
	return r
}

func (r ApiProjectsProjectIdJobsPagePageGetRequest) Error_(error_ string) ApiProjectsProjectIdJobsPagePageGetRequest {
	r.error_ = &error_
	return r
}

// Number of items on each page. It defaults to 50.
func (r ApiProjectsProjectIdJobsPagePageGetRequest) PageSize(pageSize int32) ApiProjectsProjectIdJobsPagePageGetRequest {
	r.pageSize = &pageSize
	return r
}

func (r ApiProjectsProjectIdJobsPagePageGetRequest) Execute() (*ListJobsPaginatedResponse, *http.Response, error) {
	return r.ApiService.ProjectsProjectIdJobsPagePageGetExecute(r)
}

/*
ProjectsProjectIdJobsPagePageGet List all prediction jobs created using the model

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param projectId
 @param page
 @return ApiProjectsProjectIdJobsPagePageGetRequest
*/
func (a *PredictionJobsAPIService) ProjectsProjectIdJobsPagePageGet(ctx context.Context, projectId int32, page int32) ApiProjectsProjectIdJobsPagePageGetRequest {
	return ApiProjectsProjectIdJobsPagePageGetRequest{
		ApiService: a,
		ctx:        ctx,
		projectId:  projectId,
		page:       page,
	}
}

// Execute executes the request
//  @return ListJobsPaginatedResponse
func (a *PredictionJobsAPIService) ProjectsProjectIdJobsPagePageGetExecute(r ApiProjectsProjectIdJobsPagePageGetRequest) (*ListJobsPaginatedResponse, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodGet
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue *ListJobsPaginatedResponse
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "PredictionJobsAPIService.ProjectsProjectIdJobsPagePageGet")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/projects/{project_id}/jobs/page/{page}"
	localVarPath = strings.Replace(localVarPath, "{"+"project_id"+"}", url.PathEscape(parameterValueToString(r.projectId, "projectId")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"page"+"}", url.PathEscape(parameterValueToString(r.page, "page")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.id != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "id", r.id, "")
	}
	if r.name != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "name", r.name, "")
	}
	if r.modelId != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "model_id", r.modelId, "")
	}
	if r.versionId != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "version_id", r.versionId, "")
	}
	if r.status != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "status", r.status, "")
	}
	if r.error_ != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "error", r.error_, "")
	}
	if r.pageSize != nil {
		parameterAddToHeaderOrQuery(localVarQueryParams, "page_size", r.pageSize, "")
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
