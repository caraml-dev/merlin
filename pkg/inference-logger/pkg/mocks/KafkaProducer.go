// Code generated by mockery v2.23.1. DO NOT EDIT.

package mocks

import (
	kafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"

	mock "github.com/stretchr/testify/mock"
)

// KafkaProducer is an autogenerated mock type for the KafkaProducer type
type KafkaProducer struct {
	mock.Mock
}

// GetMetadata provides a mock function with given fields: _a0, _a1, _a2
func (_m *KafkaProducer) GetMetadata(_a0 *string, _a1 bool, _a2 int) (*kafka.Metadata, error) {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 *kafka.Metadata
	var r1 error
	if rf, ok := ret.Get(0).(func(*string, bool, int) (*kafka.Metadata, error)); ok {
		return rf(_a0, _a1, _a2)
	}
	if rf, ok := ret.Get(0).(func(*string, bool, int) *kafka.Metadata); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*kafka.Metadata)
		}
	}

	if rf, ok := ret.Get(1).(func(*string, bool, int) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Produce provides a mock function with given fields: _a0, _a1
func (_m *KafkaProducer) Produce(_a0 *kafka.Message, _a1 chan kafka.Event) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(*kafka.Message, chan kafka.Event) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewKafkaProducer interface {
	mock.TestingT
	Cleanup(func())
}

// NewKafkaProducer creates a new instance of KafkaProducer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewKafkaProducer(t mockConstructorTestingTNewKafkaProducer) *KafkaProducer {
	mock := &KafkaProducer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
