/*
Merlin

API Guide for accessing Merlin's model management, deployment, and serving functionalities

API version: 0.14.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package client

import (
	"encoding/json"
	"time"
)

// checks if the Version type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &Version{}

// Version struct for Version
type Version struct {
	Id              *int32                 `json:"id,omitempty"`
	ModelId         *int32                 `json:"model_id,omitempty"`
	MlflowRunId     *string                `json:"mlflow_run_id,omitempty"`
	MlflowUrl       *string                `json:"mlflow_url,omitempty"`
	ArtifactUri     *string                `json:"artifact_uri,omitempty"`
	Endpoints       []VersionEndpoint      `json:"endpoints,omitempty"`
	Properties      map[string]interface{} `json:"properties,omitempty"`
	Labels          *map[string]string     `json:"labels,omitempty"`
	CustomPredictor *CustomPredictor       `json:"custom_predictor,omitempty"`
	CreatedAt       *time.Time             `json:"created_at,omitempty"`
	UpdatedAt       *time.Time             `json:"updated_at,omitempty"`
	PythonVersion   *string                `json:"python_version,omitempty"`
	ModelSchema     *ModelSchema           `json:"model_schema,omitempty"`
}

// NewVersion instantiates a new Version object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewVersion() *Version {
	this := Version{}
	return &this
}

// NewVersionWithDefaults instantiates a new Version object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewVersionWithDefaults() *Version {
	this := Version{}
	return &this
}

// GetId returns the Id field value if set, zero value otherwise.
func (o *Version) GetId() int32 {
	if o == nil || IsNil(o.Id) {
		var ret int32
		return ret
	}
	return *o.Id
}

// GetIdOk returns a tuple with the Id field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Version) GetIdOk() (*int32, bool) {
	if o == nil || IsNil(o.Id) {
		return nil, false
	}
	return o.Id, true
}

// HasId returns a boolean if a field has been set.
func (o *Version) HasId() bool {
	if o != nil && !IsNil(o.Id) {
		return true
	}

	return false
}

// SetId gets a reference to the given int32 and assigns it to the Id field.
func (o *Version) SetId(v int32) {
	o.Id = &v
}

// GetModelId returns the ModelId field value if set, zero value otherwise.
func (o *Version) GetModelId() int32 {
	if o == nil || IsNil(o.ModelId) {
		var ret int32
		return ret
	}
	return *o.ModelId
}

// GetModelIdOk returns a tuple with the ModelId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Version) GetModelIdOk() (*int32, bool) {
	if o == nil || IsNil(o.ModelId) {
		return nil, false
	}
	return o.ModelId, true
}

// HasModelId returns a boolean if a field has been set.
func (o *Version) HasModelId() bool {
	if o != nil && !IsNil(o.ModelId) {
		return true
	}

	return false
}

// SetModelId gets a reference to the given int32 and assigns it to the ModelId field.
func (o *Version) SetModelId(v int32) {
	o.ModelId = &v
}

// GetMlflowRunId returns the MlflowRunId field value if set, zero value otherwise.
func (o *Version) GetMlflowRunId() string {
	if o == nil || IsNil(o.MlflowRunId) {
		var ret string
		return ret
	}
	return *o.MlflowRunId
}

// GetMlflowRunIdOk returns a tuple with the MlflowRunId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Version) GetMlflowRunIdOk() (*string, bool) {
	if o == nil || IsNil(o.MlflowRunId) {
		return nil, false
	}
	return o.MlflowRunId, true
}

// HasMlflowRunId returns a boolean if a field has been set.
func (o *Version) HasMlflowRunId() bool {
	if o != nil && !IsNil(o.MlflowRunId) {
		return true
	}

	return false
}

// SetMlflowRunId gets a reference to the given string and assigns it to the MlflowRunId field.
func (o *Version) SetMlflowRunId(v string) {
	o.MlflowRunId = &v
}

// GetMlflowUrl returns the MlflowUrl field value if set, zero value otherwise.
func (o *Version) GetMlflowUrl() string {
	if o == nil || IsNil(o.MlflowUrl) {
		var ret string
		return ret
	}
	return *o.MlflowUrl
}

// GetMlflowUrlOk returns a tuple with the MlflowUrl field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Version) GetMlflowUrlOk() (*string, bool) {
	if o == nil || IsNil(o.MlflowUrl) {
		return nil, false
	}
	return o.MlflowUrl, true
}

// HasMlflowUrl returns a boolean if a field has been set.
func (o *Version) HasMlflowUrl() bool {
	if o != nil && !IsNil(o.MlflowUrl) {
		return true
	}

	return false
}

// SetMlflowUrl gets a reference to the given string and assigns it to the MlflowUrl field.
func (o *Version) SetMlflowUrl(v string) {
	o.MlflowUrl = &v
}

// GetArtifactUri returns the ArtifactUri field value if set, zero value otherwise.
func (o *Version) GetArtifactUri() string {
	if o == nil || IsNil(o.ArtifactUri) {
		var ret string
		return ret
	}
	return *o.ArtifactUri
}

// GetArtifactUriOk returns a tuple with the ArtifactUri field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Version) GetArtifactUriOk() (*string, bool) {
	if o == nil || IsNil(o.ArtifactUri) {
		return nil, false
	}
	return o.ArtifactUri, true
}

// HasArtifactUri returns a boolean if a field has been set.
func (o *Version) HasArtifactUri() bool {
	if o != nil && !IsNil(o.ArtifactUri) {
		return true
	}

	return false
}

// SetArtifactUri gets a reference to the given string and assigns it to the ArtifactUri field.
func (o *Version) SetArtifactUri(v string) {
	o.ArtifactUri = &v
}

// GetEndpoints returns the Endpoints field value if set, zero value otherwise.
func (o *Version) GetEndpoints() []VersionEndpoint {
	if o == nil || IsNil(o.Endpoints) {
		var ret []VersionEndpoint
		return ret
	}
	return o.Endpoints
}

// GetEndpointsOk returns a tuple with the Endpoints field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Version) GetEndpointsOk() ([]VersionEndpoint, bool) {
	if o == nil || IsNil(o.Endpoints) {
		return nil, false
	}
	return o.Endpoints, true
}

// HasEndpoints returns a boolean if a field has been set.
func (o *Version) HasEndpoints() bool {
	if o != nil && !IsNil(o.Endpoints) {
		return true
	}

	return false
}

// SetEndpoints gets a reference to the given []VersionEndpoint and assigns it to the Endpoints field.
func (o *Version) SetEndpoints(v []VersionEndpoint) {
	o.Endpoints = v
}

// GetProperties returns the Properties field value if set, zero value otherwise.
func (o *Version) GetProperties() map[string]interface{} {
	if o == nil || IsNil(o.Properties) {
		var ret map[string]interface{}
		return ret
	}
	return o.Properties
}

// GetPropertiesOk returns a tuple with the Properties field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Version) GetPropertiesOk() (map[string]interface{}, bool) {
	if o == nil || IsNil(o.Properties) {
		return map[string]interface{}{}, false
	}
	return o.Properties, true
}

// HasProperties returns a boolean if a field has been set.
func (o *Version) HasProperties() bool {
	if o != nil && !IsNil(o.Properties) {
		return true
	}

	return false
}

// SetProperties gets a reference to the given map[string]interface{} and assigns it to the Properties field.
func (o *Version) SetProperties(v map[string]interface{}) {
	o.Properties = v
}

// GetLabels returns the Labels field value if set, zero value otherwise.
func (o *Version) GetLabels() map[string]string {
	if o == nil || IsNil(o.Labels) {
		var ret map[string]string
		return ret
	}
	return *o.Labels
}

// GetLabelsOk returns a tuple with the Labels field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Version) GetLabelsOk() (*map[string]string, bool) {
	if o == nil || IsNil(o.Labels) {
		return nil, false
	}
	return o.Labels, true
}

// HasLabels returns a boolean if a field has been set.
func (o *Version) HasLabels() bool {
	if o != nil && !IsNil(o.Labels) {
		return true
	}

	return false
}

// SetLabels gets a reference to the given map[string]string and assigns it to the Labels field.
func (o *Version) SetLabels(v map[string]string) {
	o.Labels = &v
}

// GetCustomPredictor returns the CustomPredictor field value if set, zero value otherwise.
func (o *Version) GetCustomPredictor() CustomPredictor {
	if o == nil || IsNil(o.CustomPredictor) {
		var ret CustomPredictor
		return ret
	}
	return *o.CustomPredictor
}

// GetCustomPredictorOk returns a tuple with the CustomPredictor field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Version) GetCustomPredictorOk() (*CustomPredictor, bool) {
	if o == nil || IsNil(o.CustomPredictor) {
		return nil, false
	}
	return o.CustomPredictor, true
}

// HasCustomPredictor returns a boolean if a field has been set.
func (o *Version) HasCustomPredictor() bool {
	if o != nil && !IsNil(o.CustomPredictor) {
		return true
	}

	return false
}

// SetCustomPredictor gets a reference to the given CustomPredictor and assigns it to the CustomPredictor field.
func (o *Version) SetCustomPredictor(v CustomPredictor) {
	o.CustomPredictor = &v
}

// GetCreatedAt returns the CreatedAt field value if set, zero value otherwise.
func (o *Version) GetCreatedAt() time.Time {
	if o == nil || IsNil(o.CreatedAt) {
		var ret time.Time
		return ret
	}
	return *o.CreatedAt
}

// GetCreatedAtOk returns a tuple with the CreatedAt field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Version) GetCreatedAtOk() (*time.Time, bool) {
	if o == nil || IsNil(o.CreatedAt) {
		return nil, false
	}
	return o.CreatedAt, true
}

// HasCreatedAt returns a boolean if a field has been set.
func (o *Version) HasCreatedAt() bool {
	if o != nil && !IsNil(o.CreatedAt) {
		return true
	}

	return false
}

// SetCreatedAt gets a reference to the given time.Time and assigns it to the CreatedAt field.
func (o *Version) SetCreatedAt(v time.Time) {
	o.CreatedAt = &v
}

// GetUpdatedAt returns the UpdatedAt field value if set, zero value otherwise.
func (o *Version) GetUpdatedAt() time.Time {
	if o == nil || IsNil(o.UpdatedAt) {
		var ret time.Time
		return ret
	}
	return *o.UpdatedAt
}

// GetUpdatedAtOk returns a tuple with the UpdatedAt field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Version) GetUpdatedAtOk() (*time.Time, bool) {
	if o == nil || IsNil(o.UpdatedAt) {
		return nil, false
	}
	return o.UpdatedAt, true
}

// HasUpdatedAt returns a boolean if a field has been set.
func (o *Version) HasUpdatedAt() bool {
	if o != nil && !IsNil(o.UpdatedAt) {
		return true
	}

	return false
}

// SetUpdatedAt gets a reference to the given time.Time and assigns it to the UpdatedAt field.
func (o *Version) SetUpdatedAt(v time.Time) {
	o.UpdatedAt = &v
}

// GetPythonVersion returns the PythonVersion field value if set, zero value otherwise.
func (o *Version) GetPythonVersion() string {
	if o == nil || IsNil(o.PythonVersion) {
		var ret string
		return ret
	}
	return *o.PythonVersion
}

// GetPythonVersionOk returns a tuple with the PythonVersion field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Version) GetPythonVersionOk() (*string, bool) {
	if o == nil || IsNil(o.PythonVersion) {
		return nil, false
	}
	return o.PythonVersion, true
}

// HasPythonVersion returns a boolean if a field has been set.
func (o *Version) HasPythonVersion() bool {
	if o != nil && !IsNil(o.PythonVersion) {
		return true
	}

	return false
}

// SetPythonVersion gets a reference to the given string and assigns it to the PythonVersion field.
func (o *Version) SetPythonVersion(v string) {
	o.PythonVersion = &v
}

// GetModelSchema returns the ModelSchema field value if set, zero value otherwise.
func (o *Version) GetModelSchema() ModelSchema {
	if o == nil || IsNil(o.ModelSchema) {
		var ret ModelSchema
		return ret
	}
	return *o.ModelSchema
}

// GetModelSchemaOk returns a tuple with the ModelSchema field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Version) GetModelSchemaOk() (*ModelSchema, bool) {
	if o == nil || IsNil(o.ModelSchema) {
		return nil, false
	}
	return o.ModelSchema, true
}

// HasModelSchema returns a boolean if a field has been set.
func (o *Version) HasModelSchema() bool {
	if o != nil && !IsNil(o.ModelSchema) {
		return true
	}

	return false
}

// SetModelSchema gets a reference to the given ModelSchema and assigns it to the ModelSchema field.
func (o *Version) SetModelSchema(v ModelSchema) {
	o.ModelSchema = &v
}

func (o Version) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o Version) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Id) {
		toSerialize["id"] = o.Id
	}
	if !IsNil(o.ModelId) {
		toSerialize["model_id"] = o.ModelId
	}
	if !IsNil(o.MlflowRunId) {
		toSerialize["mlflow_run_id"] = o.MlflowRunId
	}
	if !IsNil(o.MlflowUrl) {
		toSerialize["mlflow_url"] = o.MlflowUrl
	}
	if !IsNil(o.ArtifactUri) {
		toSerialize["artifact_uri"] = o.ArtifactUri
	}
	if !IsNil(o.Endpoints) {
		toSerialize["endpoints"] = o.Endpoints
	}
	if !IsNil(o.Properties) {
		toSerialize["properties"] = o.Properties
	}
	if !IsNil(o.Labels) {
		toSerialize["labels"] = o.Labels
	}
	if !IsNil(o.CustomPredictor) {
		toSerialize["custom_predictor"] = o.CustomPredictor
	}
	if !IsNil(o.CreatedAt) {
		toSerialize["created_at"] = o.CreatedAt
	}
	if !IsNil(o.UpdatedAt) {
		toSerialize["updated_at"] = o.UpdatedAt
	}
	if !IsNil(o.PythonVersion) {
		toSerialize["python_version"] = o.PythonVersion
	}
	if !IsNil(o.ModelSchema) {
		toSerialize["model_schema"] = o.ModelSchema
	}
	return toSerialize, nil
}

type NullableVersion struct {
	value *Version
	isSet bool
}

func (v NullableVersion) Get() *Version {
	return v.value
}

func (v *NullableVersion) Set(val *Version) {
	v.value = val
	v.isSet = true
}

func (v NullableVersion) IsSet() bool {
	return v.isSet
}

func (v *NullableVersion) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableVersion(val *Version) *NullableVersion {
	return &NullableVersion{value: val, isSet: true}
}

func (v NullableVersion) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableVersion) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
