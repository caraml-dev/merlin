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

	longitude, err := jsonpath.JsonPathLookup(jsonRequestBody, longitudeJsonPath)
	if err != nil {
		return nil, err
	}

	if reflect.TypeOf(latitude) != reflect.TypeOf(longitude) {
		return nil, errors.New("latitude and longitude must have the same types")
	}

	var geohashValue interface{}

	switch latitude.(type) {
	case []interface{}:
		if len(latitude.([]interface{})) != len(longitude.([]interface{})) {
			return nil, errors.New("both latitude and longitude arrays must have the same length")
		}

		if len(latitude.([]interface{})) == 0 {
			return nil, errors.New("empty arrays of latitudes and longitude provided")
		}

		geohashArray := make([]string, len(latitude.([]interface{})))
		for index, lat := range latitude.([]interface{}) {
			latitudeFloat, err := toFloat64(lat)
			if err != nil {
				return nil, err
			}

			longitudeFloat, err := toFloat64(longitude.([]interface{})[index])
			if err != nil {
				return nil, err
			}
			geohashArray[index] = geohash.EncodeWithPrecision(latitudeFloat, longitudeFloat, precision)
		}
		geohashValue = geohashArray

	case interface{}:
		latitudeFloat, err := toFloat64(latitude)
		if err != nil {
			return nil, err
		}
		longitudeFloat, err := toFloat64(longitude)
		if err != nil {
			return nil, err
		}
		geohashValue = geohash.EncodeWithPrecision(latitudeFloat, longitudeFloat, precision)
	}

	return geohashValue, nil

}

func mustCompileUdf(udfString string) *vm.Program {
	compiled, err := expr.Compile(udfString, expr.Env(UdfEnv{}))
	if err != nil {
		panic(err)
	}
	return compiled
}
