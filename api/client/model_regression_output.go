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

// checks if the RegressionOutput type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &RegressionOutput{}

// RegressionOutput struct for RegressionOutput
type RegressionOutput struct {
	PredictionScoreColumn string                     `json:"prediction_score_column"`
	ActualScoreColumn     string                     `json:"actual_score_column"`
	OutputClass           ModelPredictionOutputClass `json:"output_class"`
}

type _RegressionOutput RegressionOutput

// NewRegressionOutput instantiates a new RegressionOutput object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewRegressionOutput(predictionScoreColumn string, actualScoreColumn string, outputClass ModelPredictionOutputClass) *RegressionOutput {
	this := RegressionOutput{}
	this.PredictionScoreColumn = predictionScoreColumn
	this.ActualScoreColumn = actualScoreColumn
	this.OutputClass = outputClass
	return &this
}

// NewRegressionOutputWithDefaults instantiates a new RegressionOutput object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewRegressionOutputWithDefaults() *RegressionOutput {
	this := RegressionOutput{}
	return &this
}

// GetPredictionScoreColumn returns the PredictionScoreColumn field value
func (o *RegressionOutput) GetPredictionScoreColumn() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.PredictionScoreColumn
}

// GetPredictionScoreColumnOk returns a tuple with the PredictionScoreColumn field value
// and a boolean to check if the value has been set.
func (o *RegressionOutput) GetPredictionScoreColumnOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.PredictionScoreColumn, true
}

// SetPredictionScoreColumn sets field value
func (o *RegressionOutput) SetPredictionScoreColumn(v string) {
	o.PredictionScoreColumn = v
}

// GetActualScoreColumn returns the ActualScoreColumn field value
func (o *RegressionOutput) GetActualScoreColumn() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.ActualScoreColumn
}

// GetActualScoreColumnOk returns a tuple with the ActualScoreColumn field value
// and a boolean to check if the value has been set.
func (o *RegressionOutput) GetActualScoreColumnOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ActualScoreColumn, true
}

// SetActualScoreColumn sets field value
func (o *RegressionOutput) SetActualScoreColumn(v string) {
	o.ActualScoreColumn = v
}

// GetOutputClass returns the OutputClass field value
func (o *RegressionOutput) GetOutputClass() ModelPredictionOutputClass {
	if o == nil {
		var ret ModelPredictionOutputClass
		return ret
	}

	return o.OutputClass
}

// GetOutputClassOk returns a tuple with the OutputClass field value
// and a boolean to check if the value has been set.
func (o *RegressionOutput) GetOutputClassOk() (*ModelPredictionOutputClass, bool) {
	if o == nil {
		return nil, false
	}
	return &o.OutputClass, true
}

// SetOutputClass sets field value
func (o *RegressionOutput) SetOutputClass(v ModelPredictionOutputClass) {
	o.OutputClass = v
}

func (o RegressionOutput) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o RegressionOutput) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["prediction_score_column"] = o.PredictionScoreColumn
	toSerialize["actual_score_column"] = o.ActualScoreColumn
	toSerialize["output_class"] = o.OutputClass
	return toSerialize, nil
}

func (o *RegressionOutput) UnmarshalJSON(bytes []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"prediction_score_column",
		"actual_score_column",
		"output_class",
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

	varRegressionOutput := _RegressionOutput{}

	err = json.Unmarshal(bytes, &varRegressionOutput)

	if err != nil {
		return err
	}

	*o = RegressionOutput(varRegressionOutput)

	return err
}

type NullableRegressionOutput struct {
	value *RegressionOutput
	isSet bool
}

func (v NullableRegressionOutput) Get() *RegressionOutput {
	return v.value
}

func (v *NullableRegressionOutput) Set(val *RegressionOutput) {
	v.value = val
	v.isSet = true
}

func (v NullableRegressionOutput) IsSet() bool {
	return v.isSet
}

func (v *NullableRegressionOutput) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableRegressionOutput(val *RegressionOutput) *NullableRegressionOutput {
	return &NullableRegressionOutput{value: val, isSet: true}
}

func (v NullableRegressionOutput) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableRegressionOutput) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
