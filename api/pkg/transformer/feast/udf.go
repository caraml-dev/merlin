package feast

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/golang/geo/s2"
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
	latLong, err := extractLatLong(env.UnmarshalledJsonRequest, latitudeJsonPath, longitudeJsonPath)
	if err != nil {
		return UdfResult{
			Error: err,
		}
	}
	switch latLong.(type) {
	case []*LatLong:
		var value []interface{}
		for _, ll := range latLong.([]*LatLong) {
			value = append(value, ll.toGeoHash(precision))
		}
		return UdfResult{
			Value: value,
		}
	case *LatLong:
		return UdfResult{
			Value: latLong.(*LatLong).toGeoHash(precision),
		}
	}
	return UdfResult{
		Error: fmt.Errorf("unknown type from type casting"),
	}
}

func (env UdfEnv) JsonExtract(keyJsonPath, nestedJsonPath string) UdfResult {
	value, err := jsonExtract(env.UnmarshalledJsonRequest, keyJsonPath, nestedJsonPath)
	return UdfResult{
		Value: value,
		Error: err,
	}
}

func (env UdfEnv) S2ID(latitudeJsonPath, longitudeJsonPath string, level int) UdfResult {
	latLong, err := extractLatLong(env.UnmarshalledJsonRequest, latitudeJsonPath, longitudeJsonPath)
	if err != nil {
		return UdfResult{
			Error: err,
		}
	}
	switch latLong.(type) {
	case []*LatLong:
		var value []interface{}
		for _, ll := range latLong.([]*LatLong) {
			value = append(value, ll.toS2ID(level))
		}
		return UdfResult{
			Value: value,
		}
	case *LatLong:
		return UdfResult{
			Value: latLong.(*LatLong).toS2ID(level),
		}
	}
	return UdfResult{
		Error: fmt.Errorf("unknown type from type casting"),
	}
}

func jsonExtract(jsonRequestBody interface{}, keyJsonPath, nestedJsonPath string) (interface{}, error) {
	node, err := jsonpath.JsonPathLookup(jsonRequestBody, keyJsonPath)
	if err != nil {
		return nil, err
	}

	nodeString, ok := node.(string)
	if ok != true {
		return nil, fmt.Errorf("the value specified in path `%s` should be of string type", keyJsonPath)
	}
	var js map[string]interface{}
	if err := json.Unmarshal([]byte(nodeString), &js); err != nil {
		return nil, fmt.Errorf("the value specified in path `%s` should be a valid JSON", keyJsonPath)
	}

	innerNode, err := jsonpath.JsonPathLookup(js, nestedJsonPath)
	if err != nil {
		return nil, err
	}
	return innerNode, nil
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

type LatLong struct {
	lat  float64
	long float64
}

func extractLatLong(jsonRequestBody interface{}, latitudeJsonPath, longitudeJsonPath string) (interface{}, error) {
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

	var latLong interface{}

	switch latitude.(type) {
	case []interface{}:
		if len(latitude.([]interface{})) != len(longitude.([]interface{})) {
			return nil, errors.New("both latitude and longitude arrays must have the same length")
		}

		if len(latitude.([]interface{})) == 0 {
			return nil, errors.New("empty arrays of latitudes and longitude provided")
		}

		var latLongArray []*LatLong
		for index, lat := range latitude.([]interface{}) {
			latitudeFloat, err := toFloat64(lat)
			if err != nil {
				return nil, err
			}

			longitudeFloat, err := toFloat64(longitude.([]interface{})[index])
			if err != nil {
				return nil, err
			}
			latLongArray = append(latLongArray, &LatLong{
				lat:  latitudeFloat,
				long: longitudeFloat,
			})
		}
		latLong = latLongArray

	case interface{}:
		latitudeFloat, err := toFloat64(latitude)
		if err != nil {
			return nil, err
		}
		longitudeFloat, err := toFloat64(longitude)
		if err != nil {
			return nil, err
		}
		latLong = &LatLong{
			lat:  latitudeFloat,
			long: longitudeFloat,
		}
	}

	return latLong, nil
}

func (l *LatLong) toGeoHash(precision uint) interface{} {
	return geohash.EncodeWithPrecision(l.lat, l.long, precision)
}

func (l *LatLong) toS2ID(level int) interface{} {
	return strconv.FormatUint(s2.CellFromLatLng(s2.LatLngFromDegrees(l.lat, l.long)).ID().Parent(level).Pos(), 10)
}

func mustCompileUdf(udfString string) *vm.Program {
	compiled, err := expr.Compile(udfString, expr.Env(UdfEnv{}))
	if err != nil {
		panic(err)
	}
	return compiled
}
