// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

	grpc "google.golang.org/grpc"

	mock "github.com/stretchr/testify/mock"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
)

// UniversalPredictionServiceClient is an autogenerated mock type for the UniversalPredictionServiceClient type
type UniversalPredictionServiceClient struct {
	mock.Mock
}

// PredictValues provides a mock function with given fields: ctx, in, opts
func (_m *UniversalPredictionServiceClient) PredictValues(ctx context.Context, in *upiv1.PredictValuesRequest, opts ...grpc.CallOption) (*upiv1.PredictValuesResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *upiv1.PredictValuesResponse
	if rf, ok := ret.Get(0).(func(context.Context, *upiv1.PredictValuesRequest, ...grpc.CallOption) *upiv1.PredictValuesResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*upiv1.PredictValuesResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *upiv1.PredictValuesRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewUniversalPredictionServiceClient interface {
	mock.TestingT
	Cleanup(func())
}

// NewUniversalPredictionServiceClient creates a new instance of UniversalPredictionServiceClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewUniversalPredictionServiceClient(t mockConstructorTestingTNewUniversalPredictionServiceClient) *UniversalPredictionServiceClient {
	mock := &UniversalPredictionServiceClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
