package symbol

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/mmcloughlin/geohash"
	"github.com/stretchr/testify/assert"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/pkg/transformer/types/series"
)

var (
	requestJSONString = []byte(`{
		"latitude" : 1.0,
		"latitudeString": "1.0",
		"latitudeWrongType": "abcde",
		"location": {
			"latitude": 0.1,
			"longitude": 2.0
		},
		"latitudeArrays": [1.0, 2.0],
		"longitude" : 2.0,
		"longitudeString" : "2.0",
		"longitudeInteger": 1,
		"longitudeArrays": [1.0, 2.0],
		"longitudeLongArrays": [1.0, 2.0, 3.0]
	}`)

	responseJSONString = []byte(`{
		"latitude" : 1.0,
		"latitudeString": "1.0",
		"latitudeWrongType": "abcde",
		"location": {
			"latitude": 0.1,
			"longitude": 2.0
		},
		"latitudeArrays": [1.0, 2.0],
		"longitude" : 2.0,
		"longitudeString" : "2.0",
		"longitudeInteger": 1,
		"longitudeArrays": [1.0, 2.0],
		"longitudeLongArrays": [1.0, 2.0, 3.0]
	}`)
)

func TestSymbolRegistry_Geohash(t *testing.T) {
	requestJSONObject, responseJSONObject := getTestJSONObjects()

	sr := NewRegistryWithCompiledJSONPath(jsonpath.NewStorage())
	sr.SetRawRequestJSON(requestJSONObject)
	sr.SetModelResponseJSON(responseJSONObject)

	type args struct {
		latitude  interface{}
		longitude interface{}
		precision uint
	}
	tests := []struct {
		name       string
		args       args
		want       interface{}
		wantErr    bool
		wantErrVal error
	}{
		{
			"geohash from json path fields",
			args{
				latitude:  "$.latitude",
				longitude: "$.longitude",
				precision: 7,
			},
			geohash.EncodeWithPrecision(1.0, 2.0, 7),
			false,
			nil,
		},
		{
			"geohash from json path fields in model response",
			args{
				latitude:  "$.model_response.latitude",
				longitude: "$.model_response.longitude",
				precision: 7,
			},
			geohash.EncodeWithPrecision(1.0, 2.0, 7),
			false,
			nil,
		},
		{
			"geohash from json struct",
			args{
				latitude:  "$.location.latitude",
				longitude: "$.location.longitude",
				precision: 7,
			},
			geohash.EncodeWithPrecision(0.1, 2.0, 7),
			false,
			nil,
		},
		{
			"geohash from latitude and longitude strings",
			args{
				latitude:  "$.latitudeString",
				longitude: "$.longitudeString",
				precision: 12,
			},
			geohash.Encode(1.0, 2.0),
			false,
			nil,
		},
		{
			"geohash from latitude and longitude arrays",
			args{
				latitude:  "$.latitudeArrays",
				longitude: "$.longitudeArrays",
				precision: 12,
			},
			[]interface{}{
				geohash.Encode(1.0, 1.0),
				geohash.Encode(2.0, 2.0),
			},
			false,
			nil,
		},
		{
			"geohash from value",
			args{
				latitude:  1.0,
				longitude: 2.0,
				precision: 7,
			},
			geohash.EncodeWithPrecision(1.0, 2.0, 7),
			false,
			nil,
		},
		{
			"geohash from array value",
			args{
				latitude:  []interface{}{1.0, 2.0, 3.0},
				longitude: []interface{}{4.0, 5.0, 6.0},
				precision: 7,
			},
			[]interface{}{
				geohash.EncodeWithPrecision(1.0, 4.0, 7),
				geohash.EncodeWithPrecision(2.0, 5.0, 7),
				geohash.EncodeWithPrecision(3.0, 6.0, 7),
			},
			false,
			nil,
		},
		{
			"geohash from series value",
			args{
				latitude:  series.New([]interface{}{1.0, 2.0, 3.0}, series.Float, "latitude"),
				longitude: series.New([]interface{}{4.0, 5.0, 6.0}, series.Float, "longitude"),
				precision: 7,
			},
			[]interface{}{
				geohash.EncodeWithPrecision(1.0, 4.0, 7),
				geohash.EncodeWithPrecision(2.0, 5.0, 7),
				geohash.EncodeWithPrecision(3.0, 6.0, 7),
			},
			false,
			nil,
		},
		{
			"error: type difference",
			args{
				latitude:  "$.latitude",
				longitude: "$.longitudeArrays",
				precision: 12,
			},
			nil,
			true,
			errors.New("latitude and longitude must have the same types"),
		},
		{
			"error: type conversion",
			args{
				latitude:  "$.latitudeWrongType",
				longitude: "$.longitudeString",
				precision: 12,
			},
			nil,
			true,
			errors.New("strconv.ParseFloat: parsing \"abcde\": invalid syntax"),
		},
		{
			"error: array length different",
			args{
				latitude:  "$.latitudeArrays",
				longitude: "$.longitudeLongArrays",
				precision: 12,
			},
			nil,
			true,
			errors.New("both latitude and longitude arrays must have the same length"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.PanicsWithError(t, tt.wantErrVal.Error(), func() {
					sr.Geohash(tt.args.latitude, tt.args.longitude, tt.args.precision)
				})
				return
			}

			got := sr.Geohash(tt.args.latitude, tt.args.longitude, tt.args.precision)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Geohash() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSymbolRegistry_S2ID(t *testing.T) {
	requestJSONObject, responseJSONObject := getTestJSONObjects()

	sr := NewRegistryWithCompiledJSONPath(jsonpath.NewStorage())
	sr.SetRawRequestJSON(requestJSONObject)
	sr.SetModelResponseJSON(responseJSONObject)

	type args struct {
		latitude  interface{}
		longitude interface{}
		level     int
	}
	tests := []struct {
		name       string
		args       args
		want       interface{}
		wantErr    bool
		wantErrVal error
	}{
		{
			"s2id from json path fields",
			args{
				latitude:  "$.latitude",
				longitude: "$.longitude",
				level:     7,
			},
			"1154680723211288576",
			false,
			nil,
		},
		{
			"s2id from json path fields in model response",
			args{
				latitude:  "$.model_response.latitude",
				longitude: "$.model_response.longitude",
				level:     7,
			},
			"1154680723211288576",
			false,
			nil,
		},
		{
			"s2id from json struct",
			args{
				latitude:  "$.location.latitude",
				longitude: "$.location.longitude",
				level:     7,
			},
			"1155102935676354560",
			false,
			nil,
		},
		{
			"s2id from latitude and longitude strings",
			args{
				latitude:  "$.latitudeString",
				longitude: "$.longitudeString",
				level:     12,
			},
			"1154732743855177728",
			false,
			nil,
		},
		{
			"s2id from latitude and longitude arrays",
			args{
				latitude:  "$.latitudeArrays",
				longitude: "$.longitudeArrays",
				level:     12,
			},
			[]interface{}{
				"1153277815093723136",
				"1154346540395921408",
			},
			false,
			nil,
		},
		{
			"s2id from literal value",
			args{
				latitude:  1.3553906,
				longitude: 103.7173325,
				level:     12,
			},
			"3592201038809006080",
			false,
			nil,
		},
		{
			"s2id from array value",
			args{
				latitude:  []interface{}{1.3553906, 1.3539049},
				longitude: []interface{}{103.7173325, 103.9724141},
				level:     12,
			},
			[]interface{}{
				"3592201038809006080",
				"3592250654271209472",
			},
			false,
			nil,
		},
		{
			"s2id from series value",
			args{
				latitude:  series.New([]interface{}{1.3553906, 1.3539049}, series.Float, "latitude"),
				longitude: series.New([]interface{}{103.7173325, 103.9724141}, series.Float, "longitude"),
				level:     12,
			},
			[]interface{}{
				"3592201038809006080",
				"3592250654271209472",
			},
			false,
			nil,
		},
		{
			"error: type difference",
			args{
				latitude:  "$.latitude",
				longitude: "$.longitudeArrays",
				level:     12,
			},
			nil,
			true,
			errors.New("latitude and longitude must have the same types"),
		},
		{
			"error: type conversion",
			args{
				latitude:  "$.latitudeWrongType",
				longitude: "$.longitudeString",
				level:     12,
			},
			nil,
			true,
			errors.New("strconv.ParseFloat: parsing \"abcde\": invalid syntax"),
		},
		{
			"error: array length different",
			args{
				latitude:  "$.latitudeArrays",
				longitude: "$.longitudeLongArrays",
				level:     12,
			},
			nil,
			true,
			errors.New("both latitude and longitude arrays must have the same length"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.PanicsWithError(t, tt.wantErrVal.Error(), func() {
					sr.S2ID(tt.args.latitude, tt.args.longitude, tt.args.level)
				})
				return
			}

			got := sr.S2ID(tt.args.latitude, tt.args.longitude, tt.args.level)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Geohash() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSymbolRegistry_JsonExtract(t *testing.T) {
	testJsonString := []byte(`{
		"details": "{\"merchant_id\": 9001}",
		"nested": "{\"child_node\": { \"grandchild_node\": \"gen-z\"}}",
		"not_json": "i_am_not_json_string",
		"not_string": 1024,
		"array": "{\"child_node\": { \"array\": [1, 2]}}"
	}`)

	var jsonObject types.JSONObject
	err := json.Unmarshal(testJsonString, &jsonObject)
	if err != nil {
		panic(err)
	}

	sr := NewRegistryWithCompiledJSONPath(jsonpath.NewStorage())
	sr.SetRawRequestJSON(jsonObject)

	type args struct {
		parentJsonPath string
		nestedJsonPath string
	}
	tests := []struct {
		name       string
		args       args
		want       interface{}
		wantErr    bool
		wantErrVal error
	}{
		{
			"should be able to extract value from JSON string",
			args{"$.details",
				"$.merchant_id",
			},
			float64(9001),
			false,
			nil,
		},
		{
			"should be able to extract value using nested key from JSON string",
			args{"$.nested",
				"$.child_node.grandchild_node",
			},
			"gen-z",
			false,
			nil,
		},
		{
			"should be able to extract array value using nested key from JSON string",
			args{"$.array",
				"$.child_node.array[*]",
			},
			[]interface{}{float64(1), float64(2)},
			false,
			nil,
		},
		{
			"should throw error when value specified by key does not exist in nested JSON",
			args{"$.nested",
				"$.child_node.does_not_exist_node",
			},
			nil,
			true,
			fmt.Errorf("key error: does_not_exist_node not found in object"),
		},
		{
			"should throw error when value obtained by key is not valid json",
			args{"$.not_json",
				"$.not_exist",
			},
			nil,
			true,
			fmt.Errorf("the value specified in path `$.not_json` should be a valid JSON"),
		},
		{
			"should throw error when value obtained by key is not string",
			args{"$.not_string",
				"$.not_exist",
			},
			nil,
			true,
			errors.New("the value specified in path `$.not_string` should be of string type"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.PanicsWithError(t, tt.wantErrVal.Error(), func() {
					sr.JsonExtract(tt.args.parentJsonPath, tt.args.nestedJsonPath)
				})
				return
			}

			if got := sr.JsonExtract(tt.args.parentJsonPath, tt.args.nestedJsonPath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JsonExtract() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSymbolRegistry_HaversineDistance(t *testing.T) {

	type arguments struct {
		lat1 interface{}
		lon1 interface{}
		lat2 interface{}
		lon2 interface{}
	}
	testCases := []struct {
		desc          string
		requestJSON   []byte
		responseJSON  []byte
		args          arguments
		expectedValue interface{}
		err           error
	}{
		{
			desc: "Using single value jsonpath from request",
			requestJSON: []byte(`{
				"latitude1": -6.2446,
				"longitude1": 106.8006,
				"latitude2": -6.2655,
				"longitude2": 106.7843
			}`),
			args: arguments{
				lat1: "$.latitude1",
				lon1: "$.longitude1",
				lat2: "$.latitude2",
				lon2: "$.longitude2",
			},
			expectedValue: 2.9405665374312218,
		},
		{
			desc: "Using literal value from request",
			args: arguments{
				lat1: -6.2446,
				lon1: 106.8006,
				lat2: -6.2655,
				lon2: 106.7843,
			},
			expectedValue: 2.9405665374312218,
		},
		{
			desc: "Using multiple value jsonpath from request",
			requestJSON: []byte(`{
				"first_points": [
					{
						"latitude": -6.2446,
						"longitude": 106.8006
					},
					{
						"latitude": -6.3053,
						"longitude": 106.6435
					}
				],
				"second_points": [
					{
						"latitude": -6.2655,
						"longitude": 106.7843
					},
					{
						"latitude": -6.2856,
						"longitude": 106.7280
					}
				]
			}`),
			args: arguments{
				lat1: "$.first_points[*].latitude",
				lon1: "$.first_points[*].longitude",
				lat2: "$.second_points[*].latitude",
				lon2: "$.second_points[*].longitude",
			},
			expectedValue: []interface{}{
				2.9405665374312218,
				9.592767304287772,
			},
		},
		{
			desc: "Using series",
			args: arguments{
				lat1: series.New([]interface{}{-6.2446, -6.3053}, series.Float, "latitude"),
				lon1: series.New([]interface{}{106.8006, 106.6435}, series.Float, "longitude"),
				lat2: series.New([]interface{}{-6.2655, -6.2856}, series.Float, "latitude"),
				lon2: series.New([]interface{}{106.7843, 106.7280}, series.Float, "longitude"),
			},
			expectedValue: []interface{}{
				2.9405665374312218,
				9.592767304287772,
			},
		},
		{
			desc: "Error when first point (lat1, lon1) and second points (lat2, lon2) has different type",
			requestJSON: []byte(`{
				"latitude1": -6.2446,
				"longitude1": 106.8006,
				"second_points": [
					{
						"latitude": -6.2655,
						"longitude": 106.7843
					}
				]
			}`),
			args: arguments{
				lat1: "$.latitude1",
				lon1: "$.longitude1",
				lat2: "$.second_points[*].latitude",
				lon2: "$.second_points[*].longitude",
			},
			err: errors.New("first point and second point has different format"),
		},
		{
			desc: "Error when first points and second points has different length",
			requestJSON: []byte(`{
				"first_points": [
					{
						"latitude": -6.2446,
						"longitude": 106.8006
					}
				],
				"second_points": [
					{
						"latitude": -6.2655,
						"longitude": 106.7843
					},
					{
						"latitude": -6.2856,
						"longitude": 106.7280
					}
				]
			}`),
			args: arguments{
				lat1: "$.first_points[*].latitude",
				lon1: "$.first_points[*].longitude",
				lat2: "$.second_points[*].latitude",
				lon2: "$.second_points[*].longitude",
			},
			err: errors.New("both first point and second point arrays must have the same length"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			sr := NewRegistryWithCompiledJSONPath(jsonpath.NewStorage())
			requestJSONObj, responseJSONObj := createTestJSONObjects(tC.requestJSON, tC.responseJSON)
			if requestJSONObj != nil {
				sr.SetRawRequestJSON(requestJSONObj)
			}
			if responseJSONObj != nil {
				sr.SetModelResponseJSON(responseJSONObj)
			}

			if tC.err != nil {
				assert.PanicsWithError(t, tC.err.Error(), func() {
					sr.HaversineDistance(tC.args.lat1, tC.args.lon1, tC.args.lat2, tC.args.lon2)
				})
				return
			}
			res := sr.HaversineDistance(tC.args.lat1, tC.args.lon1, tC.args.lat2, tC.args.lon2)
			assert.Equal(t, tC.expectedValue, res)
		})
	}
}

func TestSymbolRegistry_PolarAngle(t *testing.T) {
	type arguments struct {
		lat1 interface{}
		lon1 interface{}
		lat2 interface{}
		lon2 interface{}
	}
	testCases := []struct {
		desc          string
		requestJSON   []byte
		responseJSON  []byte
		args          arguments
		expectedValue interface{}
		err           error
	}{
		{
			desc: "Using single value jsonpath from request",
			requestJSON: []byte(`{
				"latitude1": -6.2446,
				"longitude1": 106.8006,
				"latitude2": -6.2655,
				"longitude2": 106.7843
			}`),
			args: arguments{
				lat1: "$.latitude1",
				lon1: "$.longitude1",
				lat2: "$.latitude2",
				lon2: "$.longitude2",
			},
			expectedValue: -2.2302696433651628,
		},
		{
			desc: "Using single value jsonpath from request - expected 0 if distance less than 1 m",
			requestJSON: []byte(`{
				"latitude1": -6.2446,
				"longitude1": 106.8006,
				"latitude2": -6.2446,
				"longitude2": 106.8006
			}`),
			args: arguments{
				lat1: "$.latitude1",
				lon1: "$.longitude1",
				lat2: "$.latitude2",
				lon2: "$.longitude2",
			},
			expectedValue: float64(0),
		},
		{
			desc: "Using literal value from request",
			args: arguments{
				lat1: -6.2446,
				lon1: 106.8006,
				lat2: -6.2655,
				lon2: 106.7843,
			},
			expectedValue: -2.2302696433651628,
		},
		{
			desc: "Using multiple value jsonpath from request",
			requestJSON: []byte(`{
				"first_points": [
					{
						"latitude": -6.2446,
						"longitude": 106.8006
					},
					{
						"latitude": -6.3053,
						"longitude": 106.6435
					}
				],
				"second_points": [
					{
						"latitude": -6.2655,
						"longitude": 106.7843
					},
					{
						"latitude": -6.2856,
						"longitude": 106.7280
					}
				]
			}`),
			args: arguments{
				lat1: "$.first_points[*].latitude",
				lon1: "$.first_points[*].longitude",
				lat2: "$.second_points[*].latitude",
				lon2: "$.second_points[*].longitude",
			},
			expectedValue: []interface{}{
				-2.2302696433651628,
				0.2303859540944421,
			},
		},
		{
			desc: "Using series",
			args: arguments{
				lat1: series.New([]interface{}{-6.2446, -6.3053}, series.Float, "latitude"),
				lon1: series.New([]interface{}{106.8006, 106.6435}, series.Float, "longitude"),
				lat2: series.New([]interface{}{-6.2655, -6.2856}, series.Float, "latitude"),
				lon2: series.New([]interface{}{106.7843, 106.7280}, series.Float, "longitude"),
			},
			expectedValue: []interface{}{
				-2.2302696433651628,
				0.2303859540944421,
			},
		},
		{
			desc: "Error when first point (lat1, lon1) and second points (lat2, lon2) has different type",
			requestJSON: []byte(`{
				"latitude1": -6.2446,
				"longitude1": 106.8006,
				"second_points": [
					{
						"latitude": -6.2655,
						"longitude": 106.7843
					}
				]
			}`),
			args: arguments{
				lat1: "$.latitude1",
				lon1: "$.longitude1",
				lat2: "$.second_points[*].latitude",
				lon2: "$.second_points[*].longitude",
			},
			err: errors.New("first point and second point has different format"),
		},
		{
			desc: "Error when first points and second points has different length",
			requestJSON: []byte(`{
				"first_points": [
					{
						"latitude": -6.2446,
						"longitude": 106.8006
					}
				],
				"second_points": [
					{
						"latitude": -6.2655,
						"longitude": 106.7843
					},
					{
						"latitude": -6.2856,
						"longitude": 106.7280
					}
				]
			}`),
			args: arguments{
				lat1: "$.first_points[*].latitude",
				lon1: "$.first_points[*].longitude",
				lat2: "$.second_points[*].latitude",
				lon2: "$.second_points[*].longitude",
			},
			err: errors.New("both first point and second point arrays must have the same length"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			sr := NewRegistryWithCompiledJSONPath(jsonpath.NewStorage())
			requestJSONObj, responseJSONObj := createTestJSONObjects(tC.requestJSON, tC.responseJSON)
			if requestJSONObj != nil {
				sr.SetRawRequestJSON(requestJSONObj)
			}
			if responseJSONObj != nil {
				sr.SetModelResponseJSON(responseJSONObj)
			}

			if tC.err != nil {
				assert.PanicsWithError(t, tC.err.Error(), func() {
					sr.HaversineDistance(tC.args.lat1, tC.args.lon1, tC.args.lat2, tC.args.lon2)
				})
				return
			}
			res := sr.PolarAngle(tC.args.lat1, tC.args.lon1, tC.args.lat2, tC.args.lon2)
			assert.Equal(t, tC.expectedValue, res)
		})
	}
}

func TestSymbolRegistry_ParseTimestamp(t *testing.T) {
	testCases := []struct {
		desc          string
		timestamp     interface{}
		requestJSON   []byte
		responseJSON  []byte
		expectedValue interface{}
		err           error
	}{
		{
			desc:          "Using literal value (int64) for timestamp",
			timestamp:     1619541221,
			expectedValue: time.Date(2021, 4, 27, 16, 33, 41, 0, time.UTC),
		},
		{
			desc:          "Using literal value (string) for timestamp",
			timestamp:     "1619541221",
			expectedValue: time.Date(2021, 4, 27, 16, 33, 41, 0, time.UTC),
		},
		{
			desc:      "Using single value request jsonPath for timestamp",
			timestamp: "$.timestamp",
			requestJSON: []byte(`{
				"timestamp": 1619541221
			}`),
			expectedValue: time.Date(2021, 4, 27, 16, 33, 41, 0, time.UTC),
		},
		{
			desc:      "Using single value response jsonPath for timestamp",
			timestamp: "$.model_response.timestamp",
			responseJSON: []byte(`{
				"timestamp": 1619541221
			}`),
			expectedValue: time.Date(2021, 4, 27, 16, 33, 41, 0, time.UTC),
		},
		{
			desc:      "Using array from request jsonpath for timestamp",
			timestamp: "$.timestamps[*]",
			requestJSON: []byte(`{
				"timestamps": [1619541221, 1619498021]
			}`),
			expectedValue: []interface{}{
				time.Date(2021, 4, 27, 16, 33, 41, 0, time.UTC),
				time.Date(2021, 4, 27, 4, 33, 41, 0, time.UTC),
			},
		},
		{
			desc:      "Using array from request jsonpath for timestamp - different type",
			timestamp: "$.timestamps[*]",
			requestJSON: []byte(`{
				"timestamps": [1619541221, "1619498021"]
			}`),
			expectedValue: []interface{}{
				time.Date(2021, 4, 27, 16, 33, 41, 0, time.UTC),
				time.Date(2021, 4, 27, 4, 33, 41, 0, time.UTC),
			},
		},
		{
			desc:      "Using series",
			timestamp: series.New([]interface{}{1619541221, 1619498021}, series.Int, "timestamp"),
			expectedValue: []interface{}{
				time.Date(2021, 4, 27, 16, 33, 41, 0, time.UTC),
				time.Date(2021, 4, 27, 4, 33, 41, 0, time.UTC),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			sr := NewRegistryWithCompiledJSONPath(jsonpath.NewStorage())

			requestJSONObj, responseJSONObj := createTestJSONObjects(tC.requestJSON, tC.responseJSON)
			if requestJSONObj != nil {
				sr.SetRawRequestJSON(requestJSONObj)
			}
			if responseJSONObj != nil {
				sr.SetModelResponseJSON(responseJSONObj)
			}

			got := sr.ParseTimestamp(tC.timestamp)
			assert.Equal(t, tC.expectedValue, got)
		})
	}
}

func createTestJSONObjects(requestJSON, responseJSON []byte) (types.JSONObject, types.JSONObject) {
	var requestJSONObject types.JSONObject
	if requestJSON != nil {
		if err := json.Unmarshal(requestJSON, &requestJSONObject); err != nil {
			panic(err)
		}
	}
	var responseJSONObject types.JSONObject
	if responseJSON != nil {
		if err := json.Unmarshal(responseJSON, &responseJSONObject); err != nil {
			panic(err)
		}
	}
	return requestJSONObject, responseJSONObject
}

func getTestJSONObjects() (types.JSONObject, types.JSONObject) {
	var requestJSONObject types.JSONObject
	err := json.Unmarshal(requestJSONString, &requestJSONObject)
	if err != nil {
		panic(err)
	}

	var responseJSONObject types.JSONObject
	err = json.Unmarshal(responseJSONString, &responseJSONObject)
	if err != nil {
		panic(err)
	}

	return requestJSONObject, responseJSONObject
}
