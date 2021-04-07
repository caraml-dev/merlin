package pipeline

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/gojek/merlin/pkg/transformer"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/oliveagle/jsonpath"
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
		want     *CompiledJSONPath
		wantErr  bool
	}{
		{
			"without source json",
			"$.book",
			&CompiledJSONPath{
				cpl:    jsonpath.MustCompile("$.book"),
				source: spec.FromJson_RAW_REQUEST,
			},
			false,
		},
		{
			"using raw request as source json",
			"$.raw_request.book",
			&CompiledJSONPath{
				cpl:    jsonpath.MustCompile("$.book"),
				source: spec.FromJson_RAW_REQUEST,
			},
			false,
		},
		{
			"using model response as source json",
			"$.model_response.book",
			&CompiledJSONPath{
				cpl:    jsonpath.MustCompile("$.book"),
				source: spec.FromJson_MODEL_RESPONSE,
			},
			false,
		},
		{
			"nested case",
			"$.nested.model_response.book",
			&CompiledJSONPath{
				cpl:    jsonpath.MustCompile("$.nested.model_response.book"),
				source: spec.FromJson_RAW_REQUEST,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompileJsonPath(tt.jsonPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompileJsonPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CompileJsonPath() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompiled_Lookup(t *testing.T) {

	var rawRequestData transformer.UnmarshalledJSON
	json.Unmarshal([]byte(rawRequestJson), &rawRequestData)

	type fields struct {
		cpl    *jsonpath.Compiled
		source spec.FromJson_SourceEnum
	}

	tests := []struct {
		name    string
		fields  fields
		jsonObj transformer.UnmarshalledJSON
		want    interface{}
		wantErr bool
	}{
		{
			"test lookup",
			fields{
				cpl:    jsonpath.MustCompile("$.signature_name"),
				source: spec.FromJson_RAW_REQUEST,
			},
			rawRequestData,
			"predict",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CompiledJSONPath{
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

func TestCompiled_LookupEnv(t *testing.T) {
	var rawRequestData transformer.UnmarshalledJSON
	var modelResponseData transformer.UnmarshalledJSON

	json.Unmarshal([]byte(rawRequestJson), &rawRequestData)
	json.Unmarshal([]byte(modelResponseJson), &modelResponseData)

	env := &Environment{
		sourceJson: map[spec.FromJson_SourceEnum]transformer.UnmarshalledJSON{
			spec.FromJson_RAW_REQUEST:    rawRequestData,
			spec.FromJson_MODEL_RESPONSE: modelResponseData,
		},
	}

	type fields struct {
		cpl    *jsonpath.Compiled
		source spec.FromJson_SourceEnum
	}
	tests := []struct {
		name    string
		fields  fields
		env     *Environment
		want    interface{}
		wantErr bool
	}{
		{
			"json path to the raw request",
			fields{
				cpl:    jsonpath.MustCompile("$.signature_name"),
				source: spec.FromJson_RAW_REQUEST,
			},
			env,
			"predict",
			false,
		},
		{
			"json path to the model response",
			fields{
				cpl:    jsonpath.MustCompile("$.model_name"),
				source: spec.FromJson_MODEL_RESPONSE,
			},
			env,
			"iris-classifier",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CompiledJSONPath{
				cpl:    tt.fields.cpl,
				source: tt.fields.source,
			}
			got, err := c.LookupEnv(tt.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("LookupEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LookupEnv() got = %v, want %v", got, tt.want)
			}
		})
	}
}
