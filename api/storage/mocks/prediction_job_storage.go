// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import models "github.com/gojek/merlin/models"

// PredictionJobStorage is an autogenerated mock type for the PredictionJobStorage type
type PredictionJobStorage struct {
	mock.Mock
}

// Get provides a mock function with given fields: ID
func (_m *PredictionJobStorage) Get(ID models.ID) (*models.PredictionJob, error) {
	ret := _m.Called(ID)

	var r0 *models.PredictionJob
	if rf, ok := ret.Get(0).(func(models.ID) *models.PredictionJob); ok {
		r0 = rf(ID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.PredictionJob)
		}
	}

	var r1 error
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

	var r0 map[models.ID]models.ID
	if rf, ok := ret.Get(0).(func() map[models.ID]models.ID); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[models.ID]models.ID)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: query
func (_m *PredictionJobStorage) List(query *models.PredictionJob) ([]*models.PredictionJob, error) {
	ret := _m.Called(query)

	var r0 []*models.PredictionJob
	if rf, ok := ret.Get(0).(func(*models.PredictionJob) []*models.PredictionJob); ok {
		r0 = rf(query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.PredictionJob)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*models.PredictionJob) error); ok {
		r1 = rf(query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Save provides a mock function with given fields: predictionJob
func (_m *PredictionJobStorage) Save(predictionJob *models.PredictionJob) error {
	ret := _m.Called(predictionJob)

	var r0 error
	if rf, ok := ret.Get(0).(func(*models.PredictionJob) error); ok {
		r0 = rf(predictionJob)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
