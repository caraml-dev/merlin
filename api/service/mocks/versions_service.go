// Code generated by mockery v2.0.0-alpha.14. DO NOT EDIT.

package mocks

import (
	context "context"

	config "github.com/gojek/merlin/config"

	mock "github.com/stretchr/testify/mock"

	models "github.com/gojek/merlin/models"
)

// VersionsService is an autogenerated mock type for the VersionsService type
type VersionsService struct {
	mock.Mock
}

// FindByID provides a mock function with given fields: ctx, modelID, versionID, monitoringConfig
func (_m *VersionsService) FindByID(ctx context.Context, modelID models.ID, versionID models.ID, monitoringConfig config.MonitoringConfig) (*models.Version, error) {
	ret := _m.Called(ctx, modelID, versionID, monitoringConfig)

	var r0 *models.Version
	if rf, ok := ret.Get(0).(func(context.Context, models.ID, models.ID, config.MonitoringConfig) *models.Version); ok {
		r0 = rf(ctx, modelID, versionID, monitoringConfig)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Version)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, models.ID, models.ID, config.MonitoringConfig) error); ok {
		r1 = rf(ctx, modelID, versionID, monitoringConfig)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListVersions provides a mock function with given fields: ctx, modelID, monitoringConfig
func (_m *VersionsService) ListVersions(ctx context.Context, modelID models.ID, monitoringConfig config.MonitoringConfig) ([]*models.Version, error) {
	ret := _m.Called(ctx, modelID, monitoringConfig)

	var r0 []*models.Version
	if rf, ok := ret.Get(0).(func(context.Context, models.ID, config.MonitoringConfig) []*models.Version); ok {
		r0 = rf(ctx, modelID, monitoringConfig)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.Version)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, models.ID, config.MonitoringConfig) error); ok {
		r1 = rf(ctx, modelID, monitoringConfig)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Save provides a mock function with given fields: ctx, version, monitoringConfig
func (_m *VersionsService) Save(ctx context.Context, version *models.Version, monitoringConfig config.MonitoringConfig) (*models.Version, error) {
	ret := _m.Called(ctx, version, monitoringConfig)

	var r0 *models.Version
	if rf, ok := ret.Get(0).(func(context.Context, *models.Version, config.MonitoringConfig) *models.Version); ok {
		r0 = rf(ctx, version, monitoringConfig)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Version)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *models.Version, config.MonitoringConfig) error); ok {
		r1 = rf(ctx, version, monitoringConfig)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
