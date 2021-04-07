package pipeline

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/antonmedv/expr/vm"
	"github.com/go-gota/gota/series"
	"github.com/magiconair/properties/assert"
	"github.com/mmcloughlin/geohash"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types"
)

func TestSymbolRegistry_Geohash(t *testing.T) {
	testJsonString := []byte(`{
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
	var testJsonUnmarshallled types.UnmarshalledJSON
	err := json.Unmarshal(testJsonString, &testJsonUnmarshallled)
	if err != nil {
		panic(err)
	}

	sr := SymbolRegistry{}
	env := NewEnvironment(sr, NewCompiledPipeline(make(map[string]*jsonpath.CompiledJSONPath), make(map[string]*vm.Program), nil, nil))
	env.SetSourceJSON(spec.FromJson_RAW_REQUEST, testJsonUnmarshallled)

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
				assert.Panic(t, func() {
					sr.Geohash(tt.args.latitude, tt.args.longitude, tt.args.precision)
				}, tt.wantErrVal.Error())
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
	testJsonString := []byte(`{
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
	var testJsonUnmarshallled types.UnmarshalledJSON
	err := json.Unmarshal(testJsonString, &testJsonUnmarshallled)
	if err != nil {
		panic(err)
	}

	sr := SymbolRegistry{}
	env := NewEnvironment(sr, NewCompiledPipeline(make(map[string]*jsonpath.CompiledJSONPath), make(map[string]*vm.Program), nil, nil))
	env.SetSourceJSON(spec.FromJson_RAW_REQUEST, testJsonUnmarshallled)

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
				assert.Panic(t, func() {
					sr.S2ID(tt.args.latitude, tt.args.longitude, tt.args.level)
				}, tt.wantErrVal.Error())
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
	var testJsonUnmarshallled types.UnmarshalledJSON
	err := json.Unmarshal(testJsonString, &testJsonUnmarshallled)
	if err != nil {
		panic(err)
	}

	sr := SymbolRegistry{}
	env := NewEnvironment(sr, NewCompiledPipeline(make(map[string]*jsonpath.CompiledJSONPath), make(map[string]*vm.Program), nil, nil))
	env.SetSourceJSON(spec.FromJson_RAW_REQUEST, testJsonUnmarshallled)

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
			fmt.Errorf("the value specified in path `\\$.not_json` should be a valid JSON"),
		},
		{
			"should throw error when value obtained by key is not string",
			args{"$.not_string",
				"$.not_exist",
			},
			nil,
			true,
			errors.New("the value specified in path `\\$.not_string` should be of string type"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.Panic(t, func() {
					sr.JsonExtract(tt.args.parentJsonPath, tt.args.nestedJsonPath)
				}, tt.wantErrVal.Error())
				return
			}

			if got := sr.JsonExtract(tt.args.parentJsonPath, tt.args.nestedJsonPath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JsonExtract() = %v, want %v", got, tt.want)
			}
		})
	}
}
