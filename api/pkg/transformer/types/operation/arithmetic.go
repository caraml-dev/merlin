package operation

import (
	"fmt"
	"math"

	"github.com/gojek/merlin/pkg/transformer/types/converter"
	"github.com/gojek/merlin/pkg/transformer/types/series"
)

func addOp(left, right interface{}, indexes *series.Series) (interface{}, error) {
	return doArithmeticOperation(left, right, Add, indexes)
}

func substractOp(left, right interface{}, indexes *series.Series) (interface{}, error) {
	return doArithmeticOperation(left, right, Substract, indexes)
}

func multiplyOp(left, right interface{}, indexes *series.Series) (interface{}, error) {
	return doArithmeticOperation(left, right, Multiply, indexes)
}

func divideOp(left, right interface{}, indexes *series.Series) (interface{}, error) {
	return doArithmeticOperation(left, right, Divide, indexes)
}

func moduloOp(left, right interface{}, indexes *series.Series) (interface{}, error) {
	return doArithmeticOperation(left, right, Modulo, indexes)
}

func calculateInInt64(left, right interface{}, operation ArithmeticOperator) (int64, error) {
	lValInt64, err := converter.ToInt64(left)
	if err != nil {
		return 0, err
	}
	rValInt64, err := converter.ToInt64(right)
	if err != nil {
		return 0, err
	}

	switch operation {
	case Add:
		return lValInt64 + rValInt64, nil
	case Substract:
		return lValInt64 - rValInt64, nil
	case Multiply:
		return lValInt64 * rValInt64, nil
	case Divide:
		if rValInt64 == 0 {
			return int64(math.NaN()), nil
		}
		return lValInt64 / rValInt64, nil
	case Modulo:
		if rValInt64 == 0 {
			return int64(math.NaN()), nil
		}
		return lValInt64 % rValInt64, nil
	default:
		return 0, fmt.Errorf("%s operation is not supported", operation)
	}
}

func calculateInFloat64(left, right interface{}, operation ArithmeticOperator) (float64, error) {
	lValFloat64, err := converter.ToFloat64(left)
	if err != nil {
		return 0, err
	}
	rValFloat64, err := converter.ToFloat64(right)
	if err != nil {
		return 0, err
	}

	switch operation {
	case Add:
		return lValFloat64 + rValFloat64, nil
	case Substract:
		return lValFloat64 - rValFloat64, nil
	case Multiply:
		return lValFloat64 * rValFloat64, nil
	case Divide:
		if rValFloat64 == 0 {
			return math.NaN(), nil
		}
		return lValFloat64 / rValFloat64, nil
	default:
		return 0, fmt.Errorf("%s operation is not supported", operation)
	}
}

func calculateInString(left, right interface{}, operation ArithmeticOperator) (string, error) {
	lValStr, err := converter.ToString(left)
	if err != nil {
		return "", err
	}
	rValStr, err := converter.ToString(right)
	if err != nil {
		return "", err
	}

	switch operation {
	case Add:
		return lValStr + rValStr, nil
	default:
		return "", fmt.Errorf("%s operation is not supported", operation)
	}
}

func doArithmeticOperation(left, right interface{}, operation ArithmeticOperator, indexes *series.Series) (interface{}, error) {
	var result interface{}
	var err error
	switch lVal := left.(type) {
	case int, *int, int8, *int8, int32, *int32, int64, *int64:
		switch rVal := right.(type) {
		case int, *int, int8, *int8, int32, *int32, int64, *int64:
			result, err = calculateInInt64(lVal, rVal, operation)
		case float32, *float32, float64, *float64:
			result, err = calculateInFloat64(lVal, rVal, operation)
		case series.Series, *series.Series:
			result, err = doOperationOnSeries(lVal, rVal, operation, indexes)
		default:
			err = fmt.Errorf("comparison not supported between %T and %T", lVal, rVal)
		}
	case float32, *float32, float64, *float64:
		switch rVal := right.(type) {
		case int, *int, int8, *int8, int32, *int32, int64, *int64, float32, *float32, float64, *float64:
			result, err = calculateInFloat64(lVal, rVal, operation)
		case series.Series, *series.Series:
			result, err = doOperationOnSeries(lVal, rVal, operation, indexes)
		default:
			err = fmt.Errorf("comparison not supported between %T and %T", lVal, rVal)
		}
	case string, *string:
		switch rVal := right.(type) {
		case string, *string:
			result, err = calculateInString(lVal, rVal, operation)
		case series.Series:
			result, err = doOperationOnSeries(lVal, rVal, operation, indexes)
		case *series.Series:
			result, err = doOperationOnSeries(lVal, rVal, operation, indexes)
		default:
			err = fmt.Errorf("comparison not supported between %T and %T", lVal, rVal)
		}
	case series.Series:
		lSeries := &lVal
		result, err = doOperationOnSeries(lSeries, right, operation, indexes)
	case *series.Series:
		result, err = doOperationOnSeries(lVal, right, operation, indexes)
	default:
		err = fmt.Errorf("type %T is not supported to do arithmetic operation", lVal)
	}

	if err != nil {
		return nil, err
	}
	return result, nil
}
