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

// checks if the BinaryClassificationOutput type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &BinaryClassificationOutput{}

// BinaryClassificationOutput struct for BinaryClassificationOutput
type BinaryClassificationOutput struct {
	PredictionScoreColumn string                     `json:"prediction_score_column"`
	ActualLabelColumn     *string                    `json:"actual_label_column,omitempty"`
	PositiveClassLabel    string                     `json:"positive_class_label"`
	NegativeClassLabel    string                     `json:"negative_class_label"`
	ScoreThreshold        *float32                   `json:"score_threshold,omitempty"`
	OutputClass           ModelPredictionOutputClass `json:"output_class"`
}

type _BinaryClassificationOutput BinaryClassificationOutput

// NewBinaryClassificationOutput instantiates a new BinaryClassificationOutput object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewBinaryClassificationOutput(predictionScoreColumn string, positiveClassLabel string, negativeClassLabel string, outputClass ModelPredictionOutputClass) *BinaryClassificationOutput {
	this := BinaryClassificationOutput{}
	this.PredictionScoreColumn = predictionScoreColumn
	this.PositiveClassLabel = positiveClassLabel
	this.NegativeClassLabel = negativeClassLabel
	this.OutputClass = outputClass
	return &this
}

// NewBinaryClassificationOutputWithDefaults instantiates a new BinaryClassificationOutput object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewBinaryClassificationOutputWithDefaults() *BinaryClassificationOutput {
	this := BinaryClassificationOutput{}
	return &this
}

// GetPredictionScoreColumn returns the PredictionScoreColumn field value
func (o *BinaryClassificationOutput) GetPredictionScoreColumn() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.PredictionScoreColumn
}

// GetPredictionScoreColumnOk returns a tuple with the PredictionScoreColumn field value
// and a boolean to check if the value has been set.
func (o *BinaryClassificationOutput) GetPredictionScoreColumnOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.PredictionScoreColumn, true
}

// SetPredictionScoreColumn sets field value
func (o *BinaryClassificationOutput) SetPredictionScoreColumn(v string) {
	o.PredictionScoreColumn = v
}

// GetActualLabelColumn returns the ActualLabelColumn field value if set, zero value otherwise.
func (o *BinaryClassificationOutput) GetActualLabelColumn() string {
	if o == nil || IsNil(o.ActualLabelColumn) {
		var ret string
		return ret
	}
	return *o.ActualLabelColumn
}

// GetActualLabelColumnOk returns a tuple with the ActualLabelColumn field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *BinaryClassificationOutput) GetActualLabelColumnOk() (*string, bool) {
	if o == nil || IsNil(o.ActualLabelColumn) {
		return nil, false
	}
	return o.ActualLabelColumn, true
}

// HasActualLabelColumn returns a boolean if a field has been set.
func (o *BinaryClassificationOutput) HasActualLabelColumn() bool {
	if o != nil && !IsNil(o.ActualLabelColumn) {
		return true
	}

	return false
}

// SetActualLabelColumn gets a reference to the given string and assigns it to the ActualLabelColumn field.
func (o *BinaryClassificationOutput) SetActualLabelColumn(v string) {
	o.ActualLabelColumn = &v
}

// GetPositiveClassLabel returns the PositiveClassLabel field value
func (o *BinaryClassificationOutput) GetPositiveClassLabel() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.PositiveClassLabel
}

// GetPositiveClassLabelOk returns a tuple with the PositiveClassLabel field value
// and a boolean to check if the value has been set.
func (o *BinaryClassificationOutput) GetPositiveClassLabelOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.PositiveClassLabel, true
}

// SetPositiveClassLabel sets field value
func (o *BinaryClassificationOutput) SetPositiveClassLabel(v string) {
	o.PositiveClassLabel = v
}

// GetNegativeClassLabel returns the NegativeClassLabel field value
func (o *BinaryClassificationOutput) GetNegativeClassLabel() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.NegativeClassLabel
}

// GetNegativeClassLabelOk returns a tuple with the NegativeClassLabel field value
// and a boolean to check if the value has been set.
func (o *BinaryClassificationOutput) GetNegativeClassLabelOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.NegativeClassLabel, true
}

// SetNegativeClassLabel sets field value
func (o *BinaryClassificationOutput) SetNegativeClassLabel(v string) {
	o.NegativeClassLabel = v
}

// GetScoreThreshold returns the ScoreThreshold field value if set, zero value otherwise.
func (o *BinaryClassificationOutput) GetScoreThreshold() float32 {
	if o == nil || IsNil(o.ScoreThreshold) {
		var ret float32
		return ret
	}
	return *o.ScoreThreshold
}

// GetScoreThresholdOk returns a tuple with the ScoreThreshold field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *BinaryClassificationOutput) GetScoreThresholdOk() (*float32, bool) {
	if o == nil || IsNil(o.ScoreThreshold) {
		return nil, false
	}
	return o.ScoreThreshold, true
}

// HasScoreThreshold returns a boolean if a field has been set.
func (o *BinaryClassificationOutput) HasScoreThreshold() bool {
	if o != nil && !IsNil(o.ScoreThreshold) {
		return true
	}

	return false
}

// SetScoreThreshold gets a reference to the given float32 and assigns it to the ScoreThreshold field.
func (o *BinaryClassificationOutput) SetScoreThreshold(v float32) {
	o.ScoreThreshold = &v
}

// GetOutputClass returns the OutputClass field value
func (o *BinaryClassificationOutput) GetOutputClass() ModelPredictionOutputClass {
	if o == nil {
		var ret ModelPredictionOutputClass
		return ret
	}

	return o.OutputClass
}

// GetOutputClassOk returns a tuple with the OutputClass field value
// and a boolean to check if the value has been set.
func (o *BinaryClassificationOutput) GetOutputClassOk() (*ModelPredictionOutputClass, bool) {
	if o == nil {
		return nil, false
	}
	return &o.OutputClass, true
}

// SetOutputClass sets field value
func (o *BinaryClassificationOutput) SetOutputClass(v ModelPredictionOutputClass) {
	o.OutputClass = v
}

func (o BinaryClassificationOutput) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o BinaryClassificationOutput) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["prediction_score_column"] = o.PredictionScoreColumn
	if !IsNil(o.ActualLabelColumn) {
		toSerialize["actual_label_column"] = o.ActualLabelColumn
	}
	toSerialize["positive_class_label"] = o.PositiveClassLabel
	toSerialize["negative_class_label"] = o.NegativeClassLabel
	if !IsNil(o.ScoreThreshold) {
		toSerialize["score_threshold"] = o.ScoreThreshold
	}
	toSerialize["output_class"] = o.OutputClass
	return toSerialize, nil
}

func (o *BinaryClassificationOutput) UnmarshalJSON(bytes []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"prediction_score_column",
		"positive_class_label",
		"negative_class_label",
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

	varBinaryClassificationOutput := _BinaryClassificationOutput{}

	err = json.Unmarshal(bytes, &varBinaryClassificationOutput)

	if err != nil {
		return err
	}

	*o = BinaryClassificationOutput(varBinaryClassificationOutput)

	return err
}

type NullableBinaryClassificationOutput struct {
	value *BinaryClassificationOutput
	isSet bool
}

func (v NullableBinaryClassificationOutput) Get() *BinaryClassificationOutput {
	return v.value
}

func (v *NullableBinaryClassificationOutput) Set(val *BinaryClassificationOutput) {
	v.value = val
	v.isSet = true
}

func (v NullableBinaryClassificationOutput) IsSet() bool {
	return v.isSet
}

func (v *NullableBinaryClassificationOutput) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableBinaryClassificationOutput(val *BinaryClassificationOutput) *NullableBinaryClassificationOutput {
	return &NullableBinaryClassificationOutput{value: val, isSet: true}
}

func (v NullableBinaryClassificationOutput) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableBinaryClassificationOutput) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
