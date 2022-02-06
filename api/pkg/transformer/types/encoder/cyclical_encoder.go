package encoder

import (
	"fmt"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types/converter"
	"math"
	"time"
)

const (
	floatZero   = 0.0000000001
	q1LastMonth = 3
	q2LastMonth = 6
	q3LastMonth = 9
	h1LastMonth = 6
	february    = 2

	minInSec  = 60
	hourInSec = 3600
	dayInSec  = 86400
	weekInSec = 604800

	daysInSec31 = 2678400
	daysInSec30 = 2592000
	daysInSec29 = 2505600
	daysInSec28 = 2419200

	q1InSec     = 7776000
	q1LeapInSec = 7862400
	q2InSec     = 7862400
	q3InSec     = 7948800
	q4InSec     = 7948800

	h1InSec     = 15638400
	h1LeapInSec = 15724800
	h2InSec     = 15897600

	yearInSec     = 31536000
	leapYearInSec = 31622400

	completeAngle = 2 * math.Pi

	// Unit angles for variable periods
	unitDaysInSec31 = completeAngle / daysInSec31
	unitDaysInSec30 = completeAngle / daysInSec30
	unitDaysInSec29 = completeAngle / daysInSec29
	unitDaysInSec28 = completeAngle / daysInSec28

	unitQ1InSec     = completeAngle / q1InSec
	unitQ1LeapInSec = completeAngle / q1LeapInSec
	unitQ2InSec     = completeAngle / q2InSec
	unitQ3InSec     = completeAngle / q3InSec
	unitQ4InSec     = completeAngle / q4InSec

	unitH1InSec     = completeAngle / h1InSec
	unitH1LeapInSec = completeAngle / h1LeapInSec
	unitH2InSec     = completeAngle / h2InSec

	unitYearInSec     = completeAngle / yearInSec
	unitLeapYearInSec = completeAngle / leapYearInSec
)

// Unit angles for variable periods for each month
var monthInSec = map[int]float64{
	1:  unitDaysInSec31,
	2:  unitDaysInSec28,
	3:  unitDaysInSec31,
	4:  unitDaysInSec30,
	5:  unitDaysInSec31,
	6:  unitDaysInSec30,
	7:  unitDaysInSec31,
	8:  unitDaysInSec31,
	9:  unitDaysInSec30,
	10: unitDaysInSec31,
	11: unitDaysInSec30,
	12: unitDaysInSec31,
	13: unitDaysInSec29, // Leap year Feb
}

type CyclicalEncoder struct {
	PeriodType spec.PeriodType
	Min        float64
	Max        float64
}

func NewCyclicalEncoder(config *spec.CyclicalEncoderConfig) (*CyclicalEncoder, error) {
	// by range
	byRange := config.GetByRange()
	if byRange != nil {
		if (byRange.Max - byRange.Min) < floatZero {
			return nil, fmt.Errorf("max of cyclical range must be larger than min")
		}

		return &CyclicalEncoder{
			PeriodType: spec.PeriodType_UNDEFINED,
			Min:        byRange.Min,
			Max:        byRange.Max,
		}, nil
	}

	// by epoch time
	byEpochTime := config.GetByEpochTime()
	var min, max float64 = 0, 0
	var period spec.PeriodType

	if byEpochTime != nil {
		switch byEpochTime.Period {
		case spec.PeriodType_HOUR:
			period = spec.PeriodType_UNDEFINED
			max = hourInSec
		case spec.PeriodType_DAY:
			period = spec.PeriodType_UNDEFINED
			max = dayInSec
		case spec.PeriodType_WEEK:
			period = spec.PeriodType_UNDEFINED
			max = weekInSec
		case spec.PeriodType_MONTH, spec.PeriodType_QUARTER, spec.PeriodType_HALF, spec.PeriodType_YEAR:
			period = byEpochTime.Period
			max = 0
		default:
			return nil, fmt.Errorf("invalid or unspported cycle period")
		}

		return &CyclicalEncoder{
			PeriodType: period,
			Min:        min,
			Max:        max,
		}, nil
	}

	return nil, fmt.Errorf("cyclical encoding config invalid or undefined")
}

func (oe *CyclicalEncoder) Encode(values []interface{}, column string) (map[string]interface{}, error) {
	encodedCos := make([]interface{}, 0, len(values))
	encodedSin := make([]interface{}, 0, len(values))

	// config with fixed range
	if oe.PeriodType == spec.PeriodType_UNDEFINED {
		period := oe.Max - oe.Min
		unitAngle := completeAngle / period

		for _, val := range values {
			// Check if value is missing
			if val == nil {
				return nil, fmt.Errorf("missing value")
			}

			// Check if value is valid
			valFloat, err := converter.ToFloat64(val)
			if err != nil {
				return nil, err
			}

			// Encode to sin and cos
			phase := (valFloat - oe.Min) * unitAngle
			encodedCos = append(encodedCos, math.Cos(phase))
			encodedSin = append(encodedSin, math.Sin(phase))
		}
	} else {
		// config with variable range, by epoch time (e.g. different days in each month, leap year etc.)
		for _, val := range values {
			// Check if value is missing
			if val == nil {
				return nil, fmt.Errorf("missing value")
			}

			// Check if value is valid
			valInt, err := converter.ToInt64(val)
			if err != nil {
				return nil, err
			}

			// convert epoch time to golang datetime
			t := time.Unix(valInt, 0).In(time.UTC)
			shareOfPeriod, err := getCycleTime(oe.PeriodType, t)
			if err != nil {
				return nil, err
			}
			unitAngle, err := getUnitAngle(oe.PeriodType, t)
			if err != nil {
				return nil, err
			}

			// Encode to sin and cos
			phase := float64(shareOfPeriod) * unitAngle
			encodedCos = append(encodedCos, math.Cos(phase))
			encodedSin = append(encodedSin, math.Sin(phase))
		}
	}

	return map[string]interface{}{
		column + "_x": encodedCos,
		column + "_y": encodedSin,
	}, nil
}

// Computes the number of seconds past the beginning of a pre-defined cycle
// Only works with PeriodType with variable cycle time such as Month, Year etc
// For period type with fixed cycle time, it is handled differently by encoder and
// does not need the cycle time to be computed
func getCycleTime(periodType spec.PeriodType, t time.Time) (int, error) {
	switch periodType {
	case spec.PeriodType_MONTH:
		dayElapsed := t.Day() - 1
		hr, min, sec := t.Clock()
		elapsed := getElapsedSec(dayElapsed, hr, min, sec)

		return elapsed, nil

	case spec.PeriodType_QUARTER:
		dayElapsed := t.YearDay() - 1
		hr, min, sec := t.Clock()
		elapsed := getElapsedSec(dayElapsed, hr, min, sec)
		var cycleTime int

		if t.Month() <= q1LastMonth {
			return elapsed, nil
		} else if t.Month() <= q2LastMonth {
			cycleTime = elapsed - q1InSec
		} else if t.Month() <= q3LastMonth {
			cycleTime = elapsed - h1InSec
		} else {
			cycleTime = elapsed - h1InSec - q3InSec
		}

		if isLeapYear(t.Year()) {
			cycleTime -= dayInSec //minus extra day from leap year
		}

		return cycleTime, nil

	case spec.PeriodType_HALF:
		dayElapsed := t.YearDay() - 1
		hr, min, sec := t.Clock()
		elapsed := getElapsedSec(dayElapsed, hr, min, sec)

		if t.Month() <= 6 {
			return elapsed, nil
		}

		if isLeapYear(t.Year()) {
			return elapsed - h1LeapInSec, nil
		}
		return elapsed - h1InSec, nil

	case spec.PeriodType_YEAR:
		dayElapsed := t.YearDay() - 1
		hr, min, sec := t.Clock()
		elapsed := getElapsedSec(dayElapsed, hr, min, sec)
		return elapsed, nil
	}

	return 0, fmt.Errorf("period type is undefined for this use case")
}

// Convert time duration in days, hour, min, sec to number of seconds
func getElapsedSec(dayElapsed int, hr int, min int, sec int) int {
	elapsed := dayElapsed*dayInSec + hr*hourInSec + min*minInSec + sec

	return elapsed
}

// Computes the angle in radians represented by per unit second of a pre-defined period
// This is derived from the formula for calculating phase:
// phase = time passed / period * 2pi
// By rearranging the formula (for optimizing computation) into this:
// phase = time pass * 2pi / period
// we define unit angle as (2pi / period)
// The motivation is that we can pre-compute this value once and use it repeatedly.
func getUnitAngle(periodType spec.PeriodType, t time.Time) (float64, error) {
	switch periodType {
	case spec.PeriodType_MONTH:
		if t.Month() == february && isLeapYear(t.Year()) {
			return monthInSec[13], nil
		}
		return monthInSec[int(t.Month())], nil
	case spec.PeriodType_QUARTER:
		if t.Month() <= q1LastMonth {
			if isLeapYear(t.Year()) {
				return unitQ1LeapInSec, nil
			}
			return unitQ1InSec, nil
		} else if t.Month() <= q2LastMonth {
			return unitQ2InSec, nil
		} else if t.Month() <= q3LastMonth {
			return unitQ3InSec, nil
		}
		return unitQ4InSec, nil
	case spec.PeriodType_HALF:
		if t.Month() <= h1LastMonth {
			if isLeapYear(t.Year()) {
				return unitH1LeapInSec, nil
			}
			return unitH1InSec, nil
		}
		return unitH2InSec, nil
	case spec.PeriodType_YEAR:
		if isLeapYear(t.Year()) {
			return unitLeapYearInSec, nil
		}
		return unitYearInSec, nil
	}

	return 0, fmt.Errorf("period type is undefined for this use case")
}

// test if a given year is leap year
// leap year is a year divisible by (4, but not 100) or (4, 100 and 400)
func isLeapYear(year int) bool {
	if (year%4 == 0 && year%100 != 0) || (year%4 == 0 && year%400 == 0) {
		return true
	}
	return false
}
