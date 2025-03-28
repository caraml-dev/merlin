// Code generated by mockery v2.44.1. DO NOT EDIT.

package mocks

import (
	context "context"

	webhook "github.com/caraml-dev/merlin/webhook"
	mock "github.com/stretchr/testify/mock"

	webhooks "github.com/caraml-dev/mlp/api/pkg/webhooks"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// TriggerWebhooks provides a mock function with given fields: ctx, event, opts
func (_m *Client) TriggerWebhooks(ctx context.Context, event webhooks.EventType, opts ...webhook.Option) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, event)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for TriggerWebhooks")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, webhooks.EventType, ...webhook.Option) error); ok {
		r0 = rf(ctx, event, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewClient creates a new instance of Client. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *Client {
	mock := &Client{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
