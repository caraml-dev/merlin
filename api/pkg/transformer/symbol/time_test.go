package symbol

import (
	"testing"
	"time"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/types/series"
	"github.com/stretchr/testify/assert"
)

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

func TestSymbolRegistry_IsWeekend(t *testing.T) {
	testCases := []struct {
		desc          string
		timestamp     interface{}
		timezone      string
		requestJSON   []byte
		responseJSON  []byte
		expectedValue interface{}
	}{
		{
			desc:          "single value timestamp, if timezone not specified using UTC by default (Saturday, November 20, 2021 9:50:44 PM GMT)",
			timestamp:     1637445044,
			timezone:      "",
			expectedValue: 1,
		},
		{
			desc:          "single value timestamp, falls in weekend (Sunday, November 21, 2021 4:50:44 AM GMT+07:00)",
			timestamp:     1637445044,
			timezone:      "Asia/Jakarta",
			expectedValue: 1,
		},
		{
			desc:      "multiple value timestamps",
			timestamp: series.New([]interface{}{1637441444, 1637527844}, series.Int, "timestamp"),
			timezone:  "Asia/Jakarta",
			expectedValue: []interface{}{
				1, 0,
			},
		},
		{
			desc:      "multiple value timestamps using UTC",
			timestamp: series.New([]interface{}{1637441444, 1637527844}, series.Int, "timestamp"),
			timezone:  "UTC",
			expectedValue: []interface{}{
				1, 1,
			},
		},
		{
			desc:      "Using single value request jsonPath for timestamp",
			timestamp: "$.timestamp",
			requestJSON: []byte(`{
				"timestamp": 1637445044
			}`),
			timezone:      "UTC",
			expectedValue: 1,
		},
		{
			desc:      "Using multiple value request jsonPath for timestamp",
			timestamp: "$.timestamps",
			requestJSON: []byte(`{
				"timestamps": [1637441444,1637527844]
			}`),
			timezone: "Asia/Jakarta",
			expectedValue: []interface{}{
				1, 0,
			},
		},
		{
			desc:      "Using single value response jsonPath for timestamp",
			timestamp: "$.model_response.timestamp",
			responseJSON: []byte(`{
				"timestamp": 1637445044
			}`),
			timezone:      "UTC",
			expectedValue: 1,
		},
		{
			desc:      "Using multiple value response jsonPath for timestamp",
			timestamp: "$.model_response.timestamps",
			responseJSON: []byte(`{
				"timestamps": [1637441444,1637527844]
			}`),
			timezone: "Asia/Jakarta",
			expectedValue: []interface{}{
				1, 0,
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

			got := sr.IsWeekend(tC.timestamp, tC.timezone)
			assert.Equal(t, tC.expectedValue, got)
		})
	}
}

func TestSymbolRegistry_FormatTimestamp(t *testing.T) {
	testCases := []struct {
		desc          string
		timestamp     interface{}
		format        string
		timezone      string
		requestJSON   []byte
		responseJSON  []byte
		expectedValue interface{}
	}{
		{
			desc:          "Using ANSIC Format",
			timestamp:     1637691859,
			format:        time.ANSIC,
			timezone:      "Asia/Jakarta",
			expectedValue: "Wed Nov 24 01:24:19 2021",
		},
		{
			desc:          "Using RFC1123 Format",
			timestamp:     1637691859,
			format:        time.RFC1123,
			timezone:      "Asia/Jakarta",
			expectedValue: "Wed, 24 Nov 2021 01:24:19 WIB",
		},
		{
			desc:          "Using RFC1123Z Format",
			timestamp:     1637691859,
			format:        time.RFC1123Z,
			timezone:      "Asia/Jakarta",
			expectedValue: "Wed, 24 Nov 2021 01:24:19 +0700",
		},
		{
			desc:          "Using custom Format",
			timestamp:     1637691859,
			format:        "2006-01-02",
			timezone:      "Asia/Jakarta",
			expectedValue: "2021-11-24",
		},
		{
			desc:      "Multiple timestamps from request jsonpath",
			timestamp: "$.timestamps",
			requestJSON: []byte(`{
				"timestamps":[1637441444,1637527844]
			}`),
			format:   "2006-01-02",
			timezone: "Asia/Jakarta",
			expectedValue: []interface{}{
				"2021-11-21", "2021-11-22",
			},
		},
		{
			desc:      "Multiple timestamps from response jsonpath",
			timestamp: "$.model_response.timestamps",
			responseJSON: []byte(`{
				"timestamps":[1637441444,1637527844]
			}`),
			format:   "2006-01-02",
			timezone: "Asia/Jakarta",
			expectedValue: []interface{}{
				"2021-11-21", "2021-11-22",
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

			got := sr.FormatTimestamp(tC.timestamp, tC.timezone, tC.format)
			assert.Equal(t, tC.expectedValue, got)
		})
	}
}

func TestSymbolRegistry_DayOfWeek(t *testing.T) {
	testCases := []struct {
		desc          string
		timestamp     interface{}
		timezone      string
		requestJSON   []byte
		responseJSON  []byte
		expectedValue interface{}
	}{
		{
			desc:          "single value timestamp (Monday UTC)",
			timestamp:     1637605459,
			timezone:      "",
			expectedValue: 1,
		},
		{
			desc:          "single value timestamp (Tuesday in Asia/Jakarta)",
			timestamp:     1637605459,
			timezone:      "Asia/Jakarta",
			expectedValue: 2,
		},
		{
			desc:      "multiple value timestamps",
			timestamp: series.New([]interface{}{1637441444, 1637691859}, series.Int, "timestamp"),
			timezone:  "Asia/Jakarta",
			expectedValue: []interface{}{
				0, 3,
			},
		},
		{
			desc:      "multiple value timestamps using UTC",
			timestamp: series.New([]interface{}{1637441444, 1637691859}, series.Int, "timestamp"),
			timezone:  "UTC",
			expectedValue: []interface{}{
				6, 2,
			},
		},
		{
			desc:      "Using single value request jsonPath for timestamp",
			timestamp: "$.timestamp",
			requestJSON: []byte(`{
				"timestamp": 1637445044
			}`),
			timezone:      "UTC",
			expectedValue: 6,
		},
		{
			desc:      "Using multiple value request jsonPath for timestamp",
			timestamp: "$.timestamps",
			requestJSON: []byte(`{
				"timestamps": [1637441444,1637691859]
			}`),
			timezone: "Asia/Jakarta",
			expectedValue: []interface{}{
				0, 3,
			},
		},
		{
			desc:      "Using single value response jsonPath for timestamp",
			timestamp: "$.model_response.timestamp",
			responseJSON: []byte(`{
				"timestamp": 1637445044
			}`),
			timezone:      "UTC",
			expectedValue: 6,
		},
		{
			desc:      "Using multiple value response jsonPath for timestamp",
			timestamp: "$.model_response.timestamps",
			responseJSON: []byte(`{
				"timestamps": [1637441444,1637691859]
			}`),
			timezone: "Asia/Jakarta",
			expectedValue: []interface{}{
				0, 3,
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

			got := sr.DayOfWeek(tC.timestamp, tC.timezone)
			assert.Equal(t, tC.expectedValue, got)
		})
	}
}

func TestRegistry_ParseDateTime(t *testing.T) {
	jakartaLocation, _ := time.LoadLocation("Asia/Jakarta")
	jayapuraLocation, _ := time.LoadLocation("Asia/Jayapura")
	singaporeLocation, _ := time.LoadLocation("Asia/Singapore")

	testCases := []struct {
		desc             string
		dateTime         interface{}
		dateTimeFormat   string
		dateTimeLocation string
		requestJSON      []byte
		responseJSON     []byte
		expectedValue    interface{}
		err              error
	}{
		// No TZ
		{
			desc:             "No TZ - Using literal value (string) for datetime",
			dateTime:         "2021-11-30 15:00:00",
			dateTimeFormat:   "2006-01-02 15:04:05",
			dateTimeLocation: "",
			expectedValue:    time.Date(2021, 11, 30, 15, 00, 00, 0, time.UTC),
		},
		{
			desc:             "No TZ - Using single value request jsonPath for datetime",
			dateTime:         "$.datetime",
			dateTimeFormat:   "2006-01-02 15:04:05",
			dateTimeLocation: "",
			requestJSON: []byte(`{
				"datetime": "2021-11-30 15:00:00"
			}`),
			expectedValue: time.Date(2021, 11, 30, 15, 00, 00, 0, time.UTC),
		},
		{
			desc:             "No TZ - Using single value response jsonPath for datetime",
			dateTime:         "$.model_response.datetime",
			dateTimeFormat:   "2006-01-02 15:04:05",
			dateTimeLocation: "",
			responseJSON: []byte(`{
				"datetime": "2021-11-30 15:00:00"
			}`),
			expectedValue: time.Date(2021, 11, 30, 15, 00, 00, 0, time.UTC),
		},
		{
			desc:             "No TZ - Using array from request jsonpath for datetime",
			dateTime:         "$.datetimes[*]",
			dateTimeFormat:   "2006-01-02 15:04:05",
			dateTimeLocation: "",
			requestJSON: []byte(`{
				"datetimes": ["2021-11-30 15:00:00", "2021-11-30 16:00:00"]
			}`),
			expectedValue: []interface{}{
				time.Date(2021, 11, 30, 15, 00, 00, 0, time.UTC),
				time.Date(2021, 11, 30, 16, 00, 00, 0, time.UTC),
			},
		},
		{
			desc:             "No TZ - Using series",
			dateTime:         series.New([]interface{}{"2021-11-30 15:00:00", "2021-11-30 16:00:00"}, series.String, "datetime"),
			dateTimeFormat:   "2006-01-02 15:04:05",
			dateTimeLocation: "",
			expectedValue: []interface{}{
				time.Date(2021, 11, 30, 15, 00, 00, 0, time.UTC),
				time.Date(2021, 11, 30, 16, 00, 00, 0, time.UTC),
			},
		},
		// No TZ in datetime, but location
		{
			desc:             "No TZ + Location - Using literal value (string) for datetime",
			dateTime:         "2021-11-30 15:00:00",
			dateTimeFormat:   "2006-01-02 15:04:05",
			dateTimeLocation: "Asia/Jayapura",
			expectedValue:    time.Date(2021, 11, 30, 15, 00, 00, 0, jayapuraLocation),
		},
		{
			desc:             "No TZ + Location - Using single value request jsonPath for datetime",
			dateTime:         "$.datetime",
			dateTimeFormat:   "2006-01-02 15:04:05",
			dateTimeLocation: "Asia/Jayapura",
			requestJSON: []byte(`{
				"datetime": "2021-11-30 15:00:00"
			}`),
			expectedValue: time.Date(2021, 11, 30, 15, 00, 00, 0, jayapuraLocation),
		},
		{
			desc:             "No TZ + Location - Using single value response jsonPath for datetime",
			dateTime:         "$.model_response.datetime",
			dateTimeFormat:   "2006-01-02 15:04:05",
			dateTimeLocation: "Asia/Jayapura",
			responseJSON: []byte(`{
				"datetime": "2021-11-30 15:00:00"
			}`),
			expectedValue: time.Date(2021, 11, 30, 15, 00, 00, 0, jayapuraLocation),
		},
		{
			desc:             "No TZ + Location - Using array from request jsonpath for datetime",
			dateTime:         "$.datetimes[*]",
			dateTimeFormat:   "2006-01-02 15:04:05",
			dateTimeLocation: "Asia/Jayapura",
			requestJSON: []byte(`{
				"datetimes": ["2021-11-30 15:00:00", "2021-11-30 16:00:00"]
			}`),
			expectedValue: []interface{}{
				time.Date(2021, 11, 30, 15, 00, 00, 0, jayapuraLocation),
				time.Date(2021, 11, 30, 16, 00, 00, 0, jayapuraLocation),
			},
		},
		{
			desc:             "No TZ + Location - Using series",
			dateTime:         series.New([]interface{}{"2021-11-30 15:00:00", "2021-11-30 16:00:00"}, series.String, "datetime"),
			dateTimeFormat:   "2006-01-02 15:04:05",
			dateTimeLocation: "Asia/Jayapura",
			expectedValue: []interface{}{
				time.Date(2021, 11, 30, 15, 00, 00, 0, jayapuraLocation),
				time.Date(2021, 11, 30, 16, 00, 00, 0, jayapuraLocation),
			},
		},
		// UTC+7 Asia/Jakarta
		{
			desc:             "UTC+7 Asia/Jakarta - Using literal value (string) for datetime",
			dateTime:         "2021-11-30T15:00:00+07:00",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			dateTimeLocation: "Asia/Jakarta",
			expectedValue:    time.Date(2021, 11, 30, 15, 00, 00, 0, jakartaLocation),
		},
		{
			desc:             "UTC+7 Asia/Jakarta - Using single value request jsonPath for datetime",
			dateTime:         "$.datetime",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			dateTimeLocation: "Asia/Jakarta",
			requestJSON: []byte(`{
				"datetime": "2021-11-30T15:00:00+07:00"
			}`),
			expectedValue: time.Date(2021, 11, 30, 15, 00, 00, 0, jakartaLocation),
		},
		{
			desc:             "UTC+7 Asia/Jakarta - Using single value response jsonPath for datetime",
			dateTime:         "$.model_response.datetime",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			dateTimeLocation: "Asia/Jakarta",
			responseJSON: []byte(`{
				"datetime": "2021-11-30T15:00:00+07:00"
			}`),
			expectedValue: time.Date(2021, 11, 30, 15, 00, 00, 0, jakartaLocation),
		},
		{
			desc:             "UTC+7 Asia/Jakarta - Using array from request jsonpath for datetime",
			dateTime:         "$.datetimes[*]",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			dateTimeLocation: "Asia/Jakarta",
			requestJSON: []byte(`{
				"datetimes": ["2021-11-30T15:00:00+07:00", "2021-11-30T16:00:00+07:00"]
			}`),
			expectedValue: []interface{}{
				time.Date(2021, 11, 30, 15, 00, 00, 0, jakartaLocation),
				time.Date(2021, 11, 30, 16, 00, 00, 0, jakartaLocation),
			},
		},
		{
			desc:             "UTC+7 Asia/Jakarta - Using series",
			dateTime:         series.New([]interface{}{"2021-11-30T15:00:00+07:00", "2021-11-30T16:00:00+07:00"}, series.String, "datetime"),
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			dateTimeLocation: "Asia/Jakarta",
			expectedValue: []interface{}{
				time.Date(2021, 11, 30, 15, 00, 00, 0, jakartaLocation),
				time.Date(2021, 11, 30, 16, 00, 00, 0, jakartaLocation),
			},
		},
		// UTC+8 Asia/Singapore
		{
			desc:             "UTC+8 Asia/Singapore - Using literal value (string) for datetime",
			dateTime:         "2021-11-30T15:00:00+08:00",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			dateTimeLocation: "Asia/Singapore",
			expectedValue:    time.Date(2021, 11, 30, 15, 00, 00, 0, singaporeLocation),
		},
		{
			desc:             "UTC+8 Asia/Singapore - Using single value request jsonPath for datetime",
			dateTime:         "$.datetime",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			dateTimeLocation: "Asia/Singapore",
			requestJSON: []byte(`{
				"datetime": "2021-11-30T15:00:00+08:00"
			}`),
			expectedValue: time.Date(2021, 11, 30, 15, 00, 00, 0, singaporeLocation),
		},
		{
			desc:             "UTC+8 Asia/Singapore - Using single value response jsonPath for datetime",
			dateTime:         "$.model_response.datetime",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			dateTimeLocation: "Asia/Singapore",
			responseJSON: []byte(`{
				"datetime": "2021-11-30T15:00:00+08:00"
			}`),
			expectedValue: time.Date(2021, 11, 30, 15, 00, 00, 0, singaporeLocation),
		},
		{
			desc:             "UTC+8 Asia/Singapore - Using array from request jsonpath for datetime",
			dateTime:         "$.datetimes[*]",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			dateTimeLocation: "Asia/Singapore",
			requestJSON: []byte(`{
				"datetimes": ["2021-11-30T15:00:00+08:00", "2021-11-30T16:00:00+08:00"]
			}`),
			expectedValue: []interface{}{
				time.Date(2021, 11, 30, 15, 00, 00, 0, singaporeLocation),
				time.Date(2021, 11, 30, 16, 00, 00, 0, singaporeLocation),
			},
		},
		{
			desc:             "UTC+8 Asia/Singapore - Using series",
			dateTime:         series.New([]interface{}{"2021-11-30T15:00:00+08:00", "2021-11-30T16:00:00+08:00"}, series.String, "datetime"),
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			dateTimeLocation: "Asia/Singapore",
			expectedValue: []interface{}{
				time.Date(2021, 11, 30, 15, 00, 00, 0, singaporeLocation),
				time.Date(2021, 11, 30, 16, 00, 00, 0, singaporeLocation),
			},
		},
		// Timezone in date time != location
		// Location will keep using the timezone in dateTime, which is in these tests are Jakarta (7*60*60 == 25200 offset)
		// Parsed time is not converted to target location.
		{
			desc:             "UTC+7 but Asia/Singapore - Using literal value (string) for datetime",
			dateTime:         "2021-11-30T15:00:00+07:00",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			dateTimeLocation: "Asia/Singapore",
			expectedValue:    time.Date(2021, 11, 30, 15, 00, 00, 0, time.FixedZone("", 25200)),
		},
		{
			desc:             "UTC+7 but Asia/Singapore - Using single value request jsonPath for datetime",
			dateTime:         "$.datetime",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			dateTimeLocation: "Asia/Singapore",
			requestJSON: []byte(`{
				"datetime": "2021-11-30T15:00:00+07:00"
			}`),
			expectedValue: time.Date(2021, 11, 30, 15, 00, 00, 0, time.FixedZone("", 25200)),
		},
		{
			desc:             "UTC+7 but Asia/Singapore - Using single value response jsonPath for datetime",
			dateTime:         "$.model_response.datetime",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			dateTimeLocation: "Asia/Singapore",
			responseJSON: []byte(`{
				"datetime": "2021-11-30T15:00:00+07:00"
			}`),
			expectedValue: time.Date(2021, 11, 30, 15, 00, 00, 0, time.FixedZone("", 25200)),
		},
		{
			desc:             "UTC+7 but Asia/Singapore - Using array from request jsonpath for datetime",
			dateTime:         "$.datetimes[*]",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			dateTimeLocation: "Asia/Singapore",
			requestJSON: []byte(`{
				"datetimes": ["2021-11-30T15:00:00+07:00", "2021-11-30T16:00:00+07:00"]
			}`),
			expectedValue: []interface{}{
				time.Date(2021, 11, 30, 15, 00, 00, 0, time.FixedZone("", 25200)),
				time.Date(2021, 11, 30, 16, 00, 00, 0, time.FixedZone("", 25200)),
			},
		},
		{
			desc:             "UTC+7 but Asia/Singapore - Using series",
			dateTime:         series.New([]interface{}{"2021-11-30T15:00:00+07:00", "2021-11-30T16:00:00+07:00"}, series.String, "datetime"),
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			dateTimeLocation: "Asia/Singapore",
			expectedValue: []interface{}{
				time.Date(2021, 11, 30, 15, 00, 00, 0, time.FixedZone("", 25200)),
				time.Date(2021, 11, 30, 16, 00, 00, 0, time.FixedZone("", 25200)),
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

			got := sr.ParseDateTime(tC.dateTime, tC.dateTimeFormat, tC.dateTimeLocation)
			assert.Equal(t, tC.expectedValue, got)
		})
	}
}
