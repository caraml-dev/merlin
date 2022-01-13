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
				Period: spec.PeriodType_UNDEFINED,
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
				Period: spec.PeriodType_UNDEFINED,
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
				Period: spec.PeriodType_UNDEFINED,
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
				Period: spec.PeriodType_UNDEFINED,
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
				Period: spec.PeriodType_UNDEFINED,
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
				Period: spec.PeriodType_UNDEFINED,
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
				Period: spec.PeriodType_UNDEFINED,
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

func TestCyclicalEncode(t *testing.T) {
	column := "col"
	columnX := "col_x"
	columnY := "col_y"
	testCases := []struct {
		desc            string
		cyclicalEncoder *CyclicalEncoder
		reqValues       []interface{}
		expectedResult  map[string]interface{}
		expectedError   error
	}{
		{
			desc: "Should fail: Fix period (by range): Missing value",
			cyclicalEncoder: &CyclicalEncoder{
				Period: spec.PeriodType_UNDEFINED,
				Min:    0,
				Max:    86400,
			},
			reqValues: []interface{}{
				1640995200, //3 Jan 2022, 00:00:00
				1640995201, //3 Jan 2022, 00:00:01
				nil,
				1641060000, //3 Jan 2022, 18:00:00
			},
			expectedResult: nil,
			expectedError:  fmt.Errorf("missing value"),
		},
		{
			desc: "Should fail: By Epoch time: Missing value",
			cyclicalEncoder: &CyclicalEncoder{
				Period: spec.PeriodType_MONTH,
				Min:    0,
				Max:    0,
			},
			reqValues: []interface{}{
				1640995200, //3 Jan 2022, 00:00:00
				1640995201, //3 Jan 2022, 00:00:01
				nil,
				1641060000, //3 Jan 2022, 18:00:00
			},
			expectedResult: nil,
			expectedError:  fmt.Errorf("missing value"),
		},
		{
			desc: "Should fail: Invalid value",
			cyclicalEncoder: &CyclicalEncoder{
				Period: spec.PeriodType_UNDEFINED,
				Min:    0,
				Max:    86400,
			},
			reqValues: []interface{}{
				1640995200, //3 Jan 2022, 00:00:00
				"hello",
				1641038400, //3 Jan 2022, 12:00:00
				1641060000, //3 Jan 2022, 18:00:00
			},
			expectedResult: nil,
			expectedError:  fmt.Errorf("strconv.ParseFloat: parsing \"hello\": invalid syntax"),
		},
		{
			desc: "Should succeed: Fixed period (by range)",
			cyclicalEncoder: &CyclicalEncoder{
				Period: spec.PeriodType_UNDEFINED,
				Min:    1,
				Max:    8,
			},
			reqValues: []interface{}{
				0, 1, 4.5, 8, 9.75,
			},
			expectedResult: map[string]interface{}{
				columnX: []interface{}{
					0.623489801, 1, -1, 1, 0,
				},
				columnY: []interface{}{
					-0.781831482, 0, 0, 0, 1,
				},
			},
		},
		{
			desc: "Should succeed: Fixed period (by range): with negative range",
			cyclicalEncoder: &CyclicalEncoder{
				Period: spec.PeriodType_UNDEFINED,
				Min:    -4,
				Max:    3,
			},
			reqValues: []interface{}{
				-5, -4, -0.5, 3, 4.75,
			},
			expectedResult: map[string]interface{}{
				columnX: []interface{}{
					0.623489801, 1, -1, 1, 0,
				},
				columnY: []interface{}{
					-0.781831482, 0, 0, 0, 1,
				},
			},
		},
		{
			desc: "Should succeed: EpochTime, HOUR",
			cyclicalEncoder: &CyclicalEncoder{
				Period: spec.PeriodType_UNDEFINED,
				Min:    0,
				Max:    3600,
			},
			reqValues: []interface{}{
				1640995200, //1 Jan 2022, 00:00:00
				1640995201, //1 Jan 2022, 00:00:01
				1641039300, //1 Jan 2022, 12:15:00
				1641054600, //1 Jan 2022, 16:30:00
				1641060000, //1 Jan 2022, 18:00:00
				1641063600, //1 Jan 2022, 19:00:00
				1641064500, //1 Jan 2022, 19:15:00
				1641065400, //1 Jan 2022, 19:30:00
			},
			expectedResult: map[string]interface{}{
				columnX: []interface{}{
					1, 0.999998476, 0, -1, 1, 1, 0, -1,
				},
				columnY: []interface{}{
					0, 0.001745328, 1, 0, 0, 0, 1, 0,
				},
			},
		},
		{
			desc: "Should succeed: EpochTime, DAY",
			cyclicalEncoder: &CyclicalEncoder{
				Period: spec.PeriodType_UNDEFINED,
				Min:    0,
				Max:    86400,
			},
			reqValues: []interface{}{
				//1 Jan cycle
				1640995200, //1 Jan 2022, 00:00:00
				1640995201, //1 Jan 2022, 00:00:01
				1641038400, //1 Jan 2022, 12:00:00
				1641060000, //1 Jan 2022, 18:00:00
				//2 Jan Cycle
				1641081600, //2 Jan 2022, 00:00:00
				1641081601, //2 Jan 2022, 00:00:01
				1641124800, //2 Jan 2022, 12:00:00
				1641146400, //2 Jan 2022, 18:00:00
			},
			expectedResult: map[string]interface{}{
				columnX: []interface{}{
					1, 0.999999997, -1, 0,
					1, 0.999999997, -1, 0,
				},
				columnY: []interface{}{
					0, 0.000072717, 0, -1,
					0, 0.000072717, 0, -1,
				},
			},
		},
		{
			desc: "Should succeed: EpochTime, WEEK",
			cyclicalEncoder: &CyclicalEncoder{
				Period: spec.PeriodType_UNDEFINED,
				Min:    0,
				Max:    604800,
			},
			reqValues: []interface{}{ //NOTE: Epoch time starts from Thursday, hence 0 on Thurs
				//Thursday, Friday, Sunday, Tuesday cycle 1
				1641427200, //6 Jan 2022, 00:00:00, Thursday
				1641578400, //5 Jan 2022, 18:00:00, Friday
				1641729600, //9 Jan 2022, 12:00:00, Sunday
				1641880800, //11 Jan 2022, 06:00:00, Tuesday
				//Thursday, Friday, Sunday, Tuesday cycle 2
				1642032000, //13 Jan 2022, 0:00:00, Thursday
				1642183200, //14 Jan 2022, 18:00:00, Friday
				1642334400, //16 Jan 2022, 12:00:00, Sunday
				1642485600, //18 Jan 2022, 06:00:00, Tuesday
			},
			expectedResult: map[string]interface{}{
				columnX: []interface{}{
					1, 0, -1, 0,
					1, 0, -1, 0,
				},
				columnY: []interface{}{
					0, 1, 0, -1,
					0, 1, 0, -1,
				},
			},
		},
		{
			desc: "Should succeed: EpochTime, MONTH",
			cyclicalEncoder: &CyclicalEncoder{
				Period: spec.PeriodType_MONTH,
				Min:    0,
				Max:    0,
			},
			reqValues: []interface{}{
				//Start, mid, end of Jan
				1640995200, //1 Jan 2022, 00:00:00
				1642334400, //16 Jan 2022, 12:00:00
				1643673599, //31 Jan 2022, 23:59:59
				//Start, mid, end of Feb
				1643673600, //1 Feb 2022, 00:00:00
				1644883200, //15 Feb 2022, 00:00:00
				1646092799, //28 Feb 2022, 23:59:59
				//Start, mid, end of Feb (leap)
				1580515200, //1 Feb 2020, 00:00:00
				1581768000, //15 Feb 2020, 12:00:00
				1583020799, //29 Feb 2022, 23:59:59
			},
			expectedResult: map[string]interface{}{
				columnX: []interface{}{
					1, -1, 0.999999999,
					1, -1, 0.999999999,
					1, -1, 0.999999999,
				},
				columnY: []interface{}{
					0, 0, -0.000002345,
					0, 0, -0.000002597,
					0, 0, -0.000002507,
				},
			},
		},
		{
			desc: "Should succeed: EpochTime, QUARTER",
			cyclicalEncoder: &CyclicalEncoder{
				Period: spec.PeriodType_QUARTER,
				Min:    0,
				Max:    0,
			},
			reqValues: []interface{}{
				1640995200, //1 Jan 2022, 00:00:00 - Q1
				1648771200, //1 Apr 2022, 00:00:00 - Q2
				1585699200, //1 Apr 2020, 00:00:00 - Q2 Leap
				1656633600, //1 Jul 2022, 00:00:00 - Q3
				1664582400, //1 Oct 2022, 00:00:00 - Q4
				1644883200, //15 Feb 2022, 00:00:00 - Mid Q1
				1581768000, //15 Feb 2020, 12:00:00 - Mid Q1 Leap
				1589630400, //16 May 2020, 12:00:00 - Mid Q2
				1599523200, //8 Sep 2020, 00:00:00 - 3/4 Q3
				1603497600, //24 Oct 2002, 00:00:00 - 1/4 Q4
			},
			expectedResult: map[string]interface{}{
				columnX: []interface{}{
					1, 1, 1, 1, 1, -1, -1, -1, 0, 0,
				},
				columnY: []interface{}{
					0, 0, 0, 0, 0, 0, 0, 0, -1, 1,
				},
			},
		},
		{
			desc: "Should succeed: EpochTime, HALF",
			cyclicalEncoder: &CyclicalEncoder{
				Period: spec.PeriodType_HALF,
				Min:    0,
				Max:    0,
			},
			reqValues: []interface{}{
				1640995200, //1 Jan 2022, 00:00:00 - H1
				1656633600, //1 Jul 2022, 00:00:00 - H2
				1648814400, //1 Apr 2022, 12:00:00 - Mid H1
				1652724000, //16 May 2022, 18:00:00 - 3/4 H1
				1585699200, //1 Apr 2020, 00:00:00 - Mid H1 Leap
				1664582400, //1 Oct 2022, 00:00:00 - Mid H2
				1601510400, //1 Oct 2020, 00:00:00 - Mid H2 Leap
				1668556800, //16 Nov 2022, 00:00:00 - 3/4 H2
				1597536000, //16 Aug 2002, 00:00:00 - 1/4 H2 Leap
			},
			expectedResult: map[string]interface{}{
				columnX: []interface{}{
					1, 1, -1, 0, -1, -1, -1, 0, 0,
				},
				columnY: []interface{}{
					0, 0, 0, -1, 0, 0, 0, -1, 1,
				},
			},
		},
		{
			desc: "Should succeed: EpochTime, YEAR",
			cyclicalEncoder: &CyclicalEncoder{
				Period: spec.PeriodType_YEAR,
				Min:    0,
				Max:    0,
			},
			reqValues: []interface{}{
				1514764800, //1 Jan 2018, 00:00:00 - Start
				1522648800, //2 Apr 2018, 06:00:00 - 1/4
				1530532800, //1 Apr 2018, 12:00:00 - 1/2
				1538416800, //1 Oct 2018, 18:00:00 - 3/4
				1451606400, //1 Jan 2016, 00:00:00 - Start Leap
				1459512000, //1 Apr 2016, 12:00:00 - 1/4 Leap
				1467417600, //2 Jul 2016, 00:00:00 - 1/2 Leap
				1475323200, //1 Oct 2016, 12:00:00 - 3/4 Leap
			},
			expectedResult: map[string]interface{}{
				columnX: []interface{}{
					1, 0, -1, 0, 1, 0, -1, 0,
				},
				columnY: []interface{}{
					0, 1, 0, -1, 0, 1, 0, -1,
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got, err := tC.cyclicalEncoder.Encode(tC.reqValues, column)
			if tC.expectedError != nil {
				assert.EqualError(t, err, tC.expectedError.Error())
			} else {
				assert.Equal(t, len(tC.expectedResult), len(got))
				assert.InDeltaSlice(t, tC.expectedResult[columnX], got[columnX], 0.00000001)
				assert.InDeltaSlice(t, tC.expectedResult[columnY], got[columnY], 0.00000001)
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
