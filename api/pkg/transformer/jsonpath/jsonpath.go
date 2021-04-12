package jsonpath

import (
	"fmt"
	"regexp"

	"github.com/oliveagle/jsonpath"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types"
)

const (
	Prefix = "$."

	RawRequestPrefix    = "$.raw_request"
	ModelResponsePrefix = "$.model_response"
)

var (
	sourceJsonPattern = regexp.MustCompile("\\$\\.raw_request|\\$\\.model_response")
)

// Compile compile a jsonPath string into a jsonpath.Compiled instance
// jsonPath string should follow format specified in http://goessner.net/articles/JsonPath/
// In the jsonPath string user can also specify which json payload to extract from by prefixing the json field selector
// either with "$.raw_request" or "$.model_response", in which Compiled.LookupFromContainer will either use the transformer raw request payload or
// model response payload. If the prefix is not defined Compiled.LookupFromContainer will use raw request payload as default
// E.g.:
// "$.book" : Compiled.LookupFromContainer extract "book" field from raw request payload
// "$.raw_request.book" : Compiled.LookupFromContainer extract "book" field from raw request payload
// "$.model_response.book" : Compiled.LookupFromContainer extract "book" field from model response payload
func Compile(jsonPath string) (*Compiled, error) {
	source := spec.JsonType_RAW_REQUEST
	match := sourceJsonPattern.FindString(jsonPath)
	if match != "" {
		if match == ModelResponsePrefix {
			source = spec.JsonType_MODEL_RESPONSE
		}

		jsonPath = sourceJsonPattern.ReplaceAllString(jsonPath, "$")
	}

	compiledJsonpath, err := jsonpath.Compile(jsonPath)
	if err != nil {
		return nil, err
	}

	return &Compiled{
		cpl:    compiledJsonpath,
		source: source,
	}, nil
}

func MustCompileJsonPath(jsonPath string) *Compiled {
	cpl, err := Compile(jsonPath)
	if err != nil {
		panic(err)
	}
	return cpl
}

type Compiled struct {
	cpl    *jsonpath.Compiled
	source spec.JsonType
}

func (c *Compiled) Lookup(jsonObj types.JSONObject) (interface{}, error) {
	return c.cpl.Lookup(jsonObj)
}

func (c *Compiled) LookupFromContainer(container types.JSONObjectContainer) (interface{}, error) {
	sourceJson := container[c.source]
	if sourceJson == nil {
		return nil, fmt.Errorf("container json is not set: %s", c.source.String())
	}

	return c.cpl.Lookup(sourceJson)
}
