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

// checks if the GroundTruthJob type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &GroundTruthJob{}

// GroundTruthJob struct for GroundTruthJob
type GroundTruthJob struct {
	CronSchedule             string  `json:"cron_schedule"`
	CpuRequest               *string `json:"cpu_request,omitempty"`
	CpuLimit                 *string `json:"cpu_limit,omitempty"`
	MemoryRequest            *string `json:"memory_request,omitempty"`
	MemoryLimit              *string `json:"memory_limit,omitempty"`
	StartDayOffsetFromNow    int32   `json:"start_day_offset_from_now"`
	EndDayOffsetFromNow      int32   `json:"end_day_offset_from_now"`
	GracePeriodDay           *int32  `json:"grace_period_day,omitempty"`
	ServiceAccountSecretName string  `json:"service_account_secret_name"`
}

type _GroundTruthJob GroundTruthJob

// NewGroundTruthJob instantiates a new GroundTruthJob object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGroundTruthJob(cronSchedule string, startDayOffsetFromNow int32, endDayOffsetFromNow int32, serviceAccountSecretName string) *GroundTruthJob {
	this := GroundTruthJob{}
	this.CronSchedule = cronSchedule
	this.StartDayOffsetFromNow = startDayOffsetFromNow
	this.EndDayOffsetFromNow = endDayOffsetFromNow
	this.ServiceAccountSecretName = serviceAccountSecretName
	return &this
}

// NewGroundTruthJobWithDefaults instantiates a new GroundTruthJob object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGroundTruthJobWithDefaults() *GroundTruthJob {
	this := GroundTruthJob{}
	return &this
}

// GetCronSchedule returns the CronSchedule field value
func (o *GroundTruthJob) GetCronSchedule() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.CronSchedule
}

// GetCronScheduleOk returns a tuple with the CronSchedule field value
// and a boolean to check if the value has been set.
func (o *GroundTruthJob) GetCronScheduleOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.CronSchedule, true
}

// SetCronSchedule sets field value
func (o *GroundTruthJob) SetCronSchedule(v string) {
	o.CronSchedule = v
}

// GetCpuRequest returns the CpuRequest field value if set, zero value otherwise.
func (o *GroundTruthJob) GetCpuRequest() string {
	if o == nil || IsNil(o.CpuRequest) {
		var ret string
		return ret
	}
	return *o.CpuRequest
}

// GetCpuRequestOk returns a tuple with the CpuRequest field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GroundTruthJob) GetCpuRequestOk() (*string, bool) {
	if o == nil || IsNil(o.CpuRequest) {
		return nil, false
	}
	return o.CpuRequest, true
}

// HasCpuRequest returns a boolean if a field has been set.
func (o *GroundTruthJob) HasCpuRequest() bool {
	if o != nil && !IsNil(o.CpuRequest) {
		return true
	}

	return false
}

// SetCpuRequest gets a reference to the given string and assigns it to the CpuRequest field.
func (o *GroundTruthJob) SetCpuRequest(v string) {
	o.CpuRequest = &v
}

// GetCpuLimit returns the CpuLimit field value if set, zero value otherwise.
func (o *GroundTruthJob) GetCpuLimit() string {
	if o == nil || IsNil(o.CpuLimit) {
		var ret string
		return ret
	}
	return *o.CpuLimit
}

// GetCpuLimitOk returns a tuple with the CpuLimit field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GroundTruthJob) GetCpuLimitOk() (*string, bool) {
	if o == nil || IsNil(o.CpuLimit) {
		return nil, false
	}
	return o.CpuLimit, true
}

// HasCpuLimit returns a boolean if a field has been set.
func (o *GroundTruthJob) HasCpuLimit() bool {
	if o != nil && !IsNil(o.CpuLimit) {
		return true
	}

	return false
}

// SetCpuLimit gets a reference to the given string and assigns it to the CpuLimit field.
func (o *GroundTruthJob) SetCpuLimit(v string) {
	o.CpuLimit = &v
}

// GetMemoryRequest returns the MemoryRequest field value if set, zero value otherwise.
func (o *GroundTruthJob) GetMemoryRequest() string {
	if o == nil || IsNil(o.MemoryRequest) {
		var ret string
		return ret
	}
	return *o.MemoryRequest
}

// GetMemoryRequestOk returns a tuple with the MemoryRequest field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GroundTruthJob) GetMemoryRequestOk() (*string, bool) {
	if o == nil || IsNil(o.MemoryRequest) {
		return nil, false
	}
	return o.MemoryRequest, true
}

// HasMemoryRequest returns a boolean if a field has been set.
func (o *GroundTruthJob) HasMemoryRequest() bool {
	if o != nil && !IsNil(o.MemoryRequest) {
		return true
	}

	return false
}

// SetMemoryRequest gets a reference to the given string and assigns it to the MemoryRequest field.
func (o *GroundTruthJob) SetMemoryRequest(v string) {
	o.MemoryRequest = &v
}

// GetMemoryLimit returns the MemoryLimit field value if set, zero value otherwise.
func (o *GroundTruthJob) GetMemoryLimit() string {
	if o == nil || IsNil(o.MemoryLimit) {
		var ret string
		return ret
	}
	return *o.MemoryLimit
}

// GetMemoryLimitOk returns a tuple with the MemoryLimit field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GroundTruthJob) GetMemoryLimitOk() (*string, bool) {
	if o == nil || IsNil(o.MemoryLimit) {
		return nil, false
	}
	return o.MemoryLimit, true
}

// HasMemoryLimit returns a boolean if a field has been set.
func (o *GroundTruthJob) HasMemoryLimit() bool {
	if o != nil && !IsNil(o.MemoryLimit) {
		return true
	}

	return false
}

// SetMemoryLimit gets a reference to the given string and assigns it to the MemoryLimit field.
func (o *GroundTruthJob) SetMemoryLimit(v string) {
	o.MemoryLimit = &v
}

// GetStartDayOffsetFromNow returns the StartDayOffsetFromNow field value
func (o *GroundTruthJob) GetStartDayOffsetFromNow() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.StartDayOffsetFromNow
}

// GetStartDayOffsetFromNowOk returns a tuple with the StartDayOffsetFromNow field value
// and a boolean to check if the value has been set.
func (o *GroundTruthJob) GetStartDayOffsetFromNowOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.StartDayOffsetFromNow, true
}

// SetStartDayOffsetFromNow sets field value
func (o *GroundTruthJob) SetStartDayOffsetFromNow(v int32) {
	o.StartDayOffsetFromNow = v
}

// GetEndDayOffsetFromNow returns the EndDayOffsetFromNow field value
func (o *GroundTruthJob) GetEndDayOffsetFromNow() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.EndDayOffsetFromNow
}

// GetEndDayOffsetFromNowOk returns a tuple with the EndDayOffsetFromNow field value
// and a boolean to check if the value has been set.
func (o *GroundTruthJob) GetEndDayOffsetFromNowOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.EndDayOffsetFromNow, true
}

// SetEndDayOffsetFromNow sets field value
func (o *GroundTruthJob) SetEndDayOffsetFromNow(v int32) {
	o.EndDayOffsetFromNow = v
}

// GetGracePeriodDay returns the GracePeriodDay field value if set, zero value otherwise.
func (o *GroundTruthJob) GetGracePeriodDay() int32 {
	if o == nil || IsNil(o.GracePeriodDay) {
		var ret int32
		return ret
	}
	return *o.GracePeriodDay
}

// GetGracePeriodDayOk returns a tuple with the GracePeriodDay field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GroundTruthJob) GetGracePeriodDayOk() (*int32, bool) {
	if o == nil || IsNil(o.GracePeriodDay) {
		return nil, false
	}
	return o.GracePeriodDay, true
}

// HasGracePeriodDay returns a boolean if a field has been set.
func (o *GroundTruthJob) HasGracePeriodDay() bool {
	if o != nil && !IsNil(o.GracePeriodDay) {
		return true
	}

	return false
}

// SetGracePeriodDay gets a reference to the given int32 and assigns it to the GracePeriodDay field.
func (o *GroundTruthJob) SetGracePeriodDay(v int32) {
	o.GracePeriodDay = &v
}

// GetServiceAccountSecretName returns the ServiceAccountSecretName field value
func (o *GroundTruthJob) GetServiceAccountSecretName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.ServiceAccountSecretName
}

// GetServiceAccountSecretNameOk returns a tuple with the ServiceAccountSecretName field value
// and a boolean to check if the value has been set.
func (o *GroundTruthJob) GetServiceAccountSecretNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ServiceAccountSecretName, true
}

// SetServiceAccountSecretName sets field value
func (o *GroundTruthJob) SetServiceAccountSecretName(v string) {
	o.ServiceAccountSecretName = v
}

func (o GroundTruthJob) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o GroundTruthJob) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["cron_schedule"] = o.CronSchedule
	if !IsNil(o.CpuRequest) {
		toSerialize["cpu_request"] = o.CpuRequest
	}
	if !IsNil(o.CpuLimit) {
		toSerialize["cpu_limit"] = o.CpuLimit
	}
	if !IsNil(o.MemoryRequest) {
		toSerialize["memory_request"] = o.MemoryRequest
	}
	if !IsNil(o.MemoryLimit) {
		toSerialize["memory_limit"] = o.MemoryLimit
	}
	toSerialize["start_day_offset_from_now"] = o.StartDayOffsetFromNow
	toSerialize["end_day_offset_from_now"] = o.EndDayOffsetFromNow
	if !IsNil(o.GracePeriodDay) {
		toSerialize["grace_period_day"] = o.GracePeriodDay
	}
	toSerialize["service_account_secret_name"] = o.ServiceAccountSecretName
	return toSerialize, nil
}

func (o *GroundTruthJob) UnmarshalJSON(bytes []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"cron_schedule",
		"start_day_offset_from_now",
		"end_day_offset_from_now",
		"service_account_secret_name",
	}

	allProperties := make(map[string]interface{})

	err = json.Unmarshal(bytes, &allProperties)

	if err != nil {
		return err
	}

	for _, requiredProperty := range requiredProperties {
		if _, exists := allProperties[requiredProperty]; !exists {
			return fmt.Errorf("no value given for required property %v", requiredProperty)
		}
	}

	varGroundTruthJob := _GroundTruthJob{}

	err = json.Unmarshal(bytes, &varGroundTruthJob)

	if err != nil {
		return err
	}

	*o = GroundTruthJob(varGroundTruthJob)

	return err
}

type NullableGroundTruthJob struct {
	value *GroundTruthJob
	isSet bool
}

func (v NullableGroundTruthJob) Get() *GroundTruthJob {
	return v.value
}

func (v *NullableGroundTruthJob) Set(val *GroundTruthJob) {
	v.value = val
	v.isSet = true
}

func (v NullableGroundTruthJob) IsSet() bool {
	return v.isSet
}

func (v *NullableGroundTruthJob) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGroundTruthJob(val *GroundTruthJob) *NullableGroundTruthJob {
	return &NullableGroundTruthJob{value: val, isSet: true}
}

func (v NullableGroundTruthJob) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGroundTruthJob) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
