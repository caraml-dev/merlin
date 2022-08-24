package symbol

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/symbol/function"
	"github.com/gojek/merlin/pkg/transformer/types/series"
	"github.com/mmcloughlin/geohash"
	"github.com/stretchr/testify/assert"
)

func TestSymbolRegistry_Geohash(t *testing.T) {
	requestJSONObject, responseJSONObject := getTestJSONObjects()

	sr := NewRegistryWithCompiledJSONPath(jsonpath.NewStorage())
	sr.SetRawRequest(requestJSONObject)
	sr.SetModelResponse(responseJSONObject)

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
	sr.SetRawRequest(requestJSONObject)
	sr.SetModelResponse(responseJSONObject)

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
				sr.SetRawRequest(requestJSONObj)
			}
			if responseJSONObj != nil {
				sr.SetModelResponse(responseJSONObj)
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

func TestSymbolRegistry_HaversineDistanceWithUnit(t *testing.T) {
	type arguments struct {
		lat1         interface{}
		lon1         interface{}
		lat2         interface{}
		lon2         interface{}
		distanceUnit string
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
				lat1:         "$.latitude1",
				lon1:         "$.longitude1",
				lat2:         "$.latitude2",
				lon2:         "$.longitude2",
				distanceUnit: function.KMDistanceUnit,
			},
			expectedValue: 2.9405665374312218,
		},
		{
			desc: "Using literal value from request",
			args: arguments{
				lat1:         -6.2446,
				lon1:         106.8006,
				lat2:         -6.2655,
				lon2:         106.7843,
				distanceUnit: function.MeterDistanceUnit,
			},
			expectedValue: 2940.5665374312218,
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
				lat1:         "$.first_points[*].latitude",
				lon1:         "$.first_points[*].longitude",
				lat2:         "$.second_points[*].latitude",
				lon2:         "$.second_points[*].longitude",
				distanceUnit: function.KMDistanceUnit,
			},

			expectedValue: []interface{}{
				2.9405665374312218,
				9.592767304287772,
			},
		},
		{
			desc: "Using series",
			args: arguments{
				lat1:         series.New([]interface{}{-6.2446, -6.3053}, series.Float, "latitude"),
				lon1:         series.New([]interface{}{106.8006, 106.6435}, series.Float, "longitude"),
				lat2:         series.New([]interface{}{-6.2655, -6.2856}, series.Float, "latitude"),
				lon2:         series.New([]interface{}{106.7843, 106.7280}, series.Float, "longitude"),
				distanceUnit: function.MeterDistanceUnit,
			},

			expectedValue: []interface{}{
				2940.5665374312218,
				9592.767304287772,
			},
		},
		{
			desc: "Using series; distance unit not valid / supported fallback to KM",
			args: arguments{
				lat1:         series.New([]interface{}{-6.2446, -6.3053}, series.Float, "latitude"),
				lon1:         series.New([]interface{}{106.8006, 106.6435}, series.Float, "longitude"),
				lat2:         series.New([]interface{}{-6.2655, -6.2856}, series.Float, "latitude"),
				lon2:         series.New([]interface{}{106.7843, 106.7280}, series.Float, "longitude"),
				distanceUnit: "mile",
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
				lat1:         "$.latitude1",
				lon1:         "$.longitude1",
				lat2:         "$.second_points[*].latitude",
				lon2:         "$.second_points[*].longitude",
				distanceUnit: function.KMDistanceUnit,
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
				lat1:         "$.first_points[*].latitude",
				lon1:         "$.first_points[*].longitude",
				lat2:         "$.second_points[*].latitude",
				lon2:         "$.second_points[*].longitude",
				distanceUnit: function.KMDistanceUnit,
			},
			err: errors.New("both first point and second point arrays must have the same length"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			sr := NewRegistryWithCompiledJSONPath(jsonpath.NewStorage())
			requestJSONObj, responseJSONObj := createTestJSONObjects(tC.requestJSON, tC.responseJSON)
			if requestJSONObj != nil {
				sr.SetRawRequest(requestJSONObj)
			}
			if responseJSONObj != nil {
				sr.SetModelResponse(responseJSONObj)
			}

			if tC.err != nil {
				assert.PanicsWithError(t, tC.err.Error(), func() {
					sr.HaversineDistanceWithUnit(tC.args.lat1, tC.args.lon1, tC.args.lat2, tC.args.lon2, tC.args.distanceUnit)
				})
				return
			}
			res := sr.HaversineDistanceWithUnit(tC.args.lat1, tC.args.lon1, tC.args.lat2, tC.args.lon2, tC.args.distanceUnit)
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
				sr.SetRawRequest(requestJSONObj)
			}
			if responseJSONObj != nil {
				sr.SetModelResponse(responseJSONObj)
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

func TestRegistry_GeohashDistance(t *testing.T) {
	requestJSONObject, responseJSONObject := getTestJSONObjects()

	sr := NewRegistryWithCompiledJSONPath(jsonpath.NewStorage())
	sr.SetRawRequest(requestJSONObject)
	sr.SetModelResponse(responseJSONObject)

	type args struct {
		firstGeohash  interface{}
		secondGeohash interface{}
		distanceUnit  string
	}
	tests := []struct {
		name string
		sr   Registry
		args args
		err  error
		want interface{}
	}{
		{
			name: "geohash from jsonpath",
			sr:   sr,
			args: args{
				firstGeohash:  "$.originGeohash",
				secondGeohash: "$.destinationGeohash",
				distanceUnit:  function.MeterDistanceUnit,
			},
			want: 179939.84348353752,
		},
		{
			name: "using series",
			sr:   sr,
			args: args{
				firstGeohash:  series.New([]string{"qqgggnwxx", "qqgggnweb"}, series.String, ""),
				secondGeohash: series.New([]string{"qqguh19bs", "qqguh19b7"}, series.String, ""),
				distanceUnit:  function.KMDistanceUnit,
			},
			want: []interface{}{4.45796024811067, 4.510014050192708},
		},
		{
			name: "using series, mismatch length",
			sr:   sr,
			args: args{
				firstGeohash:  series.New([]string{"qqgggnwxx", "qqgggnweb"}, series.String, ""),
				secondGeohash: series.New([]string{"qqguh19bs", "qqguh19b7", "qqguh19b8"}, series.String, ""),
				distanceUnit:  function.KMDistanceUnit,
			},
			err: fmt.Errorf("both first point and second point arrays must have the same length"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err != nil {
				assert.PanicsWithError(t, tt.err.Error(), func() {
					tt.sr.GeohashDistance(tt.args.firstGeohash, tt.args.secondGeohash, tt.args.distanceUnit)
				})
				return
			}
			if got := tt.sr.GeohashDistance(tt.args.firstGeohash, tt.args.secondGeohash, tt.args.distanceUnit); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.GeohashDistance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_GeohashAllNeighbors(t *testing.T) {
	requestJSONObject, responseJSONObject := getTestJSONObjects()

	sr := NewRegistryWithCompiledJSONPath(jsonpath.NewStorage())
	sr.SetRawRequest(requestJSONObject)
	sr.SetModelResponse(responseJSONObject)

	type args struct {
		targetGeohash interface{}
	}
	tests := []struct {
		name string
		sr   Registry
		args args
		want interface{}
		err  error
	}{
		{
			name: "from single value",
			sr:   sr,
			args: args{
				targetGeohash: "qqgggnwxx",
			},
			want: []string{"qqgggnwxz", "qqgggnwzb", "qqgggnwz8", "qqgggnwz2", "qqgggnwxr", "qqgggnwxq", "qqgggnwxw", "qqgggnwxy"},
		},
		{
			name: "from single jsonpath",
			sr:   sr,
			args: args{
				targetGeohash: "$.originGeohash",
			},
			want: []string{"bcd6h", "bcd6j", "bcd3v", "bcd3t", "bcd3s", "bcd3e", "bcd3g", "bcd65"},
		},
		{
			name: "from series",
			sr:   sr,
			args: args{
				targetGeohash: series.New([]string{"qqgggnwxx", "bcd3u"}, series.String, ""),
			},
			want: [][]string{{"qqgggnwxz", "qqgggnwzb", "qqgggnwz8", "qqgggnwz2", "qqgggnwxr", "qqgggnwxq", "qqgggnwxw", "qqgggnwxy"}, {"bcd6h", "bcd6j", "bcd3v", "bcd3t", "bcd3s", "bcd3e", "bcd3g", "bcd65"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sr.GeohashAllNeighbors(tt.args.targetGeohash); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.GeohashAllNeighbors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_GeohashNeighborForDirection(t *testing.T) {
	requestJSONObject, responseJSONObject := getTestJSONObjects()

	sr := NewRegistryWithCompiledJSONPath(jsonpath.NewStorage())
	sr.SetRawRequest(requestJSONObject)
	sr.SetModelResponse(responseJSONObject)

	type args struct {
		targetGeohash interface{}
		direction     string
	}
	tests := []struct {
		name string
		sr   Registry
		args args
		want interface{}
		err  error
	}{
		{
			name: "from single value",
			sr:   sr,
			args: args{
				targetGeohash: "qqgggnwxx",
				direction:     "north",
			},
			want: "qqgggnwxz",
		},
		{
			name: "from single value, direction in upper case",
			sr:   sr,
			args: args{
				targetGeohash: "qqgggnwxx",
				direction:     "NORTH",
			},
			want: "qqgggnwxz",
		},
		{
			name: "from single value, direction is not valid",
			sr:   sr,
			args: args{
				targetGeohash: "qqgggnwxx",
				direction:     "not sure",
			},
			err: fmt.Errorf("direction 'not sure' is not valid"),
		},
		{
			name: "from jsonpath",
			sr:   sr,
			args: args{
				targetGeohash: "$.originGeohash",
				direction:     "south",
			},
			want: "bcd3s",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err != nil {
				assert.PanicsWithError(t, tt.err.Error(), func() {
					tt.sr.GeohashNeighborForDirection(tt.args.targetGeohash, tt.args.direction)
				})
				return
			}
			if got := tt.sr.GeohashNeighborForDirection(tt.args.targetGeohash, tt.args.direction); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Registry.GeohashNeighborForDirection() = %v, want %v", got, tt.want)
			}
		})
	}
}
