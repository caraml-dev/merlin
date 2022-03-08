// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	context "context"

	spec "github.com/gojek/merlin-pyspark-app/pkg/spec"
	mock "github.com/stretchr/testify/mock"
)

// ManifestManager is an autogenerated mock type for the ManifestManager type
type ManifestManager struct {
	mock.Mock
}

type ManifestManager_CreateDriverAuthorization struct {
	*mock.Call
}

func (_m ManifestManager_CreateDriverAuthorization) Return(_a0 string, _a1 error) *ManifestManager_CreateDriverAuthorization {
	return &ManifestManager_CreateDriverAuthorization{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ManifestManager) OnCreateDriverAuthorization(ctx context.Context, namespace string) *ManifestManager_CreateDriverAuthorization {
	c := _m.On("CreateDriverAuthorization", ctx, namespace)
	return &ManifestManager_CreateDriverAuthorization{Call: c}
}

func (_m *ManifestManager) OnCreateDriverAuthorizationMatch(matchers ...interface{}) *ManifestManager_CreateDriverAuthorization {
	c := _m.On("CreateDriverAuthorization", matchers...)
	return &ManifestManager_CreateDriverAuthorization{Call: c}
}

// CreateDriverAuthorization provides a mock function with given fields: ctx, namespace
func (_m *ManifestManager) CreateDriverAuthorization(ctx context.Context, namespace string) (string, error) {
	ret := _m.Called(ctx, namespace)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, namespace)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ManifestManager_CreateJobSpec struct {
	*mock.Call
}

func (_m ManifestManager_CreateJobSpec) Return(_a0 string, _a1 error) *ManifestManager_CreateJobSpec {
	return &ManifestManager_CreateJobSpec{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ManifestManager) OnCreateJobSpec(ctx context.Context, predictionJobName string, namespace string, _a3 *spec.PredictionJob) *ManifestManager_CreateJobSpec {
	c := _m.On("CreateJobSpec", ctx, predictionJobName, namespace, _a3)
	return &ManifestManager_CreateJobSpec{Call: c}
}

func (_m *ManifestManager) OnCreateJobSpecMatch(matchers ...interface{}) *ManifestManager_CreateJobSpec {
	c := _m.On("CreateJobSpec", matchers...)
	return &ManifestManager_CreateJobSpec{Call: c}
}

// CreateJobSpec provides a mock function with given fields: ctx, predictionJobName, namespace, _a3
func (_m *ManifestManager) CreateJobSpec(ctx context.Context, predictionJobName string, namespace string, _a3 *spec.PredictionJob) (string, error) {
	ret := _m.Called(ctx, predictionJobName, namespace, _a3)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, string, *spec.PredictionJob) string); ok {
		r0 = rf(ctx, predictionJobName, namespace, _a3)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, *spec.PredictionJob) error); ok {
		r1 = rf(ctx, predictionJobName, namespace, _a3)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ManifestManager_CreateSecret struct {
	*mock.Call
}

func (_m ManifestManager_CreateSecret) Return(_a0 string, _a1 error) *ManifestManager_CreateSecret {
	return &ManifestManager_CreateSecret{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ManifestManager) OnCreateSecret(ctx context.Context, predictionJobName string, namespace string, data string) *ManifestManager_CreateSecret {
	c := _m.On("CreateSecret", ctx, predictionJobName, namespace, data)
	return &ManifestManager_CreateSecret{Call: c}
}

func (_m *ManifestManager) OnCreateSecretMatch(matchers ...interface{}) *ManifestManager_CreateSecret {
	c := _m.On("CreateSecret", matchers...)
	return &ManifestManager_CreateSecret{Call: c}
}

// CreateSecret provides a mock function with given fields: ctx, predictionJobName, namespace, data
func (_m *ManifestManager) CreateSecret(ctx context.Context, predictionJobName string, namespace string, data string) (string, error) {
	ret := _m.Called(ctx, predictionJobName, namespace, data)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) string); ok {
		r0 = rf(ctx, predictionJobName, namespace, data)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, predictionJobName, namespace, data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ManifestManager_DeleteDriverAuthorization struct {
	*mock.Call
}

func (_m ManifestManager_DeleteDriverAuthorization) Return(_a0 error) *ManifestManager_DeleteDriverAuthorization {
	return &ManifestManager_DeleteDriverAuthorization{Call: _m.Call.Return(_a0)}
}

func (_m *ManifestManager) OnDeleteDriverAuthorization(ctx context.Context, namespace string) *ManifestManager_DeleteDriverAuthorization {
	c := _m.On("DeleteDriverAuthorization", ctx, namespace)
	return &ManifestManager_DeleteDriverAuthorization{Call: c}
}

func (_m *ManifestManager) OnDeleteDriverAuthorizationMatch(matchers ...interface{}) *ManifestManager_DeleteDriverAuthorization {
	c := _m.On("DeleteDriverAuthorization", matchers...)
	return &ManifestManager_DeleteDriverAuthorization{Call: c}
}

// DeleteDriverAuthorization provides a mock function with given fields: ctx, namespace
func (_m *ManifestManager) DeleteDriverAuthorization(ctx context.Context, namespace string) error {
	ret := _m.Called(ctx, namespace)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, namespace)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type ManifestManager_DeleteJobSpec struct {
	*mock.Call
}

func (_m ManifestManager_DeleteJobSpec) Return(_a0 error) *ManifestManager_DeleteJobSpec {
	return &ManifestManager_DeleteJobSpec{Call: _m.Call.Return(_a0)}
}

func (_m *ManifestManager) OnDeleteJobSpec(ctx context.Context, predictionJobName string, namespace string) *ManifestManager_DeleteJobSpec {
	c := _m.On("DeleteJobSpec", ctx, predictionJobName, namespace)
	return &ManifestManager_DeleteJobSpec{Call: c}
}

func (_m *ManifestManager) OnDeleteJobSpecMatch(matchers ...interface{}) *ManifestManager_DeleteJobSpec {
	c := _m.On("DeleteJobSpec", matchers...)
	return &ManifestManager_DeleteJobSpec{Call: c}
}

// DeleteJobSpec provides a mock function with given fields: ctx, predictionJobName, namespace
func (_m *ManifestManager) DeleteJobSpec(ctx context.Context, predictionJobName string, namespace string) error {
	ret := _m.Called(ctx, predictionJobName, namespace)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, predictionJobName, namespace)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type ManifestManager_DeleteSecret struct {
	*mock.Call
}

func (_m ManifestManager_DeleteSecret) Return(_a0 error) *ManifestManager_DeleteSecret {
	return &ManifestManager_DeleteSecret{Call: _m.Call.Return(_a0)}
}

func (_m *ManifestManager) OnDeleteSecret(ctx context.Context, predictionJobName string, namespace string) *ManifestManager_DeleteSecret {
	c := _m.On("DeleteSecret", ctx, predictionJobName, namespace)
	return &ManifestManager_DeleteSecret{Call: c}
}

func (_m *ManifestManager) OnDeleteSecretMatch(matchers ...interface{}) *ManifestManager_DeleteSecret {
	c := _m.On("DeleteSecret", matchers...)
	return &ManifestManager_DeleteSecret{Call: c}
}

// DeleteSecret provides a mock function with given fields: ctx, predictionJobName, namespace
func (_m *ManifestManager) DeleteSecret(ctx context.Context, predictionJobName string, namespace string) error {
	ret := _m.Called(ctx, predictionJobName, namespace)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, predictionJobName, namespace)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
