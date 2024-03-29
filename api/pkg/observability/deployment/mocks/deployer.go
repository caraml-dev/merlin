// Code generated by mockery v2.39.2. DO NOT EDIT.

package mocks

import (
	context "context"

	deployment "github.com/caraml-dev/merlin/pkg/observability/deployment"
	mock "github.com/stretchr/testify/mock"

	models "github.com/caraml-dev/merlin/models"
)

// Deployer is an autogenerated mock type for the Deployer type
type Deployer struct {
	mock.Mock
}

// Deploy provides a mock function with given fields: ctx, data
func (_m *Deployer) Deploy(ctx context.Context, data *models.WorkerData) error {
	ret := _m.Called(ctx, data)

	if len(ret) == 0 {
		panic("no return value specified for Deploy")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.WorkerData) error); ok {
		r0 = rf(ctx, data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetDeployedManifest provides a mock function with given fields: ctx, data
func (_m *Deployer) GetDeployedManifest(ctx context.Context, data *models.WorkerData) (*deployment.Manifest, error) {
	ret := _m.Called(ctx, data)

	if len(ret) == 0 {
		panic("no return value specified for GetDeployedManifest")
	}

	var r0 *deployment.Manifest
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.WorkerData) (*deployment.Manifest, error)); ok {
		return rf(ctx, data)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *models.WorkerData) *deployment.Manifest); ok {
		r0 = rf(ctx, data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*deployment.Manifest)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *models.WorkerData) error); ok {
		r1 = rf(ctx, data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Undeploy provides a mock function with given fields: ctx, data
func (_m *Deployer) Undeploy(ctx context.Context, data *models.WorkerData) error {
	ret := _m.Called(ctx, data)

	if len(ret) == 0 {
		panic("no return value specified for Undeploy")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.WorkerData) error); ok {
		r0 = rf(ctx, data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewDeployer creates a new instance of Deployer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDeployer(t interface {
	mock.TestingT
	Cleanup(func())
}) *Deployer {
	mock := &Deployer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
