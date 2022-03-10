package operation

import (
	"fmt"

	"github.com/gojek/merlin/pkg/transformer/types/converter"
	"github.com/gojek/merlin/pkg/transformer/types/operator"
	"github.com/gojek/merlin/pkg/transformer/types/series"
)

func compareInt64(lValue, rValue interface{}, comparator operator.Comparator) (bool, error) {
	compareFn := func(l, r int64, c operator.Comparator) (bool, error) {
		switch c {
		case operator.Greater:
			return (l > r), nil
		case operator.GreaterEq:
			return (l >= r), nil
		case operator.Less:
			return (l < r), nil
		case operator.LessEq:
			return (l <= r), nil
		case operator.Eq:
			return (l == r), nil
		case operator.Neq:
			return (l != r), nil
		default:
			return false, fmt.Errorf("%s operator is not supported", c)
		}
	}
	lValInt64, err := converter.ToInt64(lValue)
	if err != nil {
		return false, err
	}
	rValInt64, err := converter.ToInt64(rValue)
	if err != nil {
		return false, err
	}
	return compareFn(lValInt64, rValInt64, comparator)
}

func compareFloat64(lValue, rValue interface{}, comparator operator.Comparator) (bool, error) {
	compareFn := func(l, r float64, c operator.Comparator) (bool, error) {
		switch c {
		case operator.Greater:
			return (l > r), nil
		case operator.GreaterEq:
			return (l >= r), nil
		case operator.Less:
			return (l < r), nil
		case operator.LessEq:
			return (l <= r), nil
		case operator.Eq:
			return (l == r), nil
		case operator.Neq:
			return (l != r), nil
		default:
			return false, fmt.Errorf("%s operator is not supported", c)
		}
	}
	lValFloat64, err := converter.ToFloat64(lValue)
	if err != nil {
		return false, err
	}
	rValFloat64, err := converter.ToFloat64(rValue)
	if err != nil {
		return false, err
	}
	return compareFn(lValFloat64, rValFloat64, comparator)
}

func compareString(lValue, rValue interface{}, comparator operator.Comparator) (bool, error) {
	compareFn := func(l, r string, c operator.Comparator) (bool, error) {
		switch c {
		case operator.Greater:
			return (l > r), nil
		case operator.GreaterEq:
			return (l >= r), nil
		case operator.Less:
			return (l < r), nil
		case operator.LessEq:
			return (l <= r), nil
		case operator.Eq:
			return (l == r), nil
		case operator.Neq:
			return (l != r), nil
		default:
			return false, fmt.Errorf("%s operator is not supported", c)
		}
	}
	lValString, err := converter.ToString(lValue)
	if err != nil {
		return false, err
	}
	rValString, err := converter.ToString(rValue)
	if err != nil {
		return false, err
	}
	return compareFn(lValString, rValString, comparator)
}

func compareBool(lValue, rValue interface{}, comparator operator.Comparator) (bool, error) {
	compareFn := func(l, r bool, c operator.Comparator) (bool, error) {
		switch c {
		case operator.Eq:
			return (l == r), nil
		case operator.Neq:
			return (l != r), nil
		default:
			return false, fmt.Errorf("%s operator is not supported", c)
		}
	}
	lValBool, err := converter.ToBool(lValue)
	if err != nil {
		return false, err
	}
	rValBool, err := converter.ToBool(rValue)
	if err != nil {
		return false, err
	}
	return compareFn(lValBool, rValBool, comparator)
}

func greaterOp(left, right interface{}, indexes *series.Series) (interface{}, error) {
	return doComparison(left, right, operator.Greater, indexes)
}

func greaterEqOp(left, right interface{}, indexes *series.Series) (interface{}, error) {
	return doComparison(left, right, operator.GreaterEq, indexes)
}

func lessOp(left, right interface{}, indexes *series.Series) (interface{}, error) {
	return doComparison(left, right, operator.Less, indexes)
}

func lessEqOp(left, right interface{}, indexes *series.Series) (interface{}, error) {
	return doComparison(left, right, operator.LessEq, indexes)
}

func equalOp(left, right interface{}, indexes *series.Series) (interface{}, error) {
	return doComparison(left, right, operator.Eq, indexes)
}

func neqOp(left, right interface{}, indexes *series.Series) (interface{}, error) {
	return doComparison(left, right, operator.Neq, indexes)
}

func doComparison(left, right interface{}, operation operator.Comparator, indexes *series.Series) (interface{}, error) {
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
