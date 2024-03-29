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

// checks if the Logger type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &Logger{}

// Logger struct for Logger
type Logger struct {
	Model       *LoggerConfig           `json:"model,omitempty"`
	Transformer *LoggerConfig           `json:"transformer,omitempty"`
	Prediction  *PredictionLoggerConfig `json:"prediction,omitempty"`
}

// NewLogger instantiates a new Logger object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewLogger() *Logger {
	this := Logger{}
	return &this
}

// NewLoggerWithDefaults instantiates a new Logger object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewLoggerWithDefaults() *Logger {
	this := Logger{}
	return &this
}

// GetModel returns the Model field value if set, zero value otherwise.
func (o *Logger) GetModel() LoggerConfig {
	if o == nil || IsNil(o.Model) {
		var ret LoggerConfig
		return ret
	}
	return *o.Model
}

// GetModelOk returns a tuple with the Model field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Logger) GetModelOk() (*LoggerConfig, bool) {
	if o == nil || IsNil(o.Model) {
		return nil, false
	}
	return o.Model, true
}

// HasModel returns a boolean if a field has been set.
func (o *Logger) HasModel() bool {
	if o != nil && !IsNil(o.Model) {
		return true
	}

	return false
}

// SetModel gets a reference to the given LoggerConfig and assigns it to the Model field.
func (o *Logger) SetModel(v LoggerConfig) {
	o.Model = &v
}

// GetTransformer returns the Transformer field value if set, zero value otherwise.
func (o *Logger) GetTransformer() LoggerConfig {
	if o == nil || IsNil(o.Transformer) {
		var ret LoggerConfig
		return ret
	}
	return *o.Transformer
}

// GetTransformerOk returns a tuple with the Transformer field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Logger) GetTransformerOk() (*LoggerConfig, bool) {
	if o == nil || IsNil(o.Transformer) {
		return nil, false
	}
	return o.Transformer, true
}

// HasTransformer returns a boolean if a field has been set.
func (o *Logger) HasTransformer() bool {
	if o != nil && !IsNil(o.Transformer) {
		return true
	}

	return false
}

// SetTransformer gets a reference to the given LoggerConfig and assigns it to the Transformer field.
func (o *Logger) SetTransformer(v LoggerConfig) {
	o.Transformer = &v
}

// GetPrediction returns the Prediction field value if set, zero value otherwise.
func (o *Logger) GetPrediction() PredictionLoggerConfig {
	if o == nil || IsNil(o.Prediction) {
		var ret PredictionLoggerConfig
		return ret
	}
	return *o.Prediction
}

// GetPredictionOk returns a tuple with the Prediction field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Logger) GetPredictionOk() (*PredictionLoggerConfig, bool) {
	if o == nil || IsNil(o.Prediction) {
		return nil, false
	}
	return o.Prediction, true
}

// HasPrediction returns a boolean if a field has been set.
func (o *Logger) HasPrediction() bool {
	if o != nil && !IsNil(o.Prediction) {
		return true
	}

	return false
}

// SetPrediction gets a reference to the given PredictionLoggerConfig and assigns it to the Prediction field.
func (o *Logger) SetPrediction(v PredictionLoggerConfig) {
	o.Prediction = &v
}

func (o Logger) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o Logger) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Model) {
		toSerialize["model"] = o.Model
	}
	if !IsNil(o.Transformer) {
		toSerialize["transformer"] = o.Transformer
	}
	if !IsNil(o.Prediction) {
		toSerialize["prediction"] = o.Prediction
	}
	return toSerialize, nil
}

type NullableLogger struct {
	value *Logger
	isSet bool
}

func (v NullableLogger) Get() *Logger {
	return v.value
}

func (v *NullableLogger) Set(val *Logger) {
	v.value = val
	v.isSet = true
}

func (v NullableLogger) IsSet() bool {
	return v.isSet
}

func (v *NullableLogger) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableLogger(val *Logger) *NullableLogger {
	return &NullableLogger{value: val, isSet: true}
}

func (v NullableLogger) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableLogger) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
