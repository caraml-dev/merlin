package symbol

import (
	"testing"
	"time"

	"github.com/caraml-dev/merlin/pkg/transformer/jsonpath"
	"github.com/caraml-dev/merlin/pkg/transformer/types/series"
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
				sr.SetRawRequest(requestJSONObj)
			}
			if responseJSONObj != nil {
				sr.SetModelResponse(responseJSONObj)
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
		timezone      interface{}
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
		{
			desc:      "multiple value timestamps and timezone",
			timestamp: series.New([]interface{}{1637441444, 1637527844}, series.Int, "timestamp"),
			timezone:  series.New([]interface{}{"Asia/Jakarta", "Asia/Jakarta"}, series.String, "timezone"),
			expectedValue: []interface{}{
				1, 0,
			},
		},
		{
			desc:      "multiple value timestamps and timezone using UTC",
			timestamp: series.New([]interface{}{1637441444, 1637527844}, series.Int, "timestamp"),
			timezone:  series.New([]interface{}{"UTC", "UTC"}, series.String, "timezone"),
			expectedValue: []interface{}{
				1, 1,
			},
		},
		{
			desc:      "Using single value request jsonPath for timestamp and timezone",
			timestamp: "$.timestamp",
			timezone:  "$.timezone",
			requestJSON: []byte(`{
				"timestamp": 1637445044,
				"timezone": "UTC"
			}`),
			expectedValue: 1,
		},
		{
			desc:      "Using multiple value request jsonPath for timestamp and timezone",
			timestamp: "$.timestamps",
			timezone:  "$.timezones",
			requestJSON: []byte(`{
				"timestamps": [1637441444,1637527844],
				"timezones": ["Asia/Jakarta", "Asia/Jakarta"]
			}`),
			expectedValue: []interface{}{
				1, 0,
			},
		},
		{
			desc:      "Using single value response jsonPath for timestamp and timezone",
			timestamp: "$.model_response.timestamp",
			timezone:  "$.model_response.timezone",
			responseJSON: []byte(`{
				"timestamp": 1637445044,
				"timezone": "UTC"
			}`),
			expectedValue: 1,
		},
		{
			desc:      "Using multiple value response jsonPath for timestamp",
			timestamp: "$.model_response.timestamps",
			timezone:  "$.model_response.timezones",
			responseJSON: []byte(`{
				"timestamps": [1637441444,1637527844],
				"timezones": ["Asia/Jakarta", "Asia/Jakarta"]
			}`),
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
				sr.SetRawRequest(requestJSONObj)
			}
			if responseJSONObj != nil {
				sr.SetModelResponse(responseJSONObj)
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
		{
			desc:      "Multiple timestamps and timezones from request jsonpath",
			timestamp: "$.timestamps",
			requestJSON: []byte(`{
				"timestamps":[1637441444,1637527844],
				"timezones":["Asia/Jakarta","Asia/Jakarta"]
			}`),
			format:   "2006-01-02",
			timezone: "$.timezones",
			expectedValue: []interface{}{
				"2021-11-21", "2021-11-22",
			},
		},
		{
			desc:      "Multiple timestamps and timezones from response jsonpath",
			timestamp: "$.model_response.timestamps",
			responseJSON: []byte(`{
				"timestamps":[1637441444,1637527844],
				"timezones":["Asia/Jakarta","Asia/Jakarta"]
			}`),
			format:   "2006-01-02",
			timezone: "$.model_response.timezones",
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
				sr.SetRawRequest(requestJSONObj)
			}
			if responseJSONObj != nil {
				sr.SetModelResponse(responseJSONObj)
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
		timezone      interface{}
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
		{
			desc:      "multiple value timestamps and timezones",
			timestamp: series.New([]interface{}{1637441444, 1637691859}, series.Int, "timestamp"),
			timezone:  series.New([]interface{}{"Asia/Jakarta", "Asia/Jakarta"}, series.String, "timezone"),
			expectedValue: []interface{}{
				0, 3,
			},
		},
		{
			desc:      "multiple value timestamps and timezones using UTC",
			timestamp: series.New([]interface{}{1637441444, 1637691859}, series.Int, "timestamp"),
			timezone:  series.New([]interface{}{"UTC", "UTC"}, series.String, "timezone"),
			expectedValue: []interface{}{
				6, 2,
			},
		},
		{
			desc:      "Using single value request jsonPath for timestamp and timezone",
			timestamp: "$.timestamp",
			timezone:  "$.timezone",
			requestJSON: []byte(`{
				"timestamp": 1637445044,
				"timezone": "UTC"
			}`),
			expectedValue: 6,
		},
		{
			desc:      "Using multiple value request jsonPath for timestamp and timezone",
			timestamp: "$.timestamps",
			timezone:  "$.timezones",
			requestJSON: []byte(`{
				"timestamps": [1637441444,1637691859],
				"timezones": ["Asia/Jakarta", "Asia/Jakarta"]
			}`),
			expectedValue: []interface{}{
				0, 3,
			},
		},
		{
			desc:      "Using single value response jsonPath for timestamp",
			timestamp: "$.model_response.timestamp",
			timezone:  "$.model_response.timezone",
			responseJSON: []byte(`{
				"timestamp": 1637445044,
				"timezone": "UTC"
			}`),
			expectedValue: 6,
		},
		{
			desc:      "Using multiple value response jsonPath for timestamp",
			timestamp: "$.model_response.timestamps",
			timezone:  "$.model_response.timezones",
			responseJSON: []byte(`{
				"timestamps": [1637441444,1637691859],
				"timezones": ["Asia/Jakarta", "Asia/Jakarta"]
			}`),
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
				sr.SetRawRequest(requestJSONObj)
			}
			if responseJSONObj != nil {
				sr.SetModelResponse(responseJSONObj)
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
		dateTimeLocation interface{}
		dateTimeFormat   string
		requestJSON      []byte
		responseJSON     []byte
		expectedValue    interface{}
		err              error
	}{
		// No TZ
		{
			desc:             "No TZ - Using literal value (string) for datetime",
			dateTime:         "2021-11-30 15:00:00",
			dateTimeLocation: "",
			dateTimeFormat:   "2006-01-02 15:04:05",
			expectedValue:    time.Date(2021, 11, 30, 15, 00, 00, 0, time.UTC),
		},
		{
			desc:             "No TZ - Using single value request jsonPath for datetime",
			dateTime:         "$.datetime",
			dateTimeLocation: "",
			dateTimeFormat:   "2006-01-02 15:04:05",
			requestJSON: []byte(`{
				"datetime": "2021-11-30 15:00:00"
			}`),
			expectedValue: time.Date(2021, 11, 30, 15, 00, 00, 0, time.UTC),
		},
		{
			desc:             "No TZ - Using single value response jsonPath for datetime",
			dateTime:         "$.model_response.datetime",
			dateTimeLocation: "",
			dateTimeFormat:   "2006-01-02 15:04:05",
			responseJSON: []byte(`{
				"datetime": "2021-11-30 15:00:00"
			}`),
			expectedValue: time.Date(2021, 11, 30, 15, 00, 00, 0, time.UTC),
		},
		{
			desc:             "No TZ - Using array from request jsonpath for datetime",
			dateTime:         "$.datetimes[*]",
			dateTimeLocation: "",
			dateTimeFormat:   "2006-01-02 15:04:05",
			requestJSON: []byte(`{
				"datetimes": ["2021-11-30 15:00:00", "2021-11-30 16:00:00"]
			}`),
			expectedValue: []interface{}{
				time.Date(2021, 11, 30, 15, 00, 00, 0, time.UTC),
				time.Date(2021, 11, 30, 16, 00, 00, 0, time.UTC),
			},
		},
		{
			desc:             "No TZ - Using series for datetime",
			dateTime:         series.New([]interface{}{"2021-11-30 15:00:00", "2021-11-30 16:00:00"}, series.String, "datetime"),
			dateTimeLocation: "",
			dateTimeFormat:   "2006-01-02 15:04:05",
			expectedValue: []interface{}{
				time.Date(2021, 11, 30, 15, 00, 00, 0, time.UTC),
				time.Date(2021, 11, 30, 16, 00, 00, 0, time.UTC),
			},
		},
		{
			desc:             "No TZ - Using array from request jsonpath for datetime and location",
			dateTime:         "$.datetimes[*]",
			dateTimeLocation: "$.locations[*]",
			dateTimeFormat:   "2006-01-02 15:04:05",
			requestJSON: []byte(`{
				"datetimes": ["2021-11-30 15:00:00", "2021-11-30 16:00:00"],
				"locations": ["", ""]
			}`),
			expectedValue: []interface{}{
				time.Date(2021, 11, 30, 15, 00, 00, 0, time.UTC),
				time.Date(2021, 11, 30, 16, 00, 00, 0, time.UTC),
			},
		},
		{
			desc:             "No TZ - Using series for datetime and location",
			dateTime:         series.New([]interface{}{"2021-11-30 15:00:00", "2021-11-30 16:00:00"}, series.String, "datetime"),
			dateTimeLocation: series.New([]interface{}{"", ""}, series.String, "location"),
			dateTimeFormat:   "2006-01-02 15:04:05",
			expectedValue: []interface{}{
				time.Date(2021, 11, 30, 15, 00, 00, 0, time.UTC),
				time.Date(2021, 11, 30, 16, 00, 00, 0, time.UTC),
			},
		},
		// No TZ in datetime, but location
		{
			desc:             "No TZ + Location - Using literal value (string) for datetime",
			dateTime:         "2021-11-30 15:00:00",
			dateTimeLocation: "Asia/Jayapura",
			dateTimeFormat:   "2006-01-02 15:04:05",
			expectedValue:    time.Date(2021, 11, 30, 15, 00, 00, 0, jayapuraLocation),
		},
		{
			desc:             "No TZ + Location - Using single value request jsonPath for datetime",
			dateTime:         "$.datetime",
			dateTimeLocation: "Asia/Jayapura",
			dateTimeFormat:   "2006-01-02 15:04:05",
			requestJSON: []byte(`{
				"datetime": "2021-11-30 15:00:00"
			}`),
			expectedValue: time.Date(2021, 11, 30, 15, 00, 00, 0, jayapuraLocation),
		},
		{
			desc:             "No TZ + Location - Using single value response jsonPath for datetime",
			dateTime:         "$.model_response.datetime",
			dateTimeLocation: "Asia/Jayapura",
			dateTimeFormat:   "2006-01-02 15:04:05",
			responseJSON: []byte(`{
				"datetime": "2021-11-30 15:00:00"
			}`),
			expectedValue: time.Date(2021, 11, 30, 15, 00, 00, 0, jayapuraLocation),
		},
		{
			desc:             "No TZ + Location - Using array from request jsonpath for datetime",
			dateTime:         "$.datetimes[*]",
			dateTimeLocation: "Asia/Jayapura",
			dateTimeFormat:   "2006-01-02 15:04:05",
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
			dateTimeLocation: "Asia/Jayapura",
			dateTimeFormat:   "2006-01-02 15:04:05",
			expectedValue: []interface{}{
				time.Date(2021, 11, 30, 15, 00, 00, 0, jayapuraLocation),
				time.Date(2021, 11, 30, 16, 00, 00, 0, jayapuraLocation),
			},
		},
		// UTC+7 Asia/Jakarta
		{
			desc:             "UTC+7 Asia/Jakarta - Using literal value (string) for datetime",
			dateTime:         "2021-11-30T15:00:00+07:00",
			dateTimeLocation: "Asia/Jakarta",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			expectedValue:    time.Date(2021, 11, 30, 15, 00, 00, 0, jakartaLocation),
		},
		{
			desc:             "UTC+7 Asia/Jakarta - Using single value request jsonPath for datetime",
			dateTime:         "$.datetime",
			dateTimeLocation: "Asia/Jakarta",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			requestJSON: []byte(`{
				"datetime": "2021-11-30T15:00:00+07:00"
			}`),
			expectedValue: time.Date(2021, 11, 30, 15, 00, 00, 0, jakartaLocation),
		},
		{
			desc:             "UTC+7 Asia/Jakarta - Using single value response jsonPath for datetime",
			dateTime:         "$.model_response.datetime",
			dateTimeLocation: "Asia/Jakarta",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			responseJSON: []byte(`{
				"datetime": "2021-11-30T15:00:00+07:00"
			}`),
			expectedValue: time.Date(2021, 11, 30, 15, 00, 00, 0, jakartaLocation),
		},
		{
			desc:             "UTC+7 Asia/Jakarta - Using array from request jsonpath for datetime",
			dateTime:         "$.datetimes[*]",
			dateTimeLocation: "Asia/Jakarta",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			requestJSON: []byte(`{
				"datetimes": ["2021-11-30T15:00:00+07:00", "2021-11-30T16:00:00+07:00"]
			}`),
			expectedValue: []interface{}{
				time.Date(2021, 11, 30, 15, 00, 00, 0, jakartaLocation),
				time.Date(2021, 11, 30, 16, 00, 00, 0, jakartaLocation),
			},
		},
		{
			desc:             "UTC+7 Asia/Jakarta - Using series for datetime",
			dateTime:         series.New([]interface{}{"2021-11-30T15:00:00+07:00", "2021-11-30T16:00:00+07:00"}, series.String, "datetime"),
			dateTimeLocation: "Asia/Jakarta",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			expectedValue: []interface{}{
				time.Date(2021, 11, 30, 15, 00, 00, 0, jakartaLocation),
				time.Date(2021, 11, 30, 16, 00, 00, 0, jakartaLocation),
			},
		},
		{
			desc:             "UTC+7 Asia/Jakarta - Using array from request jsonpath for datetime and location",
			dateTime:         "$.datetimes[*]",
			dateTimeLocation: "$.locations[*]",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			requestJSON: []byte(`{
				"datetimes": ["2021-11-30T15:00:00+07:00", "2021-11-30T16:00:00+07:00"],
				"locations": ["Asia/Jakarta", "Asia/Jakarta"]
			}`),
			expectedValue: []interface{}{
				time.Date(2021, 11, 30, 15, 00, 00, 0, jakartaLocation),
				time.Date(2021, 11, 30, 16, 00, 00, 0, jakartaLocation),
			},
		},
		{
			desc:             "UTC+7 Asia/Jakarta - Using series for datetime and location",
			dateTime:         series.New([]interface{}{"2021-11-30T15:00:00+07:00", "2021-11-30T16:00:00+07:00"}, series.String, "datetime"),
			dateTimeLocation: series.New([]interface{}{"Asia/Jakarta", "Asia/Jakarta"}, series.String, "location"),
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			expectedValue: []interface{}{
				time.Date(2021, 11, 30, 15, 00, 00, 0, jakartaLocation),
				time.Date(2021, 11, 30, 16, 00, 00, 0, jakartaLocation),
			},
		},
		// UTC+8 Asia/Singapore
		{
			desc:             "UTC+8 Asia/Singapore - Using literal value (string) for datetime",
			dateTime:         "2021-11-30T15:00:00+08:00",
			dateTimeLocation: "Asia/Singapore",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			expectedValue:    time.Date(2021, 11, 30, 15, 00, 00, 0, singaporeLocation),
		},
		{
			desc:             "UTC+8 Asia/Singapore - Using single value request jsonPath for datetime",
			dateTime:         "$.datetime",
			dateTimeLocation: "Asia/Singapore",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			requestJSON: []byte(`{
				"datetime": "2021-11-30T15:00:00+08:00"
			}`),
			expectedValue: time.Date(2021, 11, 30, 15, 00, 00, 0, singaporeLocation),
		},
		{
			desc:             "UTC+8 Asia/Singapore - Using single value response jsonPath for datetime",
			dateTime:         "$.model_response.datetime",
			dateTimeLocation: "Asia/Singapore",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			responseJSON: []byte(`{
				"datetime": "2021-11-30T15:00:00+08:00"
			}`),
			expectedValue: time.Date(2021, 11, 30, 15, 00, 00, 0, singaporeLocation),
		},
		{
			desc:             "UTC+8 Asia/Singapore - Using array from request jsonpath for datetime",
			dateTime:         "$.datetimes[*]",
			dateTimeLocation: "Asia/Singapore",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
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
			dateTimeLocation: "Asia/Singapore",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
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
			dateTimeLocation: "Asia/Singapore",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			expectedValue:    time.Date(2021, 11, 30, 15, 00, 00, 0, time.FixedZone("", 25200)),
		},
		{
			desc:             "UTC+7 but Asia/Singapore - Using single value request jsonPath for datetime",
			dateTime:         "$.datetime",
			dateTimeLocation: "Asia/Singapore",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			requestJSON: []byte(`{
				"datetime": "2021-11-30T15:00:00+07:00"
			}`),
			expectedValue: time.Date(2021, 11, 30, 15, 00, 00, 0, time.FixedZone("", 25200)),
		},
		{
			desc:             "UTC+7 but Asia/Singapore - Using single value response jsonPath for datetime",
			dateTime:         "$.model_response.datetime",
			dateTimeLocation: "Asia/Singapore",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
			responseJSON: []byte(`{
				"datetime": "2021-11-30T15:00:00+07:00"
			}`),
			expectedValue: time.Date(2021, 11, 30, 15, 00, 00, 0, time.FixedZone("", 25200)),
		},
		{
			desc:             "UTC+7 but Asia/Singapore - Using array from request jsonpath for datetime",
			dateTime:         "$.datetimes[*]",
			dateTimeLocation: "Asia/Singapore",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
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
			dateTimeLocation: "Asia/Singapore",
			dateTimeFormat:   "2006-01-02T15:04:05Z07:00",
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
				sr.SetRawRequest(requestJSONObj)
			}
			if responseJSONObj != nil {
				sr.SetModelResponse(responseJSONObj)
			}

			got := sr.ParseDateTime(tC.dateTime, tC.dateTimeLocation, tC.dateTimeFormat)
			assert.Equal(t, tC.expectedValue, got)
		})
	}
}
