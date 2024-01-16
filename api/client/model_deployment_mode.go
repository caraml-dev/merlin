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

// DeploymentMode the model 'DeploymentMode'
type DeploymentMode string

// List of DeploymentMode
const (
	DEPLOYMENTMODE_SERVERLESS     DeploymentMode = "serverless"
	DEPLOYMENTMODE_RAW_DEPLOYMENT DeploymentMode = "raw_deployment"
)

// All allowed values of DeploymentMode enum
var AllowedDeploymentModeEnumValues = []DeploymentMode{
	"serverless",
	"raw_deployment",
}

func (v *DeploymentMode) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := DeploymentMode(value)
	for _, existing := range AllowedDeploymentModeEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid DeploymentMode", value)
}

// NewDeploymentModeFromValue returns a pointer to a valid DeploymentMode
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewDeploymentModeFromValue(v string) (*DeploymentMode, error) {
	ev := DeploymentMode(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for DeploymentMode: valid values are %v", v, AllowedDeploymentModeEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v DeploymentMode) IsValid() bool {
	for _, existing := range AllowedDeploymentModeEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to DeploymentMode value
func (v DeploymentMode) Ptr() *DeploymentMode {
	return &v
}

type NullableDeploymentMode struct {
	value *DeploymentMode
	isSet bool
}

func (v NullableDeploymentMode) Get() *DeploymentMode {
	return v.value
}

func (v *NullableDeploymentMode) Set(val *DeploymentMode) {
	v.value = val
	v.isSet = true
}

func (v NullableDeploymentMode) IsSet() bool {
	return v.isSet
}

func (v *NullableDeploymentMode) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableDeploymentMode(val *DeploymentMode) *NullableDeploymentMode {
	return &NullableDeploymentMode{value: val, isSet: true}
}

func (v NullableDeploymentMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableDeploymentMode) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
