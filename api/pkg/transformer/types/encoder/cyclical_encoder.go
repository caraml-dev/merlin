package encoder

import (
	"fmt"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types/converter"
	"math"
	"time"
)

const (
	MinInSec      = 60
	HourInSec     = 3600
	DayInSec      = 86400
	WeekInSec     = 604800
	completeAngle = 2 * math.Pi

	//Unit angles for variable periods
	DaysInSec31 = completeAngle / 2678400
	DaysInSec30 = completeAngle / 2592000
	DaysInSec29 = completeAngle / 2505600
	DaysInSec28 = completeAngle / 2419200

	Q1InSec     = completeAngle / 7776000
	Q1LeapInSec = completeAngle / 7862400
	Q2InSec     = completeAngle / 7862400
	Q3InSec     = completeAngle / 7948800
	Q4InSec     = completeAngle / 7948800

	H1InSec     = completeAngle / 15638400
	H1LeapInSec = completeAngle / 15724800
	H2InSec     = completeAngle / 15897600

	YearInSec     = completeAngle / 31536000
	LeapYearInSec = completeAngle / 31622400
)

//Unit angles for variable periods for each month
var MonthInSec = map[int]float64{
	1:  DaysInSec31,
	2:  DaysInSec28,
	3:  DaysInSec31,
	4:  DaysInSec30,
	5:  DaysInSec31,
	6:  DaysInSec30,
	7:  DaysInSec31,
	8:  DaysInSec31,
	9:  DaysInSec30,
	10: DaysInSec31,
	11: DaysInSec30,
	12: DaysInSec31,
	13: DaysInSec29, //Leap year Feb
}

type CyclicalEncoder struct {
	Period interface{}
	Min    float64
	Max    float64
}

func NewCyclicalEncoder(config *spec.CyclicalEncoderConfig) (*CyclicalEncoder, error) {
	// by range
	byRange := config.GetByRange()
	if byRange != nil {
		if byRange.Max < byRange.Min {
			return nil, fmt.Errorf("max of cyclical range must be larger than min")
		}

		return &CyclicalEncoder{
			Period: nil,
			Min:    byRange.Min,
			Max:    byRange.Max,
		}, nil
	}

	// by epoch time
	byEpochTime := config.GetByEpochTime()
	var min, max float64 = 0, 0
	var period interface{}

	if byEpochTime != nil {
		switch byEpochTime.Period {
		case spec.PeriodType_HOUR:
			period = nil
			max = HourInSec
		case spec.PeriodType_DAY:
			period = nil
			max = DayInSec
		case spec.PeriodType_WEEK:
			period = nil
			max = WeekInSec
		case spec.PeriodType_MONTH, spec.PeriodType_QUARTER, spec.PeriodType_HALF, spec.PeriodType_YEAR:
			period = byEpochTime.Period
			max = 0
		}

		return &CyclicalEncoder{
			Period: period,
			Min:    min,
			Max:    max,
		}, nil
	}

	return nil, fmt.Errorf("cyclical encoding config invalid or undefined")
}

func (oe *CyclicalEncoder) Encode(values []interface{}, column string) (map[string]interface{}, error) {
	encodedCos := make([]interface{}, 0, len(values))
	encodedSin := make([]interface{}, 0, len(values))

	// config with fixed range
	if oe.Period == nil {
		period := oe.Max - oe.Min
		unitAngle := completeAngle / period

		for _, val := range values {
			//Check if value is missing
			if val == nil {
				return nil, fmt.Errorf("missing value")
			}

			//Check if value is valid
			valFloat, err := converter.ToFloat64(val)
			if err != nil {
				return nil, err
			}

			//Encode to sin and cos
			phase := (valFloat - oe.Min) * unitAngle
			encodedCos = append(encodedCos, math.Cos(phase))
			encodedSin = append(encodedSin, math.Sin(phase))
		}
	} else {
		//config with variable range, by epoch time (e.g. different days in each month, leap year etc.)
		for _, val := range values {
			//Check if value is missing
			if val == nil {
				return nil, fmt.Errorf("missing value")
			}

			//Check if value is valid
			valInt, err := converter.ToInt64(val)
			if err != nil {
				return nil, err
			}

			//convert epoch time to golang datetime
			t := time.Unix(valInt, 0)
			shareOfPeriod, err := getCycleTime(oe.Period, t)
			if err != nil {
				return nil, err
			}
			unitAngle, err := getUnitAngle(oe.Period, t)
			if err != nil {
				return nil, err
			}

			//Encode to sin and cos
			phase := shareOfPeriod * unitAngle
			encodedCos = append(encodedCos, math.Cos(phase))
			encodedSin = append(encodedSin, math.Sin(phase))
		}
	}

	return map[string]interface{}{
		column + "_x": encodedCos,
		column + "_y": encodedSin,
	}, nil
}

//Computes the number of seconds past the beginning of a pre-defined cycle
func getCycleTime(periodType interface{}, t time.Time) (float64, error) {
	_, pType := periodType.(spec.PeriodType)
	if !pType {
		return 0, fmt.Errorf("invalid type for periodType")
	}

	switch periodType {
	case spec.PeriodType_MONTH:
		day := t.Day()
		hr, min, sec := t.Clock()
		elapsed := float64(day*DayInSec + hr*HourInSec + min*MinInSec + sec)
		return elapsed, nil
	case spec.PeriodType_QUARTER:
		day := t.YearDay()
		hr, min, sec := t.Clock()
		elapsed := float64(day*DayInSec + hr*HourInSec + min*MinInSec + sec)

		if t.Month() <= 3 {
			return elapsed, nil
		} else if t.Month() <= 6 {
			if isLeapYear(t.Year()) {
				return elapsed - Q1LeapInSec, nil
			}
			return elapsed - Q1InSec, nil
		} else if t.Month() <= 9 {
			if isLeapYear(t.Year()) {
				return elapsed - H1LeapInSec, nil
			}
			return elapsed - H1InSec, nil
		}

		if isLeapYear(t.Year()) {
			return elapsed - H1LeapInSec - Q3InSec, nil
		}
		return elapsed - H1InSec - Q3InSec, nil

	case spec.PeriodType_HALF:
		day := t.YearDay()
		hr, min, sec := t.Clock()
		elapsed := float64(day*DayInSec + hr*HourInSec + min*MinInSec + sec)

		if t.Month() <= 6 {
			return elapsed, nil
		}

		if isLeapYear(t.Year()) {
			return elapsed - H1LeapInSec, nil
		}
		return elapsed - H1InSec, nil

	case spec.PeriodType_YEAR:
		day := t.YearDay()
		hr, min, sec := t.Clock()
		elapsed := float64(day*DayInSec + hr*HourInSec + min*MinInSec + sec)
		return elapsed, nil
	}

	return 0, fmt.Errorf("PeriodType is unknown")
}

//Computes the angle in radians represented by per unit second of a pre-defined period
//This is derived from the formula for calculating phase:
//phase = time passed / period * 2pi
//By rearranging the formula (for optimizing computation) into this:
//phase = time pass * 2pi / period
//we define unit angle as (2pi / period)
//The motivation is that we can pre-compute this value once and use it repeatedly.
func getUnitAngle(periodType interface{}, t time.Time) (float64, error) {
	_, pType := periodType.(spec.PeriodType)
	if !pType {
		return 0, fmt.Errorf("invalid type for periodType")
	}

	switch periodType {
	case spec.PeriodType_MONTH:
		if t.Month() == 2 && isLeapYear(t.Year()) {
			return MonthInSec[13], nil
		}
		return MonthInSec[int(t.Month())], nil
	case spec.PeriodType_QUARTER:
		if t.Month() <= 3 {
			if isLeapYear(t.Year()) {
				return Q1LeapInSec, nil
			}
			return Q1InSec, nil
		} else if t.Month() <= 6 {
			return Q2InSec, nil
		} else if t.Month() <= 9 {
			return Q3InSec, nil
		}
		return Q4InSec, nil
	case spec.PeriodType_HALF:
		if t.Month() <= 6 {
			if isLeapYear(t.Year()) {
				return H1LeapInSec, nil
			}
			return H1InSec, nil
		}
		return H2InSec, nil
	case spec.PeriodType_YEAR:
		if isLeapYear(t.Year()) {
			return LeapYearInSec, nil
		}
		return YearInSec, nil
	}

	return 0, fmt.Errorf("PeriodType is unknown")
}

func isLeapYear(year int) bool {
	if (year%4 == 0 && year%100 != 0) || (year%4 == 0 && year%400 == 0) {
		return true
	}
	return false
}
