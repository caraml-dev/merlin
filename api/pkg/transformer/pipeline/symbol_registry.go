package pipeline

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-gota/gota/series"

	"github.com/gojek/merlin/pkg/transformer/pipeline/function"
)

// All exported methods in SymbolRegistry class will be accessible from any expression as built-in function
type SymbolRegistry map[string]interface{}

func NewRegistry() SymbolRegistry {
	return SymbolRegistry{}
}

func (sr SymbolRegistry) JsonExtract(parentJsonPath, nestedJsonPath string) interface{} {
	jsonBody, err := sr.evalArg(parentJsonPath)
	if err != nil {
		panic(err)
	}

	pipelineEnv := sr.getEnvironment()
	cplJsonPath := pipelineEnv.CompiledJSONPath(nestedJsonPath)
	if cplJsonPath == nil {
		c, err := CompileJsonPath(nestedJsonPath)
		if err != nil {
			panic(err)
		}
		pipelineEnv.AddCompiledJsonPath(nestedJsonPath, c)
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

func (sr SymbolRegistry) Now() time.Time {
	return time.Now()
}

func (sr SymbolRegistry) Geohash(latitude interface{}, longitude interface{}, precision uint) interface{} {
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

func (sr SymbolRegistry) S2ID(latitude interface{}, longitude interface{}, level int) interface{} {
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

func (sr SymbolRegistry) getEnvironment() *Environment {
	return sr[EnvironmentKey].(*Environment)
}

// evalArg evaluate argument
// the argument can be: values or json path string
// if it's json path string, evalArg will extract the value from json path otherwise it will return as is
func (sr SymbolRegistry) evalArg(arg interface{}) (interface{}, error) {
	switch val := arg.(type) {
	case string:
		if !strings.HasPrefix(val, jsonPathPrefix) {
			return arg, nil
		}

		pipelineEnv := sr.getEnvironment()
		cplJsonPath := pipelineEnv.CompiledJSONPath(val)
		if cplJsonPath == nil {
			c, err := CompileJsonPath(val)
			if err != nil {
				return nil, err
			}
			pipelineEnv.AddCompiledJsonPath(val, c)
			cplJsonPath = c
		}

		return cplJsonPath.LookupEnv(pipelineEnv)
	case series.Series:
		switch val.Type() {
		case series.String:
			return val.Records(), nil
		case series.Float:
			return val.Float(), nil
		case series.Int:
			return val.Int()
		case series.Bool:
			return val.Bool()
		default:
			return nil, fmt.Errorf("unknown series type")
		}
	default:
		return arg, nil
	}
}
