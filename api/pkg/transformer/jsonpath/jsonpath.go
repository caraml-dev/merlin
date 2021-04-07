package jsonpath

import (
	"fmt"
	"regexp"

	"github.com/oliveagle/jsonpath"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types"
)

type CompiledJSONPath struct {
	cpl    *jsonpath.Compiled
	source spec.FromJson_SourceEnum
}

const (
	Prefix = "$."

	RawRequestPrefix    = "$.raw_request"
	ModelResponsePrefix = "$.model_response"
)

var (
	sourceJsonPattern = regexp.MustCompile("\\$\\.raw_request|\\$\\.model_response")
)

func CompileJsonPath(jsonPath string) (*CompiledJSONPath, error) {
	source := spec.FromJson_RAW_REQUEST
	match := sourceJsonPattern.FindString(jsonPath)
	if match != "" {
		if match == ModelResponsePrefix {
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

func (c *CompiledJSONPath) Lookup(jsonObj types.UnmarshalledJSON) (interface{}, error) {
	return c.cpl.Lookup(jsonObj)
}

func (c *CompiledJSONPath) LookupFromSource(source types.SourceJSON) (interface{}, error) {
	sourceJson := source[c.source]
	if sourceJson == nil {
		return nil, fmt.Errorf("source json is not set: %s", c.source.String())
	}

	return c.cpl.Lookup(sourceJson)
}
