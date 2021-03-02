package feast

import (
	"encoding/json"
	"errors"
	"github.com/mmcloughlin/geohash"
	"github.com/stretchr/testify/assert"
	"fmt"
	"testing"
)

func TestGeohash(t *testing.T) {
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
	var testJsonUnmarshallled interface{}
	err := json.Unmarshal(testJsonString, &testJsonUnmarshallled)
	if err != nil {
		panic(err)
	}

	tests := []struct {
		name              string
		latitudeJsonPath  string
		longitudeJsonPath string
		precision         uint
		expValue          interface{}
		expError          error
	}{
		{
			name:              "geohash from json fields",
			latitudeJsonPath:  "$.latitude",
			longitudeJsonPath: "$.longitude",
			precision:         7,
			expValue:          geohash.EncodeWithPrecision(1.0, 2.0, 7),
		},
		{
			name:              "geohash from json struct",
			latitudeJsonPath:  "$.location.latitude",
			longitudeJsonPath: "$.location.longitude",
			precision:         7,
			expValue:          geohash.EncodeWithPrecision(0.1, 2.0, 7),
		},
		{
			name:              "type conversion for latitude and longitude input",
			latitudeJsonPath:  "$.latitudeString",
			longitudeJsonPath: "$.longitudeString",
			precision:         12,
			expValue:          geohash.Encode(1.0, 2.0),
		},
		{
			name:              "Type difference error",
			latitudeJsonPath:  "$.latitude",
			longitudeJsonPath: "$.longitudeArrays",
			expError:          errors.New("latitude and longitude must have the same types"),
		},
		{
			name:              "type conversion error",
			latitudeJsonPath:  "$.latitudeWrongType",
			longitudeJsonPath: "$.longitudeString",
			expError:          errors.New("strconv.ParseFloat: parsing \"abcde\": invalid syntax"),
		},
		{
			name:              "array length difference error",
			latitudeJsonPath:  "$.latitudeArrays",
			longitudeJsonPath: "$.longitudeLongArrays",
			precision:         12,
			expError:          errors.New("both latitude and longitude arrays must have the same length"),
		},
		{
			name:              "latitude and longitude arrays",
			latitudeJsonPath:  "$.latitudeArrays",
			longitudeJsonPath: "$.longitudeArrays",
			precision:         12,
			expValue: []string{
				geohash.Encode(1.0, 1.0),
				geohash.Encode(2.0, 2.0),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			udf := &UdfEnv{UnmarshalledJsonRequest:testJsonUnmarshallled}
			udfResult := udf.Geohash(test.latitudeJsonPath, test.longitudeJsonPath, test.precision)
			if udfResult.Error != nil {
				if test.expError != nil {
					assert.EqualError(t, udfResult.Error, test.expError.Error())
					return
				} else {
					assert.Fail(t, udfResult.Error.Error())
				}
			}
			assert.Equal(t, test.expValue, udfResult.Value)
		})
	}
}

func TestS2ID(t *testing.T) {
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
	var testJsonUnmarshallled interface{}
	err := json.Unmarshal(testJsonString, &testJsonUnmarshallled)
	if err != nil {
		panic(err)
	}

	tests := []struct {
		name              string
		latitudeJsonPath  string
		longitudeJsonPath string
		level             int
		expValue          interface{}
		expError          error
	}{
		{
			name:              "geohash from json fields",
			latitudeJsonPath:  "$.latitude",
			longitudeJsonPath: "$.longitude",
			level:             7,
			expValue:          "1154680723211288576",
		},
		{
			name:              "geohash from json struct",
			latitudeJsonPath:  "$.location.latitude",
			longitudeJsonPath: "$.location.longitude",
			level:             7,
			expValue:          "1155102935676354560",
		},
		{
			name:              "type conversion for latitude and longitude input",
			latitudeJsonPath:  "$.latitudeString",
			longitudeJsonPath: "$.longitudeString",
			level:             12,
			expValue:          "1154732743855177728",
		},
		{
			name:              "Type difference error",
			latitudeJsonPath:  "$.latitude",
			longitudeJsonPath: "$.longitudeArrays",
			expError:          errors.New("latitude and longitude must have the same types"),
		},
		{
			name:              "type conversion error",
			latitudeJsonPath:  "$.latitudeWrongType",
			longitudeJsonPath: "$.longitudeString",
			expError:          errors.New("strconv.ParseFloat: parsing \"abcde\": invalid syntax"),
		},
		{
			name:              "array length difference error",
			latitudeJsonPath:  "$.latitudeArrays",
			longitudeJsonPath: "$.longitudeLongArrays",
			level:             12,
			expError:          errors.New("both latitude and longitude arrays must have the same length"),
		},
		{
			name:              "latitude and longitude arrays",
			latitudeJsonPath:  "$.latitudeArrays",
			longitudeJsonPath: "$.longitudeArrays",
			level:             12,
			expValue: []string{
				"1153277815093723136",
				"1154346540395921408",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			udf := &UdfEnv{UnmarshalledJsonRequest:testJsonUnmarshallled}
			udfResult := udf.S2ID(test.latitudeJsonPath, test.longitudeJsonPath, test.level)
			if udfResult.Error != nil {
				if test.expError != nil {
					assert.EqualError(t, udfResult.Error, test.expError.Error())
					return
				} else {
					assert.Fail(t, udfResult.Error.Error())
				}
			}
			assert.Equal(t, test.expValue, udfResult.Value)
		})
	}
}

func TestJsonExtract(t *testing.T) {
	testJsonString := []byte(`{
		"details": "{\"merchant_id\": 9001}",
		"nested": "{\"child_node\": { \"grandchild_node\": \"gen-z\"}}",
		"not_json": "i_am_not_json_string",
		"not_string": 1024,
		"array": "{\"child_node\": { \"array\": [1, 2]}}"
	}`)
	var testJsonUnmarshallled interface{}
	err := json.Unmarshal(testJsonString, &testJsonUnmarshallled)
	if err != nil {
		panic(err)
	}

	tests := []struct {
		name           string
		keyJsonPath    string
		nestedJsonPath string
		extractedValue interface{}
		expError       error
	}{
		{
			name:           "should be able to extract value from JSON string",
			keyJsonPath:    "$.details",
			nestedJsonPath: "$.merchant_id",
			extractedValue: float64(9001),
		},
		{
			name:           "should be able to extract value using nested key from JSON string",
			keyJsonPath:    "$.nested",
			nestedJsonPath: "$.child_node.grandchild_node",
			extractedValue: "gen-z",
		},
		{
			name:           "should be able to extract array value using nested key from JSON string",
			keyJsonPath:    "$.array",
			nestedJsonPath: "$.child_node.array[*]",
			extractedValue: []interface {}{float64(1), float64(2)},
		},
		{
			name:           "should throw error when value specified by key does not exist in nested JSON",
			keyJsonPath:    "$.nested",
			nestedJsonPath: "$.child_node.does_not_exist_node",
			expError: fmt.Errorf("key error: does_not_exist_node not found in object"),
		},
		{
			name:           "should throw error when value obtained by key is not valid json",
			keyJsonPath:    "$.not_json",
			nestedJsonPath: "$.not_exist",
			expError: fmt.Errorf("the value specified in path `$.not_json` should be a valid JSON"),
		},
		{
			name:           "should throw error when value obtained by key is not string",
			keyJsonPath:    "$.not_string",
			nestedJsonPath: "$.not_exist",
			expError: fmt.Errorf("the value specified in path `$.not_string` should be of string type"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := jsonExtract(testJsonUnmarshallled, test.keyJsonPath, test.nestedJsonPath)
			if err != nil {
				if test.expError != nil {
					assert.EqualError(t, err, test.expError.Error())
					return
				} else {
					assert.Fail(t, err.Error())
				}
			}
			assert.Equal(t, test.extractedValue, actual)
		})
	}
}