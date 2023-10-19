// Code generated by mockery v2.20.0. DO NOT EDIT.

package mocks

import (
	models "github.com/caraml-dev/merlin/models"
	mock "github.com/stretchr/testify/mock"
)

// ModelEndpointAlertService is an autogenerated mock type for the ModelEndpointAlertService type
type ModelEndpointAlertService struct {
	mock.Mock
}

// CreateModelEndpointAlert provides a mock function with given fields: user, alert
func (_m *ModelEndpointAlertService) CreateModelEndpointAlert(user string, alert *models.ModelEndpointAlert) (*models.ModelEndpointAlert, error) {
	ret := _m.Called(user, alert)

	var r0 *models.ModelEndpointAlert
	var r1 error
	if rf, ok := ret.Get(0).(func(string, *models.ModelEndpointAlert) (*models.ModelEndpointAlert, error)); ok {
		return rf(user, alert)
	}
	if rf, ok := ret.Get(0).(func(string, *models.ModelEndpointAlert) *models.ModelEndpointAlert); ok {
		r0 = rf(user, alert)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.ModelEndpointAlert)
		}
	}

	if rf, ok := ret.Get(1).(func(string, *models.ModelEndpointAlert) error); ok {
		r1 = rf(user, alert)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetModelEndpointAlert provides a mock function with given fields: modelID, modelEndpointID
func (_m *ModelEndpointAlertService) GetModelEndpointAlert(modelID models.ID, modelEndpointID models.ID) (*models.ModelEndpointAlert, error) {
	ret := _m.Called(modelID, modelEndpointID)

	var r0 *models.ModelEndpointAlert
	var r1 error
	if rf, ok := ret.Get(0).(func(models.ID, models.ID) (*models.ModelEndpointAlert, error)); ok {
		return rf(modelID, modelEndpointID)
	}
	if rf, ok := ret.Get(0).(func(models.ID, models.ID) *models.ModelEndpointAlert); ok {
		r0 = rf(modelID, modelEndpointID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.ModelEndpointAlert)
		}
	}

	if rf, ok := ret.Get(1).(func(models.ID, models.ID) error); ok {
		r1 = rf(modelID, modelEndpointID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListModelAlerts provides a mock function with given fields: modelID
func (_m *ModelEndpointAlertService) ListModelAlerts(modelID models.ID) ([]*models.ModelEndpointAlert, error) {
	ret := _m.Called(modelID)

	var r0 []*models.ModelEndpointAlert
	var r1 error
	if rf, ok := ret.Get(0).(func(models.ID) ([]*models.ModelEndpointAlert, error)); ok {
		return rf(modelID)
	}
	if rf, ok := ret.Get(0).(func(models.ID) []*models.ModelEndpointAlert); ok {
		r0 = rf(modelID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.ModelEndpointAlert)
		}
	}

	if rf, ok := ret.Get(1).(func(models.ID) error); ok {
		r1 = rf(modelID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListTeams provides a mock function with given fields:
func (_m *ModelEndpointAlertService) ListTeams() ([]string, error) {
	ret := _m.Called()

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]string, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateModelEndpointAlert provides a mock function with given fields: user, alert
func (_m *ModelEndpointAlertService) UpdateModelEndpointAlert(user string, alert *models.ModelEndpointAlert) (*models.ModelEndpointAlert, error) {
	ret := _m.Called(user, alert)

	var r0 *models.ModelEndpointAlert
	var r1 error
	if rf, ok := ret.Get(0).(func(string, *models.ModelEndpointAlert) (*models.ModelEndpointAlert, error)); ok {
		return rf(user, alert)
	}
	if rf, ok := ret.Get(0).(func(string, *models.ModelEndpointAlert) *models.ModelEndpointAlert); ok {
		r0 = rf(user, alert)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.ModelEndpointAlert)
		}
	}

	if rf, ok := ret.Get(1).(func(string, *models.ModelEndpointAlert) error); ok {
		r1 = rf(user, alert)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewModelEndpointAlertService interface {
	mock.TestingT
	Cleanup(func())
}

// NewModelEndpointAlertService creates a new instance of ModelEndpointAlertService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewModelEndpointAlertService(t mockConstructorTestingTNewModelEndpointAlertService) *ModelEndpointAlertService {
	mock := &ModelEndpointAlertService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
