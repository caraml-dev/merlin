package symbol

import (
	"fmt"

	"github.com/gojek/merlin/pkg/transformer/types/comparator"
	"github.com/gojek/merlin/pkg/transformer/types/series"
)

func (sr Registry) Greater(left, right interface{}) interface{} {
	return doComparison(left, right, comparator.Greater)
}

func (sr Registry) GreaterEq(left, right interface{}) interface{} {
	return doComparison(left, right, comparator.GreaterEq)
}

func (sr Registry) Less(left, right interface{}) interface{} {
	return doComparison(left, right, comparator.Less)
}

func (sr Registry) LessEq(left, right interface{}) interface{} {
	return doComparison(left, right, comparator.LessEq)
}

func (sr Registry) Equal(left, right interface{}) interface{} {
	return doComparison(left, right, comparator.Eq)
}

func (sr Registry) Neq(left, right interface{}) interface{} {
	return doComparison(left, right, comparator.Neq)
}

func doComparison(left, right interface{}, operation comparator.Comparator) interface{} {
	var result interface{}
	var err error
	switch lVal := left.(type) {
	case int, *int, int8, *int8, int32, *int32, int64, *int64:
		switch rVal := right.(type) {
		case int, *int, int8, *int8, int32, *int32, int64, *int64:
			result, err = comparator.CompareInt64(lVal, rVal, operation)
		case float32, *float32, float64, *float64:
			result, err = comparator.CompareFloat64(lVal, rVal, operation)
		case series.Series, *series.Series:
			result, err = compareWithSeries(lVal, right, operation)
		default:
			panic(fmt.Errorf("comparison not supported between %T and %T", lVal, rVal))
		}
	case float32, *float32, float64, *float64:
		switch rVal := right.(type) {
		case int, *int, int8, *int8, int32, *int32, int64, *int64, float32, *float32, float64, *float64:
			result, err = comparator.CompareFloat64(lVal, rVal, operation)
		case series.Series, *series.Series:
			result, err = compareWithSeries(lVal, right, operation)
		default:
			panic(fmt.Errorf("comparison not supported between %T and %T", lVal, rVal))
		}
	case string, *string:
		switch rVal := right.(type) {
		case string, *string:
			result, err = comparator.CompareString(lVal, rVal, operation)
		case series.Series, *series.Series:
			result, err = compareWithSeries(lVal, right, operation)
		default:
			panic(fmt.Errorf("comparison not supported between %T and %T", lVal, rVal))
		}
	case bool, *bool:
		switch rVal := right.(type) {
		case bool, *bool:
			result, err = comparator.CompareBool(lVal, rVal, operation)
		case series.Series, *series.Series:
			result, err = compareWithSeries(lVal, right, operation)
		default:
			panic(fmt.Errorf("comparison not supported between %T and %T", lVal, rVal))
		}
	case series.Series:
		lSeries := &lVal
		result, err = lSeries.Compare(operation, right)
	case *series.Series:
		result, err = lVal.Compare(operation, right)
	default:
		panic(fmt.Errorf("unknown type %T", lVal))
	}

	if err != nil {
		panic(err)
	}
	return result
}

func compareWithSeries(left, right interface{}, operation comparator.Comparator) (interface{}, error) {
	var rSeries *series.Series
	if val, ok := right.(series.Series); ok {
		rSeries = &val
	}
	if val, ok := right.(*series.Series); ok {
		rSeries = val
	}

	if rSeries == nil {
		return nil, fmt.Errorf("right value is not in series type")
	}

	switch lVal := left.(type) {
	case series.Series, *series.Series:
		var lSeries *series.Series
		if val, ok := lVal.(series.Series); ok {
			lSeries = &val
		}
		if val, ok := lVal.(*series.Series); ok {
			lSeries = val
		}
		return lSeries.Compare(operation, right)
	default:
		lSeries, err := broadcastedSeries(left, rSeries.Series().Len(), "leftSeries")
		if err != nil {
			return nil, err
		}
		return lSeries.Compare(operation, right)
	}
}

func broadcastedSeries(value interface{}, numOfRow int, colName string) (*series.Series, error) {
	values := make([]interface{}, 0, numOfRow)
	for i := 0; i < numOfRow; i++ {
		values = append(values, value)
	}
	return series.NewInferType(values, colName)
}

func (sr Registry) And(left series.Series, right series.Series) series.Series {
	return series.Series{}
}

func (sr Registry) Or(left series.Series, right series.Series) series.Series {
	return series.Series{}
}
