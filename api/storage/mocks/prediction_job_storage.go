// Code generated by mockery v2.43.0. DO NOT EDIT.

package mocks

import (
	models "github.com/caraml-dev/merlin/models"
	mock "github.com/stretchr/testify/mock"
)

// PredictionJobStorage is an autogenerated mock type for the PredictionJobStorage type
type PredictionJobStorage struct {
	mock.Mock
}

// Count provides a mock function with given fields: query
func (_m *PredictionJobStorage) Count(query *models.PredictionJob) int {
	ret := _m.Called(query)

	if len(ret) == 0 {
		panic("no return value specified for Count")
	}

	var r0 int
	if rf, ok := ret.Get(0).(func(*models.PredictionJob) int); ok {
		r0 = rf(query)
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// Delete provides a mock function with given fields: _a0
func (_m *PredictionJobStorage) Delete(_a0 *models.PredictionJob) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*models.PredictionJob) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: ID
func (_m *PredictionJobStorage) Get(ID models.ID) (*models.PredictionJob, error) {
	ret := _m.Called(ID)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *models.PredictionJob
	var r1 error
	if rf, ok := ret.Get(0).(func(models.ID) (*models.PredictionJob, error)); ok {
		return rf(ID)
	}
	if rf, ok := ret.Get(0).(func(models.ID) *models.PredictionJob); ok {
		r0 = rf(ID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.PredictionJob)
		}
	}

	if rf, ok := ret.Get(1).(func(models.ID) error); ok {
		r1 = rf(ID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetFirstSuccessModelVersionPerModel provides a mock function with given fields:
func (_m *PredictionJobStorage) GetFirstSuccessModelVersionPerModel() (map[models.ID]models.ID, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetFirstSuccessModelVersionPerModel")
	}

	var r0 map[models.ID]models.ID
	var r1 error
	if rf, ok := ret.Get(0).(func() (map[models.ID]models.ID, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() map[models.ID]models.ID); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[models.ID]models.ID)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: query, offset, limit
func (_m *PredictionJobStorage) List(query *models.PredictionJob, offset *int, limit *int) ([]*models.PredictionJob, error) {
	ret := _m.Called(query, offset, limit)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []*models.PredictionJob
	var r1 error
	if rf, ok := ret.Get(0).(func(*models.PredictionJob, *int, *int) ([]*models.PredictionJob, error)); ok {
		return rf(query, offset, limit)
	}
	if rf, ok := ret.Get(0).(func(*models.PredictionJob, *int, *int) []*models.PredictionJob); ok {
		r0 = rf(query, offset, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.PredictionJob)
		}
	}

	if rf, ok := ret.Get(1).(func(*models.PredictionJob, *int, *int) error); ok {
		r1 = rf(query, offset, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Save provides a mock function with given fields: predictionJob
func (_m *PredictionJobStorage) Save(predictionJob *models.PredictionJob) error {
	ret := _m.Called(predictionJob)

	if len(ret) == 0 {
		panic("no return value specified for Save")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*models.PredictionJob) error); ok {
		r0 = rf(predictionJob)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewPredictionJobStorage creates a new instance of PredictionJobStorage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPredictionJobStorage(t interface {
	mock.TestingT
	Cleanup(func())
}) *PredictionJobStorage {
	mock := &PredictionJobStorage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
