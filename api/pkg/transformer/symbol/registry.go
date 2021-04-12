package symbol

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol/function"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/pkg/transformer/types/series"
)

// Registry contains all symbol (variable and functions) that can be used for expression evaluation
// All keys within Registry can be used as variable in an expression
// All exported method of Registry is accessible as built-in function
type Registry map[string]interface{}

func NewRegistryWithCompiledJSONPath(compiledJSONPaths map[string]*jsonpath.Compiled) Registry {
	r := Registry{}
	r[compiledJSONPathKey] = compiledJSONPaths
	r[sourceJSONKey] = types.JSONObjectContainer{}

	return r
}

func NewRegistry() Registry {
	r := Registry{}
	r[compiledJSONPathKey] = make(map[string]*jsonpath.Compiled)
	r[sourceJSONKey] = types.JSONObjectContainer{}

	return r
}

const (
	sourceJSONKey       = "__source_json_key__"
	compiledJSONPathKey = "__compiled_jsonpath_key__"

	rawRequestHeadersKey    = "raw_request_headers"
	modelResponseHeadersKey = "model_response_headers"
)

// JsonExtract extract json field pointed by nestedJsonPath within a json string pointed by nestedJsonPath
func (sr Registry) JsonExtract(parentJsonPath, nestedJsonPath string) interface{} {
	jsonBody, err := sr.evalArg(parentJsonPath)
	if err != nil {
		panic(err)
	}

	cplJsonPath := sr.getCompiledJSONPath(nestedJsonPath)
	if cplJsonPath == nil {
		c, err := jsonpath.Compile(nestedJsonPath)
		if err != nil {
			panic(err)
		}
		sr.addCompiledJsonPath(nestedJsonPath, c)
		cplJsonPath = c
	}

	var js map[string]interface{}
	jsonBodyStr, ok := jsonBody.(string)
	if !ok {
		panic(fmt.Errorf("the value specified in path `%s` should be of string type", parentJsonPath))
	}

	if err := json.Unmarshal([]byte(jsonBodyStr), &js); err != nil {
		panic(fmt.Errorf("the value specified in path `%s` should be a valid JSON", parentJsonPath))
	}

	value, err := cplJsonPath.Lookup(js)
	if err != nil {
		panic(err)
	}

	return value
}

// Geohash calculate geohash of latitude and longitude with the given character precision
// latitude and longitude can be:
// - Json path string
// - Slice / gota.Series
// - float64 value
func (sr Registry) Geohash(latitude interface{}, longitude interface{}, precision uint) interface{} {
	lat, err := sr.evalArg(latitude)
	if err != nil {
		panic(err)
	}

	lon, err := sr.evalArg(longitude)
	if err != nil {
		panic(err)
	}

	result, err := function.Geohash(lat, lon, precision)
	if err != nil {
		panic(err)
	}

	return result
}

// S2ID calculate S2 ID of latitude and longitude of the given level
// latitude and longitude can be:
// - Json path string
// - Slice / gota.Series
// - float64 value
func (sr Registry) S2ID(latitude interface{}, longitude interface{}, level int) interface{} {
	lat, err := sr.evalArg(latitude)
	if err != nil {
		panic(err)
	}

	lon, err := sr.evalArg(longitude)
	if err != nil {
		panic(err)
	}

	result, err := function.S2ID(lat, lon, level)
	if err != nil {
		panic(err)
	}

	return result
}

// Now() returns current local time
func (sr Registry) Now() time.Time {
	return time.Now()
}

func (sr Registry) SetRawRequestJSON(jsonObj types.JSONObject) {
	sr[sourceJSONKey].(types.JSONObjectContainer)[spec.FromJson_RAW_REQUEST] = jsonObj
}

func (sr Registry) SetModelResponseJSON(jsonObj types.JSONObject) {
	sr[sourceJSONKey].(types.JSONObjectContainer)[spec.FromJson_MODEL_RESPONSE] = jsonObj
}

func (sr Registry) JSONContainer() types.JSONObjectContainer {
	return sr[sourceJSONKey].(types.JSONObjectContainer)
}

// evalArg evaluate argument
// the argument can be: values or json path string
// if it's json path string, evalArg will extract the value from json path otherwise it will return as is
func (sr Registry) evalArg(arg interface{}) (interface{}, error) {
	switch val := arg.(type) {
	case string:
		if !strings.HasPrefix(val, jsonpath.Prefix) {
			return arg, nil
		}

		cplJsonPath := sr.getCompiledJSONPath(val)
		if cplJsonPath == nil {
			c, err := jsonpath.Compile(val)
			if err != nil {
				return nil, err
			}
			sr.addCompiledJsonPath(val, c)
			cplJsonPath = c
		}

		return cplJsonPath.LookupFromContainer(sr.jsonObjectContainer())
	case *series.Series:
		switch val.Type() {
		case series.String:
			return val.Series().Records(), nil
		case series.Float:
			return val.Series().Float(), nil
		case series.Int:
			return val.Series().Int()
		case series.Bool:
			return val.Series().Bool()
		default:
			return nil, fmt.Errorf("unknown series type")
		}
	default:
		return arg, nil
	}
}

func (sr Registry) getCompiledJSONPath(jsonPath string) *jsonpath.Compiled {
	compiledJSONPaths, ok := sr[compiledJSONPathKey].(map[string]*jsonpath.Compiled)
	if !ok {
		return nil
	}

	return compiledJSONPaths[jsonPath]
}

func (sr Registry) addCompiledJsonPath(jsonPath string, c *jsonpath.Compiled) {
	sr[compiledJSONPathKey].(map[string]*jsonpath.Compiled)[jsonPath] = c
}

func (sr Registry) jsonObjectContainer() types.JSONObjectContainer {
	p, ok := sr[sourceJSONKey]
	if !ok {
		return nil
	}
	return p.(types.JSONObjectContainer)
}

func (sr Registry) SetRawRequestHeaders(headers map[string]string) {
	sr[rawRequestHeadersKey] = headers
}

func (sr Registry) SetModelResponseHeaders(headers map[string]string) {
	sr[modelResponseHeadersKey] = headers
}
