package operation

import (
	"fmt"

	"github.com/gojek/merlin/pkg/transformer/types/converter"
	"github.com/gojek/merlin/pkg/transformer/types/operator"
	"github.com/gojek/merlin/pkg/transformer/types/series"
)

func andOp(left interface{}, right interface{}) (interface{}, error) {
	return doLogicalOperation(left, right, operator.And, nil)
}

func orOp(left interface{}, right interface{}) (interface{}, error) {
	return doLogicalOperation(left, right, operator.Or, nil)
}

func boolLogicalOperation(left, right interface{}, operation operator.LogicalOperator) (interface{}, error) {
	lBool, err := converter.ToBool(left)
	if err != nil {
		return false, err
	}
	rBool, err := converter.ToBool(right)
	if err != nil {
		return false, err
	}
	switch operation {
	case operator.And:
		return lBool && rBool, nil
	case operator.Or:
		return lBool || rBool, nil
	default:
		return false, fmt.Errorf("%s operation is not supported for logical operation", operation)
	}
}

func doLogicalOperation(left, right interface{}, operation operator.LogicalOperator, indexes *series.Series) (interface{}, error) {
	var res interface{}
	var err error
	switch lVal := left.(type) {
	case bool:
		switch rVal := right.(type) {
		case bool, *bool:
			res, err = boolLogicalOperation(lVal, rVal, operation)
		case series.Series, *series.Series:
			res, err = doOperationOnSeries(lVal, rVal, operation, indexes)
		default:
			err = fmt.Errorf("logical operation %s is not supported for type %T and %T", operation, lVal, rVal)
		}
	case series.Series:
		res, err = doOperationOnSeries(lVal, right, operation, indexes)
	case *series.Series:
		res, err = doOperationOnSeries(lVal, right, operation, indexes)
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}
