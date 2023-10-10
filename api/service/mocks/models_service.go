// Code generated by mockery v2.20.0. DO NOT EDIT.

package mocks

import (
	context "context"

	models "github.com/caraml-dev/merlin/models"
	mock "github.com/stretchr/testify/mock"
)

// ModelsService is an autogenerated mock type for the ModelsService type
type ModelsService struct {
	mock.Mock
}

// Delete provides a mock function with given fields: model
func (_m *ModelsService) Delete(model *models.Model) error {
	ret := _m.Called(model)

	var r0 error
	if rf, ok := ret.Get(0).(func(*models.Model) error); ok {
		r0 = rf(model)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FindByID provides a mock function with given fields: ctx, modelID
func (_m *ModelsService) FindByID(ctx context.Context, modelID models.ID) (*models.Model, error) {
	ret := _m.Called(ctx, modelID)

	var r0 *models.Model
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, models.ID) (*models.Model, error)); ok {
		return rf(ctx, modelID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, models.ID) *models.Model); ok {
		r0 = rf(ctx, modelID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Model)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, models.ID) error); ok {
		r1 = rf(ctx, modelID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListModels provides a mock function with given fields: ctx, projectID, name
func (_m *ModelsService) ListModels(ctx context.Context, projectID models.ID, name string) ([]*models.Model, error) {
	ret := _m.Called(ctx, projectID, name)

	var r0 []*models.Model
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, models.ID, string) ([]*models.Model, error)); ok {
		return rf(ctx, projectID, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, models.ID, string) []*models.Model); ok {
		r0 = rf(ctx, projectID, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.Model)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, models.ID, string) error); ok {
		r1 = rf(ctx, projectID, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Save provides a mock function with given fields: ctx, model
func (_m *ModelsService) Save(ctx context.Context, model *models.Model) (*models.Model, error) {
	ret := _m.Called(ctx, model)

	var r0 *models.Model
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.Model) (*models.Model, error)); ok {
		return rf(ctx, model)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *models.Model) *models.Model); ok {
		r0 = rf(ctx, model)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Model)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *models.Model) error); ok {
		r1 = rf(ctx, model)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, model
func (_m *ModelsService) Update(ctx context.Context, model *models.Model) (*models.Model, error) {
	ret := _m.Called(ctx, model)

	var r0 *models.Model
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.Model) (*models.Model, error)); ok {
		return rf(ctx, model)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *models.Model) *models.Model); ok {
		r0 = rf(ctx, model)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Model)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *models.Model) error); ok {
		r1 = rf(ctx, model)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewModelsService interface {
	mock.TestingT
	Cleanup(func())
}

// NewModelsService creates a new instance of ModelsService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewModelsService(t mockConstructorTestingTNewModelsService) *ModelsService {
	mock := &ModelsService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
