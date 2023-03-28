package symbol

import (
	"encoding/json"

	"github.com/caraml-dev/merlin/pkg/transformer/types"
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
		"longitudeLongArrays": [1.0, 2.0, 3.0],
		"originGeohash": "bcd3u",
		"destinationGeohash": "bc83n"
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
