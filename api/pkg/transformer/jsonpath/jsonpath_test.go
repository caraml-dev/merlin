package jsonpath

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/oliveagle/jsonpath"
	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/pkg/transformer/types"
)

const (
	rawRequestJson = `
		{
		  "signature_name" : "predict",
		  "instances": [
			{"sepal_length":2.8, "sepal_width":1.0, "petal_length":6.8, "petal_width":0.4},
			{"sepal_length":0.1, "sepal_width":0.5, "petal_length":1.8, "petal_width":2.4}
		  ]
		}
		`

	modelResponseJson = `
		{
		  "predictions": [
			1, 2
		  ],
          "model_name" : "iris-classifier"
		}
    `
)

func TestCompileJsonPath(t *testing.T) {
	tests := []struct {
		name     string
		jsonPath string
		want     *Compiled
		wantErr  bool
	}{
		{
			"without jsonContainer json",
			"$.book",
			&Compiled{
				cpl:    jsonpath.MustCompile("$.book"),
				source: spec.JsonType_RAW_REQUEST,
			},
			false,
		},
		{
			"using raw request as jsonContainer json",
			"$.raw_request.book",
			&Compiled{
				cpl:    jsonpath.MustCompile("$.book"),
				source: spec.JsonType_RAW_REQUEST,
			},
			false,
		},
		{
			"using model response as jsonContainer json",
			"$.model_response.book",
			&Compiled{
				cpl:    jsonpath.MustCompile("$.book"),
				source: spec.JsonType_MODEL_RESPONSE,
			},
			false,
		},
		{
			"nested case",
			"$.nested.model_response.book",
			&Compiled{
				cpl:    jsonpath.MustCompile("$.nested.model_response.book"),
				source: spec.JsonType_RAW_REQUEST,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Compile(tt.jsonPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Compile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Compile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMustCompileJsonPath(t *testing.T) {

	tests := []struct {
		name     string
		jsonPath string
		want     *Compiled
		wantErr  bool
		expError error
	}{
		{
			"succeful compile",
			"$.book",
			&Compiled{
				cpl:    jsonpath.MustCompile("$.book"),
				source: spec.JsonType_RAW_REQUEST,
			},
			false,
			nil,
		},
		{
			"without jsonContainer json",
			".book",
			nil,
			true,
			errors.New("should start with '$'"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.PanicsWithError(t, tt.expError.Error(), func() {
					MustCompileJsonPath(tt.jsonPath)
				})
				return
			}

			if got := MustCompileJsonPath(tt.jsonPath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MustCompileJsonPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompiled_Lookup(t *testing.T) {

	var rawRequestData types.JSONObject
	json.Unmarshal([]byte(rawRequestJson), &rawRequestData)

	type fields struct {
		cpl    *jsonpath.Compiled
		source spec.JsonType
	}

	tests := []struct {
		name    string
		fields  fields
		jsonObj types.JSONObject
		want    interface{}
		wantErr bool
	}{
		{
			"test lookup",
			fields{
				cpl:    jsonpath.MustCompile("$.signature_name"),
				source: spec.JsonType_RAW_REQUEST,
			},
			rawRequestData,
			"predict",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Compiled{
				cpl:    tt.fields.cpl,
				source: tt.fields.source,
			}
			got, err := c.Lookup(tt.jsonObj)
			if (err != nil) != tt.wantErr {
				t.Errorf("Lookup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Lookup() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompiled_LookupFromContainer(t *testing.T) {
	var rawRequestData types.JSONObject
	var modelResponseData types.JSONObject

	json.Unmarshal([]byte(rawRequestJson), &rawRequestData)
	json.Unmarshal([]byte(modelResponseJson), &modelResponseData)

	jsonContainer := types.JSONObjectContainer{
		spec.JsonType_RAW_REQUEST:    rawRequestData,
		spec.JsonType_MODEL_RESPONSE: modelResponseData,
	}

	type fields struct {
		cpl           *jsonpath.Compiled
		jsonContainer spec.JsonType
	}
	tests := []struct {
		name        string
		fields      fields
		sourceJSONs types.JSONObjectContainer
		want        interface{}
		wantErr     bool
		expErr      error
	}{
		{
			"json path to the raw request",
			fields{
				cpl:           jsonpath.MustCompile("$.signature_name"),
				jsonContainer: spec.JsonType_RAW_REQUEST,
			},
			jsonContainer,
			"predict",
			false,
			nil,
		},
		{
			"json path to the model response",
			fields{
				cpl:           jsonpath.MustCompile("$.model_name"),
				jsonContainer: spec.JsonType_MODEL_RESPONSE,
			},
			jsonContainer,
			"iris-classifier",
			false,
			nil,
		},
		{
			"json path from unset json source",
			fields{
				cpl:           jsonpath.MustCompile("$.model_name"),
				jsonContainer: spec.JsonType_INVALID,
			},
			jsonContainer,
			nil,
			true,
			errors.New("container json is not set: INVALID"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Compiled{
				cpl:    tt.fields.cpl,
				source: tt.fields.jsonContainer,
			}
			got, err := c.LookupFromContainer(tt.sourceJSONs)
			if (err != nil) != tt.wantErr {
				t.Errorf("LookupFromContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LookupFromContainer() got = %v, want %v", got, tt.want)
			}
		})
	}
}
