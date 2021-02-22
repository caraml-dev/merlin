package feast

import (
	"errors"
	"fmt"
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/mmcloughlin/geohash"
	"github.com/oliveagle/jsonpath"
	"reflect"
	"strconv"
)

type UdfResult struct {
	Value interface{}
	Error error
}

type UdfEnv struct {
	UnmarshalledJsonRequest interface{}
}

func (env UdfEnv) Geohash(latitudeJsonPath, longitudeJsonPath string, precision uint) UdfResult {
	value, err := extractGeohash(env.UnmarshalledJsonRequest, latitudeJsonPath, longitudeJsonPath, precision)
	return UdfResult{
		Value: value,
		Error: err,
	}
}

func toList(o interface{}) []interface{} {
	var listInterface []interface{}
	switch o.(type) {
	case []interface{}:
		listInterface = o.([]interface{})
	case interface{}:
		listInterface = []interface{}{o}
	}
	return listInterface
}

func toFloat64(o interface{}) (float64, error) {
	var floatValue float64
	switch o.(type) {
	case float64:
		floatValue = o.(float64)
	case string:
		parsed, err := strconv.ParseFloat(o.(string), 64)
		if err != nil {
			return 0, err
		}
		floatValue = parsed
	case int:
		floatValue = float64(o.(int))
	default:
		return 0, errors.New(fmt.Sprintf("%v cannot be parsed into float64", reflect.TypeOf(o)))
	}
	return floatValue, nil

}

func extractGeohash(jsonRequestBody interface{}, latitudeJsonPath, longitudeJsonPath string, precision uint) (interface{}, error) {
	latitude, err := jsonpath.JsonPathLookup(jsonRequestBody, latitudeJsonPath)
	if err != nil {
		return nil, err
	}
	latitudes := toList(latitude)
	if len(latitudes) == 0 {
		return nil, errors.New("empty arrays of latitudes provided")
	}

	longitude, err := jsonpath.JsonPathLookup(jsonRequestBody, longitudeJsonPath)
	if err != nil {
		return nil, err
	}
	longitudes := toList(longitude)
	if len(longitudes) == 0 {
		return nil, errors.New("empty arrays of longitudes provided")
	}

	geohashes := make([]string, 0)
	for _, lat := range latitudes {
		for _, long := range longitudes {
			latFloat, err := toFloat64(lat)
			if err != nil {
				return nil, err
			}
			longFloat, err := toFloat64(long)
			if err != nil {
				return nil, err
			}
			geohashes = append(geohashes, geohash.EncodeWithPrecision(latFloat, longFloat, precision))
		}
	}

	if len(geohashes) == 1 {
		return geohashes[0], nil
	} else {
		return geohashes, nil
	}

}

func mustCompileUdf(udfString string) *vm.Program {
	compiled, err := expr.Compile(udfString, expr.Env(UdfEnv{}))
	if err != nil {
		panic(err)
	}
	return compiled
}
