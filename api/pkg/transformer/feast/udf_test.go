package feast

import (
	"encoding/json"
	"errors"
	"github.com/mmcloughlin/geohash"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExtractGeohash(t *testing.T) {
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
		"longitudeInteger": 1,
		"longitudeArrays": [1.0, 2.0]
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
			longitudeJsonPath: "$.longitudeInteger",
			precision:         12,
			expValue:          geohash.Encode(1.0, 1.0),
		},
		{
			name:              "type conversion error",
			latitudeJsonPath:  "$.latitudeWrongType",
			longitudeJsonPath: "$.longitude",
			expError:          errors.New("strconv.ParseFloat: parsing \"abcde\": invalid syntax"),
		},
		{
			name:              "latitude arrays and single longitude",
			latitudeJsonPath:  "$.latitudeArrays",
			longitudeJsonPath: "$.longitude",
			precision:         12,
			expValue: []string{
				geohash.Encode(1.0, 2.0),
				geohash.Encode(2.0, 2.0),
			},
		},
		{
			name:              "latitude and longitude arrays",
			latitudeJsonPath:  "$.latitudeArrays",
			longitudeJsonPath: "$.longitudeArrays",
			precision:         12,
			expValue: []string{
				geohash.Encode(1.0, 1.0),
				geohash.Encode(1.0, 2.0),
				geohash.Encode(2.0, 1.0),
				geohash.Encode(2.0, 2.0),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := extractGeohash(testJsonUnmarshallled, test.latitudeJsonPath, test.longitudeJsonPath, test.precision)
			if err != nil {
				if test.expError != nil {
					assert.EqualError(t, err, test.expError.Error())
					return
				} else {
					assert.Fail(t, err.Error())
				}
			}
			assert.Equal(t, test.expValue, actual)
		})
	}
}
