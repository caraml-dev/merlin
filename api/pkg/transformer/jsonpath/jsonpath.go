package jsonpath

import (
	"context"
	"fmt"
	"reflect"
	"regexp"

	"github.com/caraml-dev/protopath"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/pkg/transformer/types/converter"
	"github.com/gojekfarm/jsonpath"
	"google.golang.org/protobuf/proto"
)

// SourceType indicates type of source for jsonpath value extraction
type SourceType int

const (
	// Map source type is in map type
	Map SourceType = iota
	// Proto source type is in protobuf.Message type
	Proto

	Prefix = "$."

	RawRequestPrefix    = "$.raw_request"
	ModelResponsePrefix = "$.model_response"
)

var sourceJsonPattern = regexp.MustCompile(`\$\.raw_request|\$.model_response`)

// JsonpathExtractor is interface that extract value given the source and path
type JsonpathExtractor interface {
	Lookup(obj any) (any, error)
}

// protopathImpl JsonpathExtractor for proto source type
type protopathImpl struct {
	compiled *protopath.Compiled
}

// Lookup extract value based on the compiled jsonpath syntax
func (p *protopathImpl) Lookup(obj any) (any, error) {
	return p.compiled.Lookup(context.TODO(), obj.(proto.Message))
}

// JsonPathOption holds information about the required jsonpath value and the optional default value
type JsonPathOption struct {
	JsonPath     string
	DefaultValue string
	TargetType   spec.ValueType
	SrcType      SourceType
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
func compile(jsonPath string, sourceType SourceType) (*Compiled, error) {
	source := spec.PayloadType_RAW_REQUEST
	match := sourceJsonPattern.FindString(jsonPath)
	if match != "" {
		if match == ModelResponsePrefix {
			source = spec.PayloadType_MODEL_RESPONSE
		}

		jsonPath = sourceJsonPattern.ReplaceAllString(jsonPath, "$")
	}

	compiledJsonpath, err := createJsonpathCompiler(jsonPath, sourceType)
	if err != nil {
		return nil, err
	}

	return &Compiled{
		cpl:    compiledJsonpath,
		source: source,
	}, nil
}

func createJsonpathCompiler(path string, sourceType SourceType) (JsonpathExtractor, error) {
	if sourceType == Map {
		return jsonpath.Compile(path)
	}
	compiledJsonpath, err := protopath.NewJsonPathCompiler(path)
	if err != nil {
		return nil, err
	}
	return &protopathImpl{compiledJsonpath}, nil
}

// CompileWithOption compiles JsonPathOption into a jsonpath.Compiled instance
// JsonPathOption allow setting default value when existing value of given jsonpath is nil
func CompileWithOption(option JsonPathOption) (*Compiled, error) {
	compiled, err := compile(option.JsonPath, option.SrcType)
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

// MustCompileJsonPath will compile jsonpath using map as source type
// will panic if got error
func MustCompileJsonPath(jsonPath string) *Compiled {
	cpl, err := compile(jsonPath, Map)
	if err != nil {
		panic(err)
	}
	return cpl
}

// MustCompileJsonPathWithOption will compile jsonpath by specifying the option
// will panic if got error
func MustCompileJsonPathWithOption(option JsonPathOption) *Compiled {
	cpl, err := CompileWithOption(option)
	if err != nil {
		panic(err)
	}
	return cpl
}

// Compiled contains information about the compilation of operation based on given jsonpath syntax
type Compiled struct {
	cpl          JsonpathExtractor
	source       spec.PayloadType
	defaultValue interface{}
}

// Lookup extract value based on the compiled jsonpath syntax by giving the source object
func (c *Compiled) Lookup(obj types.Payload) (interface{}, error) {
	val, err := c.cpl.Lookup(obj.OriginalValue())
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

// LookupFromContainer extract the value from payload container where it contains request and model response payload
func (c *Compiled) LookupFromContainer(container types.PayloadObjectContainer) (interface{}, error) {
	sourceJson := container[c.source]
	if sourceJson == nil {
		return nil, fmt.Errorf("container json is not set: %s", c.source.String())
	}

	return c.Lookup(sourceJson)
}
