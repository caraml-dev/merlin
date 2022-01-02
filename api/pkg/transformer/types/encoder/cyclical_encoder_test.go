package encoder

import (
	"fmt"
	"testing"
	"time"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/stretchr/testify/assert"
)

func TestNewCyclicalEncoder(t *testing.T) {
	testCases := []struct {
		desc            string
		config          *spec.CyclicalEncoderConfig
		expectedEncoder *CyclicalEncoder
		expectedErr     error
	}{
		{
			desc: "Should succeed - by positive int range",
			config: &spec.CyclicalEncoderConfig{
				EncodeBy: &spec.CyclicalEncoderConfig_ByRange{
					ByRange: &spec.ByRange{
						Min: 0,
						Max: 7,
					},
				},
			},
			expectedEncoder: &CyclicalEncoder{
				Period: nil,
				Min:    0,
				Max:    7,
			},
		},
		{
			desc: "Should succeed - by negative int range",
			config: &spec.CyclicalEncoderConfig{
				EncodeBy: &spec.CyclicalEncoderConfig_ByRange{
					ByRange: &spec.ByRange{
						Min: -2,
						Max: -1,
					},
				},
			},
			expectedEncoder: &CyclicalEncoder{
				Period: nil,
				Min:    -2,
				Max:    -1,
			},
		},
		{
			desc: "Should succeed - by positive float range",
			config: &spec.CyclicalEncoderConfig{
				EncodeBy: &spec.CyclicalEncoderConfig_ByRange{
					ByRange: &spec.ByRange{
						Min: 3.66,
						Max: 7.0,
					},
				},
			},
			expectedEncoder: &CyclicalEncoder{
				Period: nil,
				Min:    3.66,
				Max:    7.0,
			},
		},
		{
			desc: "Should succeed - mix +/- range",
			config: &spec.CyclicalEncoderConfig{
				EncodeBy: &spec.CyclicalEncoderConfig_ByRange{
					ByRange: &spec.ByRange{
						Min: -2,
						Max: 1.8,
					},
				},
			},
			expectedEncoder: &CyclicalEncoder{
				Period: nil,
				Min:    -2,
				Max:    1.8,
			},
		},
		{
			desc: "Should fail - Max == Min",
			config: &spec.CyclicalEncoderConfig{
				EncodeBy: &spec.CyclicalEncoderConfig_ByRange{
					ByRange: &spec.ByRange{
						Min: 7,
						Max: 7,
					},
				},
			},
			expectedEncoder: nil,
			expectedErr:     fmt.Errorf("max of cyclical range must be larger than min"),
		},
		{
			desc: "Should fail - Max < Min",
			config: &spec.CyclicalEncoderConfig{
				EncodeBy: &spec.CyclicalEncoderConfig_ByRange{
					ByRange: &spec.ByRange{
						Min: 9,
						Max: 7,
					},
				},
			},
			expectedEncoder: nil,
			expectedErr:     fmt.Errorf("max of cyclical range must be larger than min"),
		},
		{
			desc: "Should succeed - by epoch time: HOUR",
			config: &spec.CyclicalEncoderConfig{
				EncodeBy: &spec.CyclicalEncoderConfig_ByEpochTime{
					ByEpochTime: &spec.ByEpochTime{
						Period: spec.PeriodType_HOUR,
					},
				},
			},
			expectedEncoder: &CyclicalEncoder{
				Period: nil,
				Min:    0,
				Max:    HourInSec,
			},
		},
		{
			desc: "Should succeed - by epoch time: DAY",
			config: &spec.CyclicalEncoderConfig{
				EncodeBy: &spec.CyclicalEncoderConfig_ByEpochTime{
					ByEpochTime: &spec.ByEpochTime{
						Period: spec.PeriodType_DAY,
					},
				},
			},
			expectedEncoder: &CyclicalEncoder{
				Period: nil,
				Min:    0,
				Max:    DayInSec,
			},
		},
		{
			desc: "Should succeed - by epoch time: WEEK",
			config: &spec.CyclicalEncoderConfig{
				EncodeBy: &spec.CyclicalEncoderConfig_ByEpochTime{
					ByEpochTime: &spec.ByEpochTime{
						Period: spec.PeriodType_WEEK,
					},
				},
			},
			expectedEncoder: &CyclicalEncoder{
				Period: nil,
				Min:    0,
				Max:    WeekInSec,
			},
		},
		{
			desc: "Should succeed - by epoch time: MONTH",
			config: &spec.CyclicalEncoderConfig{
				EncodeBy: &spec.CyclicalEncoderConfig_ByEpochTime{
					ByEpochTime: &spec.ByEpochTime{
						Period: spec.PeriodType_MONTH,
					},
				},
			},
			expectedEncoder: &CyclicalEncoder{
				Period: spec.PeriodType_MONTH,
				Min:    0,
				Max:    0,
			},
		},
		{
			desc: "Should succeed - by epoch time: QUARTER",
			config: &spec.CyclicalEncoderConfig{
				EncodeBy: &spec.CyclicalEncoderConfig_ByEpochTime{
					ByEpochTime: &spec.ByEpochTime{
						Period: spec.PeriodType_QUARTER,
					},
				},
			},
			expectedEncoder: &CyclicalEncoder{
				Period: spec.PeriodType_QUARTER,
				Min:    0,
				Max:    0,
			},
		},
		{
			desc: "Should succeed - by epoch time: HALF",
			config: &spec.CyclicalEncoderConfig{
				EncodeBy: &spec.CyclicalEncoderConfig_ByEpochTime{
					ByEpochTime: &spec.ByEpochTime{
						Period: spec.PeriodType_HALF,
					},
				},
			},
			expectedEncoder: &CyclicalEncoder{
				Period: spec.PeriodType_HALF,
				Min:    0,
				Max:    0,
			},
		},
		{
			desc: "Should succeed - by epoch time: YEAR",
			config: &spec.CyclicalEncoderConfig{
				EncodeBy: &spec.CyclicalEncoderConfig_ByEpochTime{
					ByEpochTime: &spec.ByEpochTime{
						Period: spec.PeriodType_YEAR,
					},
				},
			},
			expectedEncoder: &CyclicalEncoder{
				Period: spec.PeriodType_YEAR,
				Min:    0,
				Max:    0,
			},
		},
		{
			desc: "Should fail - by epoch time: unsupported cycle period",
			config: &spec.CyclicalEncoderConfig{
				EncodeBy: &spec.CyclicalEncoderConfig_ByEpochTime{
					ByEpochTime: &spec.ByEpochTime{
						Period: 8,
					},
				},
			},
			expectedEncoder: nil,
			expectedErr:     fmt.Errorf("invalid or unspported cycle period"),
		},
		{
			desc: "Should fail - by epoch time: missing cycle period",
			config: &spec.CyclicalEncoderConfig{
				EncodeBy: &spec.CyclicalEncoderConfig_ByEpochTime{
					ByEpochTime: &spec.ByEpochTime{},
				},
			},
			expectedEncoder: nil,
			expectedErr:     fmt.Errorf("invalid or unspported cycle period"),
		},
		{
			desc: "Should fail - by epoch time: invalid config",
			config: &spec.CyclicalEncoderConfig{
				EncodeBy: &spec.CyclicalEncoderConfig_ByEpochTime{},
			},
			expectedEncoder: nil,
			expectedErr:     fmt.Errorf("cyclical encoding config invalid or undefined"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			gotEncoder, err := NewCyclicalEncoder(tC.config)
			if tC.expectedErr != nil {
				assert.EqualError(t, err, tC.expectedErr.Error())
			} else {
				assert.Equal(t, tC.expectedEncoder, gotEncoder)
			}
		})
	}
}

func TestGetCycleTime(t *testing.T) {
	testCases := []struct {
		desc              string
		periodType        interface{}
		epochTime         time.Time
		expectedCycleTime float64
		expectedErr       error
	}{
		{
			desc:              "Should fail - Invalid periodType",
			periodType:        123,
			epochTime:         time.Date(2021, 12, 31, 3, 15, 16, 0, time.UTC),
			expectedCycleTime: 0,
			expectedErr:       fmt.Errorf("invalid type for periodType"),
		},
		{
			desc:              "Should fail - Period: HOUR",
			periodType:        spec.PeriodType_HOUR,
			epochTime:         time.Date(2021, 12, 31, 3, 15, 16, 0, time.UTC),
			expectedCycleTime: 0,
			expectedErr:       fmt.Errorf("period type is undefined for this use case"),
		},
		{
			desc:              "Should fail - Period: DAY",
			periodType:        spec.PeriodType_DAY,
			epochTime:         time.Date(2021, 12, 31, 3, 15, 16, 0, time.UTC),
			expectedCycleTime: 0,
			expectedErr:       fmt.Errorf("period type is undefined for this use case"),
		},
		{
			desc:              "Should fail - Period: WEEK",
			periodType:        spec.PeriodType_WEEK,
			epochTime:         time.Date(2021, 12, 31, 3, 15, 16, 0, time.UTC),
			expectedCycleTime: 0,
			expectedErr:       fmt.Errorf("period type is undefined for this use case"),
		},
		{
			desc:              "Should succeed - Period: MONTH, leap",
			periodType:        spec.PeriodType_MONTH,
			epochTime:         time.Date(2016, 2, 29, 23, 59, 59, 0, time.UTC),
			expectedCycleTime: 2505599,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: MONTH, leap 2",
			periodType:        spec.PeriodType_MONTH,
			epochTime:         time.Date(2016, 6, 29, 23, 59, 59, 0, time.UTC),
			expectedCycleTime: 2505599,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: MONTH, non-leap",
			periodType:        spec.PeriodType_MONTH,
			epochTime:         time.Date(2017, 2, 28, 23, 59, 59, 0, time.UTC),
			expectedCycleTime: 2419199,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: MONTH, non-leap 2",
			periodType:        spec.PeriodType_MONTH,
			epochTime:         time.Date(2017, 9, 28, 23, 59, 59, 0, time.UTC),
			expectedCycleTime: 2419199,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: MONTH, edge: beginning of month",
			periodType:        spec.PeriodType_MONTH,
			epochTime:         time.Date(2017, 9, 1, 0, 0, 0, 0, time.UTC),
			expectedCycleTime: 0,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: QUARTER, Q1 Non-leap",
			periodType:        spec.PeriodType_QUARTER,
			epochTime:         time.Date(2017, 3, 15, 6, 44, 38, 0, time.UTC),
			expectedCycleTime: 6331478,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: QUARTER, Q1 Leap",
			periodType:        spec.PeriodType_QUARTER,
			epochTime:         time.Date(2016, 3, 15, 6, 44, 38, 0, time.UTC),
			expectedCycleTime: 6417878,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: QUARTER, Q2 Non-leap",
			periodType:        spec.PeriodType_QUARTER,
			epochTime:         time.Date(2017, 4, 15, 6, 44, 38, 0, time.UTC),
			expectedCycleTime: 1233878,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: QUARTER, Q2 Leap",
			periodType:        spec.PeriodType_QUARTER,
			epochTime:         time.Date(2016, 5, 1, 0, 0, 0, 0, time.UTC),
			expectedCycleTime: 2592000,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: QUARTER, Q3 Non-leap",
			periodType:        spec.PeriodType_QUARTER,
			epochTime:         time.Date(2017, 7, 15, 6, 44, 38, 0, time.UTC),
			expectedCycleTime: 1233878,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: QUARTER, Q3 Leap",
			periodType:        spec.PeriodType_QUARTER,
			epochTime:         time.Date(2016, 9, 5, 0, 0, 0, 0, time.UTC),
			expectedCycleTime: 5702400,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: QUARTER, Q4 Non-leap",
			periodType:        spec.PeriodType_QUARTER,
			epochTime:         time.Date(2013, 11, 15, 6, 44, 38, 0, time.UTC),
			expectedCycleTime: 3912278,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: QUARTER, Q4 Leap",
			periodType:        spec.PeriodType_QUARTER,
			epochTime:         time.Date(2020, 12, 5, 0, 0, 0, 0, time.UTC),
			expectedCycleTime: 5616000,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: HALF, H1 Non-leap",
			periodType:        spec.PeriodType_HALF,
			epochTime:         time.Date(2021, 6, 30, 23, 59, 59, 0, time.UTC),
			expectedCycleTime: 15638399,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: HALF, H1 Leap",
			periodType:        spec.PeriodType_HALF,
			epochTime:         time.Date(2016, 6, 30, 23, 59, 59, 0, time.UTC),
			expectedCycleTime: 15724799,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: HALF, H2 Non-leap",
			periodType:        spec.PeriodType_HALF,
			epochTime:         time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
			expectedCycleTime: 0,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: HALF, H2 Leap",
			periodType:        spec.PeriodType_HALF,
			epochTime:         time.Date(2016, 7, 1, 0, 0, 0, 0, time.UTC),
			expectedCycleTime: 0,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: YEAR, Non-leap",
			periodType:        spec.PeriodType_YEAR,
			epochTime:         time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
			expectedCycleTime: 15638400,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: YEAR, Leap",
			periodType:        spec.PeriodType_YEAR,
			epochTime:         time.Date(2016, 7, 1, 0, 0, 0, 0, time.UTC),
			expectedCycleTime: 15724800,
			expectedErr:       nil,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			cycleTime, err := getCycleTime(tC.periodType, tC.epochTime)
			if tC.expectedErr != nil {
				assert.EqualError(t, err, tC.expectedErr.Error())
			} else {
				assert.InDelta(t, tC.expectedCycleTime, cycleTime, 0.0000000001)
			}
		})
	}
}

func TestGetUnitAngle(t *testing.T) {
	testCases := []struct {
		desc              string
		periodType        interface{}
		epochTime         time.Time
		expectedUnitAngle float64
		expectedErr       error
	}{
		{
			desc:              "Should fail - Invalid periodType",
			periodType:        123,
			epochTime:         time.Date(2021, 12, 31, 3, 15, 16, 0, time.UTC),
			expectedUnitAngle: 0,
			expectedErr:       fmt.Errorf("invalid type for periodType"),
		},
		{
			desc:              "Should fail - Period: HOUR",
			periodType:        spec.PeriodType_HOUR,
			epochTime:         time.Date(2021, 12, 31, 3, 15, 16, 0, time.UTC),
			expectedUnitAngle: 0,
			expectedErr:       fmt.Errorf("period type is undefined for this use case"),
		},
		{
			desc:              "Should fail - Period: DAY",
			periodType:        spec.PeriodType_DAY,
			epochTime:         time.Date(2021, 12, 31, 3, 15, 16, 0, time.UTC),
			expectedUnitAngle: 0,
			expectedErr:       fmt.Errorf("period type is undefined for this use case"),
		},
		{
			desc:              "Should fail - Period: WEEK",
			periodType:        spec.PeriodType_WEEK,
			epochTime:         time.Date(2021, 12, 31, 3, 15, 16, 0, time.UTC),
			expectedUnitAngle: 0,
			expectedErr:       fmt.Errorf("period type is undefined for this use case"),
		},
		{
			desc:              "Should succeed - Period: MONTH, Jan",
			periodType:        spec.PeriodType_MONTH,
			epochTime:         time.Date(2016, 1, 28, 23, 59, 59, 0, time.UTC),
			expectedUnitAngle: 0.000002345,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: MONTH, Feb",
			periodType:        spec.PeriodType_MONTH,
			epochTime:         time.Date(2013, 2, 28, 23, 59, 59, 0, time.UTC),
			expectedUnitAngle: 0.000002597,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: MONTH, leap Feb",
			periodType:        spec.PeriodType_MONTH,
			epochTime:         time.Date(2016, 2, 29, 23, 59, 59, 0, time.UTC),
			expectedUnitAngle: 0.000002507,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: MONTH, Mar",
			periodType:        spec.PeriodType_MONTH,
			epochTime:         time.Date(2016, 3, 28, 23, 59, 59, 0, time.UTC),
			expectedUnitAngle: 0.000002345,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: MONTH, Apr",
			periodType:        spec.PeriodType_MONTH,
			epochTime:         time.Date(2016, 4, 28, 23, 59, 59, 0, time.UTC),
			expectedUnitAngle: 0.000002424,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: MONTH, May",
			periodType:        spec.PeriodType_MONTH,
			epochTime:         time.Date(2016, 5, 28, 23, 59, 59, 0, time.UTC),
			expectedUnitAngle: 0.000002345,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: MONTH, Jun",
			periodType:        spec.PeriodType_MONTH,
			epochTime:         time.Date(2016, 6, 28, 23, 59, 59, 0, time.UTC),
			expectedUnitAngle: 0.000002424,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: MONTH, Jul",
			periodType:        spec.PeriodType_MONTH,
			epochTime:         time.Date(2016, 7, 28, 23, 59, 59, 0, time.UTC),
			expectedUnitAngle: 0.000002345,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: MONTH, Aug",
			periodType:        spec.PeriodType_MONTH,
			epochTime:         time.Date(2016, 8, 28, 23, 59, 59, 0, time.UTC),
			expectedUnitAngle: 0.000002345,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: MONTH, Sep",
			periodType:        spec.PeriodType_MONTH,
			epochTime:         time.Date(2016, 9, 28, 23, 59, 59, 0, time.UTC),
			expectedUnitAngle: 0.000002424,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: MONTH, Oct",
			periodType:        spec.PeriodType_MONTH,
			epochTime:         time.Date(2016, 10, 28, 23, 59, 59, 0, time.UTC),
			expectedUnitAngle: 0.000002345,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: MONTH, Nov",
			periodType:        spec.PeriodType_MONTH,
			epochTime:         time.Date(2016, 11, 28, 23, 59, 59, 0, time.UTC),
			expectedUnitAngle: 0.000002424,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: MONTH, Dec",
			periodType:        spec.PeriodType_MONTH,
			epochTime:         time.Date(2016, 12, 28, 23, 59, 59, 0, time.UTC),
			expectedUnitAngle: 0.000002345,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: QUARTER, Q1",
			periodType:        spec.PeriodType_QUARTER,
			epochTime:         time.Date(2013, 2, 28, 23, 59, 59, 0, time.UTC),
			expectedUnitAngle: 0.000000808,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: QUARTER, Q1 Leap",
			periodType:        spec.PeriodType_QUARTER,
			epochTime:         time.Date(2016, 2, 28, 23, 59, 59, 0, time.UTC),
			expectedUnitAngle: 0.000000799,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: QUARTER, Q2",
			periodType:        spec.PeriodType_QUARTER,
			epochTime:         time.Date(2013, 4, 1, 0, 0, 0, 0, time.UTC),
			expectedUnitAngle: 0.000000799,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: QUARTER, Q3",
			periodType:        spec.PeriodType_QUARTER,
			epochTime:         time.Date(2013, 7, 1, 0, 0, 0, 0, time.UTC),
			expectedUnitAngle: 0.00000079,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: QUARTER, Q4",
			periodType:        spec.PeriodType_QUARTER,
			epochTime:         time.Date(2013, 12, 31, 23, 59, 59, 0, time.UTC),
			expectedUnitAngle: 0.00000079,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: HALF, H1",
			periodType:        spec.PeriodType_HALF,
			epochTime:         time.Date(2013, 1, 1, 0, 0, 0, 0, time.UTC),
			expectedUnitAngle: 0.000000401,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: HALF, H1 Leap",
			periodType:        spec.PeriodType_HALF,
			epochTime:         time.Date(2016, 6, 30, 23, 59, 59, 0, time.UTC),
			expectedUnitAngle: 0.000000399,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: HALF, H2",
			periodType:        spec.PeriodType_HALF,
			epochTime:         time.Date(2013, 7, 1, 0, 0, 0, 0, time.UTC),
			expectedUnitAngle: 0.000000395,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: YEAR, Non-Leap",
			periodType:        spec.PeriodType_YEAR,
			epochTime:         time.Date(2013, 1, 1, 0, 0, 0, 0, time.UTC),
			expectedUnitAngle: 0.000000199,
			expectedErr:       nil,
		},
		{
			desc:              "Should succeed - Period: Year, Leap",
			periodType:        spec.PeriodType_YEAR,
			epochTime:         time.Date(2016, 12, 31, 23, 59, 59, 0, time.UTC),
			expectedUnitAngle: 0.000000198,
			expectedErr:       nil,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			unitAngle, err := getUnitAngle(tC.periodType, tC.epochTime)
			if tC.expectedErr != nil {
				assert.EqualError(t, err, tC.expectedErr.Error())
			} else {
				assert.InDelta(t, tC.expectedUnitAngle, unitAngle, 0.000000001)
			}
		})
	}
}

func TestIsLeapYear(t *testing.T) {
	testCases := []struct {
		desc           string
		year           int
		expectedResult bool
	}{
		{
			desc:           "Leap year - divisible by 4 but not 100",
			year:           2012,
			expectedResult: true,
		},
		{
			desc:           "Non-leap year - divisible by 4 and 100 but not 400",
			year:           2100,
			expectedResult: false,
		},
		{
			desc:           "Leap year- divisible by 4, 100 and 400",
			year:           2400,
			expectedResult: true,
		},
		{
			desc:           "Non-leap year - others",
			year:           2021,
			expectedResult: false,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			result := isLeapYear(tC.year)
			assert.Equal(t, tC.expectedResult, result)
		})
	}
}

/*
   func TestCyclicalEncode(t *testing.T) {
   	column := "col"
   	testCases := []struct {
   		desc           string
   		ordinalEncoder *OrdinalEncoder
   		reqValues      []interface{}
   		expectedResult map[string]interface{}
   		expectedError  error
   	}{
   		{
   			desc: "Should success",
   			ordinalEncoder: &OrdinalEncoder{
   				DefaultValue: 0,
   				Mapping: map[string]interface{}{
   					"suv":   1,
   					"sedan": 2,
   					"mpv":   3,
   				},
   			},
   			reqValues: []interface{}{
   				"suv", "sedan", "mpv", nil, "sport car",
   			},
   			expectedResult: map[string]interface{}{
   				column: []interface{}{
   					1, 2, 3, 0, 0,
   				},
   			},
   		},
   		{
   			desc: "Should success, orginal value is integer",
   			ordinalEncoder: &OrdinalEncoder{
   				DefaultValue: nil,
   				Mapping: map[string]interface{}{
   					"1": 1,
   					"2": 2,
   					"3": 3,
   				},
   			},
   			reqValues: []interface{}{
   				1, 2, 3, nil, 4,
   			},
   			expectedResult: map[string]interface{}{
   				column: []interface{}{
   					1, 2, 3, nil, nil,
   				},
   			},
   		},
   		{
   			desc: "Should success, orginal value is integer",
   			ordinalEncoder: &OrdinalEncoder{
   				DefaultValue: float64(1),
   				Mapping: map[string]interface{}{
   					"true":  float64(100),
   					"false": float64(0),
   				},
   			},
   			reqValues: []interface{}{
   				true, false, nil,
   			},
   			expectedResult: map[string]interface{}{
   				column: []interface{}{
   					float64(100),
   					float64(0),
   					float64(1),
   				},
   			},
   		},
   	}
   	for _, tC := range testCases {
   		t.Run(tC.desc, func(t *testing.T) {
   			got, err := tC.ordinalEncoder.Encode(tC.reqValues, column)
   			if tC.expectedError != nil {
   				assert.EqualError(t, err, tC.expectedError.Error())
   			} else {
   				assert.Equal(t, tC.expectedResult, got)
   			}
   		})
   	}
   }*/
