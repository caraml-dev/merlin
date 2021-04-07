package pipeline

import (
	"fmt"
	"regexp"

	"github.com/oliveagle/jsonpath"

	"github.com/gojek/merlin/pkg/transformer"
	"github.com/gojek/merlin/pkg/transformer/spec"
)

type CompiledJSONPath struct {
	cpl    *jsonpath.Compiled
	source spec.FromJson_SourceEnum
}

const (
	jsonPathPrefix              = "$."
	jsonPathRawRequestPrefix    = "$.raw_request"
	jsonPathModelResponsePrefix = "$.model_response"
)

var (
	sourceJsonPattern = regexp.MustCompile("\\$\\.raw_request|\\$\\.model_response")
)

func CompileJsonPath(jsonPath string) (*CompiledJSONPath, error) {
	source := spec.FromJson_RAW_REQUEST
	match := sourceJsonPattern.FindString(jsonPath)
	if match != "" {
		if match == jsonPathModelResponsePrefix {
			source = spec.FromJson_MODEL_RESPONSE
		}

		jsonPath = sourceJsonPattern.ReplaceAllString(jsonPath, "$")
	}

	compiledJsonpath, err := jsonpath.Compile(jsonPath)
	if err != nil {
		return nil, err
	}

	return &CompiledJSONPath{
		cpl:    compiledJsonpath,
		source: source,
	}, nil
}

func MustCompileJsonPath(jsonPath string) *CompiledJSONPath {
	cpl, err := CompileJsonPath(jsonPath)
	if err != nil {
		panic(err)
	}
	return cpl
}

func (c *CompiledJSONPath) Lookup(jsonObj transformer.UnmarshalledJSON) (interface{}, error) {
	return c.cpl.Lookup(jsonObj)
}

func (c *CompiledJSONPath) LookupEnv(env *Environment) (interface{}, error) {
	sourceJson := env.SourceJSON(c.source)
	if sourceJson == nil {
		return nil, fmt.Errorf("source json is not set: %s", c.source.String())
	}

	return c.cpl.Lookup(sourceJson)
}
