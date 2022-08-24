package jsonpath

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/gojekfarm/jsonpath"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/pkg/transformer/types/converter"
)

const (
	Prefix = "$."

	RawRequestPrefix    = "$.raw_request"
	ModelResponsePrefix = "$.model_response"
)

var sourceJsonPattern = regexp.MustCompile("\\$\\.raw_request|\\$\\.model_response")

// JsonPathOption holds information about the required jsonpath value and the optional default value
type JsonPathOption struct {
	JsonPath     string
	DefaultValue string
	TargetType   spec.ValueType
}

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

// CompileWithOption compiles JsonPathOption into a jsonpath.Compiled instance
// JsonPathOption allow setting default value when existing value of given jsonpath is nil
func CompileWithOption(option JsonPathOption) (*Compiled, error) {
	compiled, err := Compile(option.JsonPath)
	if err != nil {
		return nil, err
	}
	if option.DefaultValue == "" {
		return compiled, nil
	}

	defaultValue, err := converter.ToTargetType(option.DefaultValue, option.TargetType)
	if err != nil {
		return nil, err
	}

	compiled.defaultValue = defaultValue
	return compiled, nil
}

func MustCompileJsonPath(jsonPath string) *Compiled {
	cpl, err := Compile(jsonPath)
	if err != nil {
		panic(err)
	}
	return cpl
}

func MustCompileJsonPathWithOption(option JsonPathOption) *Compiled {
	cpl, err := CompileWithOption(option)
	if err != nil {
		panic(err)
	}
	return cpl
}

type Compiled struct {
	cpl          *jsonpath.Compiled
	source       spec.JsonType
	defaultValue interface{}
}

func (c *Compiled) Lookup(obj types.Payload) (interface{}, error) {
	val, err := c.cpl.Lookup(obj)
	if err != nil {
		return nil, err
	}
	if val == nil {
		val = c.defaultValue
	}

	if reflectVal := reflect.ValueOf(val); reflectVal.Kind() == reflect.Slice {
		if reflectVal.Len() == 0 && c.defaultValue != nil {
			return []interface{}{c.defaultValue}, nil
		}

		sliceVal, _ := val.([]interface{})
		for idx, v := range sliceVal {
			if v == nil {
				sliceVal[idx] = c.defaultValue
			}
		}

		return sliceVal, nil
	}

	return val, nil
}

func (c *Compiled) LookupFromContainer(container types.PayloadObjectContainer) (interface{}, error) {
	sourceJson := container[c.source]
	if sourceJson == nil {
		return nil, fmt.Errorf("container json is not set: %s", c.source.String())
	}

	return c.Lookup(sourceJson)
}
