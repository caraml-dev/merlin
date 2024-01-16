/*
Merlin

API Guide for accessing Merlin's model management, deployment, and serving functionalities

API version: 0.14.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package client

import (
	"encoding/json"
)

// checks if the ModelEndpointRuleDestination type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ModelEndpointRuleDestination{}

// ModelEndpointRuleDestination struct for ModelEndpointRuleDestination
type ModelEndpointRuleDestination struct {
	VersionEndpointId *string          `json:"version_endpoint_id,omitempty"`
	VersionEndpoint   *VersionEndpoint `json:"version_endpoint,omitempty"`
	Weight            *int32           `json:"weight,omitempty"`
}

// NewModelEndpointRuleDestination instantiates a new ModelEndpointRuleDestination object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewModelEndpointRuleDestination() *ModelEndpointRuleDestination {
	this := ModelEndpointRuleDestination{}
	return &this
}

// NewModelEndpointRuleDestinationWithDefaults instantiates a new ModelEndpointRuleDestination object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewModelEndpointRuleDestinationWithDefaults() *ModelEndpointRuleDestination {
	this := ModelEndpointRuleDestination{}
	return &this
}

// GetVersionEndpointId returns the VersionEndpointId field value if set, zero value otherwise.
func (o *ModelEndpointRuleDestination) GetVersionEndpointId() string {
	if o == nil || IsNil(o.VersionEndpointId) {
		var ret string
		return ret
	}
	return *o.VersionEndpointId
}

// GetVersionEndpointIdOk returns a tuple with the VersionEndpointId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ModelEndpointRuleDestination) GetVersionEndpointIdOk() (*string, bool) {
	if o == nil || IsNil(o.VersionEndpointId) {
		return nil, false
	}
	return o.VersionEndpointId, true
}

// HasVersionEndpointId returns a boolean if a field has been set.
func (o *ModelEndpointRuleDestination) HasVersionEndpointId() bool {
	if o != nil && !IsNil(o.VersionEndpointId) {
		return true
	}

	return false
}

// SetVersionEndpointId gets a reference to the given string and assigns it to the VersionEndpointId field.
func (o *ModelEndpointRuleDestination) SetVersionEndpointId(v string) {
	o.VersionEndpointId = &v
}

// GetVersionEndpoint returns the VersionEndpoint field value if set, zero value otherwise.
func (o *ModelEndpointRuleDestination) GetVersionEndpoint() VersionEndpoint {
	if o == nil || IsNil(o.VersionEndpoint) {
		var ret VersionEndpoint
		return ret
	}
	return *o.VersionEndpoint
}

// GetVersionEndpointOk returns a tuple with the VersionEndpoint field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ModelEndpointRuleDestination) GetVersionEndpointOk() (*VersionEndpoint, bool) {
	if o == nil || IsNil(o.VersionEndpoint) {
		return nil, false
	}
	return o.VersionEndpoint, true
}

// HasVersionEndpoint returns a boolean if a field has been set.
func (o *ModelEndpointRuleDestination) HasVersionEndpoint() bool {
	if o != nil && !IsNil(o.VersionEndpoint) {
		return true
	}

	return false
}

// SetVersionEndpoint gets a reference to the given VersionEndpoint and assigns it to the VersionEndpoint field.
func (o *ModelEndpointRuleDestination) SetVersionEndpoint(v VersionEndpoint) {
	o.VersionEndpoint = &v
}

// GetWeight returns the Weight field value if set, zero value otherwise.
func (o *ModelEndpointRuleDestination) GetWeight() int32 {
	if o == nil || IsNil(o.Weight) {
		var ret int32
		return ret
	}
	return *o.Weight
}

// GetWeightOk returns a tuple with the Weight field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ModelEndpointRuleDestination) GetWeightOk() (*int32, bool) {
	if o == nil || IsNil(o.Weight) {
		return nil, false
	}
	return o.Weight, true
}

// HasWeight returns a boolean if a field has been set.
func (o *ModelEndpointRuleDestination) HasWeight() bool {
	if o != nil && !IsNil(o.Weight) {
		return true
	}

	return false
}

// SetWeight gets a reference to the given int32 and assigns it to the Weight field.
func (o *ModelEndpointRuleDestination) SetWeight(v int32) {
	o.Weight = &v
}

func (o ModelEndpointRuleDestination) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ModelEndpointRuleDestination) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.VersionEndpointId) {
		toSerialize["version_endpoint_id"] = o.VersionEndpointId
	}
	if !IsNil(o.VersionEndpoint) {
		toSerialize["version_endpoint"] = o.VersionEndpoint
	}
	if !IsNil(o.Weight) {
		toSerialize["weight"] = o.Weight
	}
	return toSerialize, nil
}

type NullableModelEndpointRuleDestination struct {
	value *ModelEndpointRuleDestination
	isSet bool
}

func (v NullableModelEndpointRuleDestination) Get() *ModelEndpointRuleDestination {
	return v.value
}

func (v *NullableModelEndpointRuleDestination) Set(val *ModelEndpointRuleDestination) {
	v.value = val
	v.isSet = true
}

func (v NullableModelEndpointRuleDestination) IsSet() bool {
	return v.isSet
}

func (v *NullableModelEndpointRuleDestination) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableModelEndpointRuleDestination(val *ModelEndpointRuleDestination) *NullableModelEndpointRuleDestination {
	return &NullableModelEndpointRuleDestination{value: val, isSet: true}
}

func (v NullableModelEndpointRuleDestination) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableModelEndpointRuleDestination) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
