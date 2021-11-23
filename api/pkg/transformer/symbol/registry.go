package symbol

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol/function"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/pkg/transformer/types/converter"
	"github.com/gojek/merlin/pkg/transformer/types/series"
)

// Registry contains all symbol (variable and functions) that can be used for expression evaluation
// All keys within Registry can be used as variable in an expression
// All exported method of Registry is accessible as built-in function
type Registry map[string]interface{}

func NewRegistryWithCompiledJSONPath(compiledJSONPaths *jsonpath.Storage) Registry {
	r := Registry{}
	r[compiledJSONPathKey] = compiledJSONPaths
	r[sourceJSONKey] = types.JSONObjectContainer{}

	return r
}

func NewRegistry() Registry {
	r := Registry{}
	r[compiledJSONPathKey] = jsonpath.NewStorage()
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

// HaversineDistance of two points (latitude, longitude)
// latitude and longitude can be:
// - Json path string
// - Slice / gota.Series
// - float64 value
func (sr Registry) HaversineDistance(latitude1 interface{}, longitude1 interface{}, latitude2 interface{}, longitude2 interface{}) interface{} {
	lat1, err := sr.evalArg(latitude1)
	if err != nil {
		panic(err)
	}

	lon1, err := sr.evalArg(longitude1)
	if err != nil {
		panic(err)
	}

	lat2, err := sr.evalArg(latitude2)
	if err != nil {
		panic(err)
	}

	lon2, err := sr.evalArg(longitude2)
	if err != nil {
		panic(err)
	}
	result, err := function.HaversineDistance(lat1, lon1, lat2, lon2)
	if err != nil {
		panic(err)
	}

	return result
}

// PolarAngle calculate polar angle of two locations given latitude1, longitude1, latitude1, latitude2
// latitude and longitude can be:
// - Json path string
// - Slice / gota.Series
// - float64 value
func (sr Registry) PolarAngle(latitude1 interface{}, longitude1 interface{}, latitude2 interface{}, longitude2 interface{}) interface{} {
	lat1, err := sr.evalArg(latitude1)
	if err != nil {
		panic(err)
	}

	lon1, err := sr.evalArg(longitude1)
	if err != nil {
		panic(err)
	}

	lat2, err := sr.evalArg(latitude2)
	if err != nil {
		panic(err)
	}

	lon2, err := sr.evalArg(longitude2)
	if err != nil {
		panic(err)
	}
	result, err := function.PolarAngle(lat1, lon1, lat2, lon2)
	if err != nil {
		panic(err)
	}

	return result
}

// ParseTimeStamp convert timestamp value into time
func (sr Registry) ParseTimestamp(timestamp interface{}) interface{} {
	timeFn := func(ts int64, tz *time.Location) interface{} {
		return time.Unix(ts, 0).In(tz)
	}
	// empty timezone mean using UTC
	return sr.processTimestampFunction(timestamp, "", timeFn)
}

// IsWeekend check wheter given timestamps falls in weekend, given timestamp and timezone
// timestamp can be:
// - Json path string
// - Slice / gota.Series
// - int64 value
func (sr Registry) IsWeekend(timestamp interface{}, timezone string) interface{} {
	timeFn := func(ts int64, tz *time.Location) interface{} {
		return function.IsWeekend(ts, tz)
	}
	return sr.processTimestampFunction(timestamp, timezone, timeFn)
}

func (sr Registry) ParseTimeStampsWithFormat(timestamp interface{}, timezone, format string) interface{} {
	timeFn := func(ts int64, tz *time.Location) interface{} {
		return function.ParseTimestampIntoFormattedString(ts, tz, format)
	}
	return sr.processTimestampFunction(timestamp, timezone, timeFn)
}

// DayOfWeek will return number represent day in a week, given timestamp and timezone
// SUNDAY(0), MONDAY(1), TUESDAY(2), WEDNESDAY(3), THURSDAY(4), FRIDAY(5), SATURDAY(6)
// timestamp can be:
// - Json path string
// - Slice / gota.Series
// - int64 value
func (sr Registry) DayOfWeek(timestamp interface{}, timezone string) interface{} {
	timeFn := func(ts int64, tz *time.Location) interface{} {
		return function.DayOfWeek(ts, tz)
	}
	return sr.processTimestampFunction(timestamp, timezone, timeFn)
}

func (sr Registry) processTimestampFunction(timestamps interface{}, timezone string, timeTransformerFn func(timestamp int64, tz *time.Location) interface{}) interface{} {
	timeLocation, err := time.LoadLocation(timezone)
	if err != nil {
		panic(err)
	}

	ts, err := sr.evalArg(timestamps)
	if err != nil {
		panic(err)
	}
	timestampVals := reflect.ValueOf(ts)
	switch timestampVals.Kind() {
	case reflect.Slice:
		var values []interface{}
		for idx := 0; idx < timestampVals.Len(); idx++ {
			val := timestampVals.Index(idx)
			tsInt64, err := converter.ToInt64(val.Interface())
			if err != nil {
				panic(err)
			}
			values = append(values, timeTransformerFn(tsInt64, timeLocation))
		}
		return values
	default:
		tsInt64, err := converter.ToInt64(ts)
		if err != nil {
			panic(err)
		}
		return timeTransformerFn(tsInt64, timeLocation)
	}
}

// CumulativeValue is function that accumulate values based on the index
// e.g values [1,2,3] => [1, 1+2, 1+2+3] => [1, 3, 6]
// values should be in array format
func (sr Registry) CumulativeValue(values interface{}) []float64 {
	evalValues, err := sr.evalArg(values)
	if err != nil {
		panic(err)
	}
	vals := reflect.ValueOf(evalValues)
	switch vals.Kind() {
	case reflect.Slice:
		values := make([]float64, 0, vals.Len())
		for idx := 0; idx < vals.Len(); idx++ {
			val := vals.Index(idx)
			floatVal, err := converter.ToFloat64(val.Interface())
			if err != nil {
				panic(err)
			}
			prevCumulativeVal := float64(0)
			if idx > 0 {
				prevCumulativeVal = values[idx-1]
			}
			values = append(values, prevCumulativeVal+floatVal)
		}
		return values
	default:
		panic("the values should in array format")
	}
}

// Now() returns current local time
func (sr Registry) Now() time.Time {
	return time.Now()
}

func (sr Registry) SetRawRequestJSON(jsonObj types.JSONObject) {
	sr[sourceJSONKey].(types.JSONObjectContainer)[spec.JsonType_RAW_REQUEST] = jsonObj
}

func (sr Registry) SetModelResponseJSON(jsonObj types.JSONObject) {
	sr[sourceJSONKey].(types.JSONObjectContainer)[spec.JsonType_MODEL_RESPONSE] = jsonObj
}

func (sr Registry) JSONContainer() types.JSONObjectContainer {
	return sr[sourceJSONKey].(types.JSONObjectContainer)
}

func (sr Registry) SetRawRequestHeaders(headers map[string]string) {
	sr[rawRequestHeadersKey] = headers
}

func (sr Registry) SetModelResponseHeaders(headers map[string]string) {
	sr[modelResponseHeadersKey] = headers
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
		return val.GetRecords(), nil
	default:
		return arg, nil
	}
}

func (sr Registry) getCompiledJSONPath(jsonPath string) *jsonpath.Compiled {
	compiledJSONPaths, ok := sr[compiledJSONPathKey].(*jsonpath.Storage)
	if !ok {
		return nil
	}

	return compiledJSONPaths.Get(jsonPath)
}

func (sr Registry) addCompiledJsonPath(jsonPath string, c *jsonpath.Compiled) {
	sr[compiledJSONPathKey].(*jsonpath.Storage).Set(jsonPath, c)
}

func (sr Registry) jsonObjectContainer() types.JSONObjectContainer {
	p, ok := sr[sourceJSONKey]
	if !ok {
		return nil
	}
	return p.(types.JSONObjectContainer)
}
