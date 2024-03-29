/*
Merlin

API Guide for accessing Merlin's model management, deployment, and serving functionalities

API version: 0.14.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package client

import (
	"encoding/json"
	"fmt"
)

// LoggerMode the model 'LoggerMode'
type LoggerMode string

// List of LoggerMode
const (
	LOGGERMODE_ALL      LoggerMode = "all"
	LOGGERMODE_REQUEST  LoggerMode = "request"
	LOGGERMODE_RESPONSE LoggerMode = "response"
)

// All allowed values of LoggerMode enum
var AllowedLoggerModeEnumValues = []LoggerMode{
	"all",
	"request",
	"response",
}

func (v *LoggerMode) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := LoggerMode(value)
	for _, existing := range AllowedLoggerModeEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid LoggerMode", value)
}

// NewLoggerModeFromValue returns a pointer to a valid LoggerMode
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewLoggerModeFromValue(v string) (*LoggerMode, error) {
	ev := LoggerMode(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for LoggerMode: valid values are %v", v, AllowedLoggerModeEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v LoggerMode) IsValid() bool {
	for _, existing := range AllowedLoggerModeEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to LoggerMode value
func (v LoggerMode) Ptr() *LoggerMode {
	return &v
}

type NullableLoggerMode struct {
	value *LoggerMode
	isSet bool
}

func (v NullableLoggerMode) Get() *LoggerMode {
	return v.value
}

func (v *NullableLoggerMode) Set(val *LoggerMode) {
	v.value = val
	v.isSet = true
}

func (v NullableLoggerMode) IsSet() bool {
	return v.isSet
}

func (v *NullableLoggerMode) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableLoggerMode(val *LoggerMode) *NullableLoggerMode {
	return &NullableLoggerMode{value: val, isSet: true}
}

func (v NullableLoggerMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableLoggerMode) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
