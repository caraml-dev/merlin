package operation

import (
	"fmt"

	"github.com/gojek/merlin/pkg/transformer/types/converter"
	"github.com/gojek/merlin/pkg/transformer/types/series"
)

func compareInt64(lValue, rValue interface{}, comparator Comparator) (bool, error) {
	lValInt64, err := converter.ToInt64(lValue)
	if err != nil {
		return false, err
	}
	rValInt64, err := converter.ToInt64(rValue)
	if err != nil {
		return false, err
	}

	switch comparator {
	case Greater:
		return (lValInt64 > rValInt64), nil
	case GreaterEq:
		return (lValInt64 >= rValInt64), nil
	case Less:
		return (lValInt64 < rValInt64), nil
	case LessEq:
		return (lValInt64 <= rValInt64), nil
	case Eq:
		return (lValInt64 == rValInt64), nil
	case Neq:
		return (lValInt64 != rValInt64), nil
	default:
		return false, fmt.Errorf("%s operator is not supported", comparator)
	}
}

func compareFloat64(lValue, rValue interface{}, comparator Comparator) (bool, error) {
	lValFloat64, err := converter.ToFloat64(lValue)
	if err != nil {
		return false, err
	}
	rValFloat64, err := converter.ToFloat64(rValue)
	if err != nil {
		return false, err
	}
	switch comparator {
	case Greater:
		return (lValFloat64 > rValFloat64), nil
	case GreaterEq:
		return (lValFloat64 >= rValFloat64), nil
	case Less:
		return (lValFloat64 < rValFloat64), nil
	case LessEq:
		return (lValFloat64 <= rValFloat64), nil
	case Eq:
		return (lValFloat64 == rValFloat64), nil
	case Neq:
		return (lValFloat64 != rValFloat64), nil
	default:
		return false, fmt.Errorf("%s operator is not supported", comparator)
	}
}

func compareString(lValue, rValue interface{}, comparator Comparator) (bool, error) {
	lValString, err := converter.ToString(lValue)
	if err != nil {
		return false, err
	}
	rValString, err := converter.ToString(rValue)
	if err != nil {
		return false, err
	}
	switch comparator {
	case Greater:
		return (lValString > rValString), nil
	case GreaterEq:
		return (lValString >= rValString), nil
	case Less:
		return (lValString < rValString), nil
	case LessEq:
		return (lValString <= rValString), nil
	case Eq:
		return (lValString == rValString), nil
	case Neq:
		return (lValString != rValString), nil
	default:
		return false, fmt.Errorf("%s operator is not supported", comparator)
	}
}

func compareBool(lValue, rValue interface{}, comparator Comparator) (bool, error) {
	lValBool, err := converter.ToBool(lValue)
	if err != nil {
		return false, err
	}
	rValBool, err := converter.ToBool(rValue)
	if err != nil {
		return false, err
	}
	switch comparator {
	case Eq:
		return (lValBool == rValBool), nil
	case Neq:
		return (lValBool != rValBool), nil
	default:
		return false, fmt.Errorf("%s operator is not supported", comparator)
	}
}

func greaterOp(left, right interface{}, indexes *series.Series) (interface{}, error) {
	return doComparison(left, right, Greater, indexes)
}

func greaterEqOp(left, right interface{}, indexes *series.Series) (interface{}, error) {
	return doComparison(left, right, GreaterEq, indexes)
}

func lessOp(left, right interface{}, indexes *series.Series) (interface{}, error) {
	return doComparison(left, right, Less, indexes)
}

func lessEqOp(left, right interface{}, indexes *series.Series) (interface{}, error) {
	return doComparison(left, right, LessEq, indexes)
}

func equalOp(left, right interface{}, indexes *series.Series) (interface{}, error) {
	return doComparison(left, right, Eq, indexes)
}

func neqOp(left, right interface{}, indexes *series.Series) (interface{}, error) {
	return doComparison(left, right, Neq, indexes)
}

func doComparison(left, right interface{}, operation Comparator, indexes *series.Series) (interface{}, error) {
	var result interface{}
	var err error
	switch lVal := left.(type) {
	case int, *int, int8, *int8, int32, *int32, int64, *int64:
		switch rVal := right.(type) {
		case int, *int, int8, *int8, int32, *int32, int64, *int64:
			result, err = compareInt64(lVal, rVal, operation)
		case float32, *float32, float64, *float64:
			result, err = compareFloat64(lVal, rVal, operation)
		case series.Series, *series.Series:
			result, err = doOperationOnSeries(lVal, right, operation, indexes)
		default:
			err = fmt.Errorf("comparison not supported between %T and %T", lVal, rVal)
		}
	case float32, *float32, float64, *float64:
		switch rVal := right.(type) {
		case int, *int, int8, *int8, int32, *int32, int64, *int64, float32, *float32, float64, *float64:
			result, err = compareFloat64(lVal, rVal, operation)
		case series.Series, *series.Series:
			result, err = doOperationOnSeries(lVal, right, operation, indexes)
		default:
			err = fmt.Errorf("comparison not supported between %T and %T", lVal, rVal)
		}
	case string, *string:
		switch rVal := right.(type) {
		case string, *string:
			result, err = compareString(lVal, rVal, operation)
		case series.Series, *series.Series:
			result, err = doOperationOnSeries(lVal, right, operation, indexes)
		default:
			err = fmt.Errorf("comparison not supported between %T and %T", lVal, rVal)
		}
	case bool, *bool:
		switch rVal := right.(type) {
		case bool, *bool:
			result, err = compareBool(lVal, rVal, operation)
		case series.Series, *series.Series:
			result, err = doOperationOnSeries(lVal, right, operation, indexes)
		default:
			err = fmt.Errorf("comparison not supported between %T and %T", lVal, rVal)
		}
	case series.Series, *series.Series:
		result, err = doOperationOnSeries(lVal, right, operation, indexes)
	default:
		err = fmt.Errorf("unknown type %T", lVal)
	}

	if err != nil {
		return nil, err
	}
	return result, nil
}
