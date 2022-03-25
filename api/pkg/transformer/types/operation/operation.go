package operation

import (
	"fmt"

	"github.com/gojek/merlin/pkg/transformer/types/operator"
	"github.com/gojek/merlin/pkg/transformer/types/series"
)

type Operator interface {
	Type() string
	Name() string
}

type OperationNode struct {
	LeftVal   interface{}
	RightVal  interface{}
	Operation Operator
	Next      *OperationNode
}

func RegisterOperation(leftVal, rightVal interface{}, operator Operator) *OperationNode {
	switch lVal := leftVal.(type) {
	case *OperationNode:
		return lVal.addNextOperation(&OperationNode{
			RightVal:  rightVal,
			Operation: operator,
		})
	case OperationNode:
		return lVal.addNextOperation(&OperationNode{
			RightVal:  rightVal,
			Operation: operator,
		})
	default:
		operationNode := &OperationNode{
			LeftVal:   lVal,
			RightVal:  rightVal,
			Operation: operator,
		}
		return operationNode
	}
}

// Run to whole dataset
func (s *OperationNode) Execute() (interface{}, error) {
	return s.runExecution(nil)
}

func (s *OperationNode) ExecuteSubset(indexes *series.Series) (interface{}, error) {
	return s.runExecution(indexes)
}

func (s *OperationNode) addNextOperation(nextOperation *OperationNode) *OperationNode {
	s.Next = nextOperation
	return s
}

func (s *OperationNode) getLeftValue(indexes *series.Series) (interface{}, error) {
	return getValue(s.LeftVal, indexes)
}

func (s *OperationNode) getRightValue(indexes *series.Series) (interface{}, error) {
	return getValue(s.RightVal, indexes)
}

func getValue(val interface{}, indexes *series.Series) (interface{}, error) {
	result := val
	var err error
	switch v := val.(type) {
	case *OperationNode:
		result, err = v.runExecution(indexes)
	case OperationNode:
		result, err = v.runExecution(indexes)
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *OperationNode) runExecution(indexes *series.Series) (interface{}, error) {
	node := s
	var result interface{}
	for ; node != nil; node = node.Next {
		leftVal, err := node.getLeftValue(indexes)
		if err != nil {
			return nil, err
		}
		rightVal, err := node.getRightValue(indexes)
		if err != nil {
			return nil, err
		}

		if leftVal == nil {
			leftVal = result
		}
		switch node.Operation {
		case operator.Greater:
			result, err = greaterOp(leftVal, rightVal, indexes)
		case operator.GreaterEq:
			result, err = greaterEqOp(leftVal, rightVal, indexes)
		case operator.Less:
			result, err = lessOp(leftVal, rightVal, indexes)
		case operator.LessEq:
			result, err = lessEqOp(leftVal, rightVal, indexes)
		case operator.Eq:
			result, err = equalOp(leftVal, rightVal, indexes)
		case operator.Neq:
			result, err = neqOp(leftVal, rightVal, indexes)
		case operator.And:
			result, err = andOp(leftVal, rightVal)
		case operator.Or:
			result, err = orOp(leftVal, rightVal)
		case operator.Add:
			result, err = addOp(leftVal, rightVal, indexes)
		case operator.Substract:
			result, err = substractOp(leftVal, rightVal, indexes)
		case operator.Multiply:
			result, err = multiplyOp(leftVal, rightVal, indexes)
		case operator.Divide:
			result, err = divideOp(leftVal, rightVal, indexes)
		case operator.Modulo:
			result, err = moduloOp(leftVal, rightVal, indexes)
		}
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func doOperationOnSeries(left interface{}, right interface{}, operation Operator, indexes *series.Series) (interface{}, error) {
	leftSeries, err := getOrCreateSeries(left)
	if err != nil {
		return nil, err
	}
	rightSeries, err := getOrCreateSeries(right)
	if err != nil {
		return nil, err
	}

	switch operation {
	case operator.Greater:
		return leftSeries.Greater(rightSeries)
	case operator.GreaterEq:
		return leftSeries.GreaterEq(rightSeries)
	case operator.Less:
		return leftSeries.Less(rightSeries)
	case operator.LessEq:
		return leftSeries.LessEq(rightSeries)
	case operator.Eq:
		return leftSeries.Eq(rightSeries)
	case operator.Neq:
		return leftSeries.Neq(rightSeries)
	case operator.And:
		return leftSeries.And(rightSeries)
	case operator.Or:
		return leftSeries.Or(rightSeries)
	case operator.Add:
		return leftSeries.Add(rightSeries, indexes)
	case operator.Substract:
		return leftSeries.Substract(rightSeries, indexes)
	case operator.Multiply:
		return leftSeries.Multiply(rightSeries, indexes)
	case operator.Divide:
		return leftSeries.Divide(rightSeries, indexes)
	case operator.Modulo:
		return leftSeries.Modulo(rightSeries, indexes)
	default:
		return nil, fmt.Errorf("%s operation is not supported for arithmetic", operation)
	}
}

func createSeries(value interface{}, colName string) (*series.Series, error) {
	return series.NewInferType(value, colName)
}

func getOrCreateSeries(value interface{}) (*series.Series, error) {
	var res *series.Series
	switch val := value.(type) {
	case series.Series:
		res = &val
	case *series.Series:
		res = val
	case nil:
		res = nil
	default:
		sr, err := createSeries(val, "leftSeries")
		if err != nil {
			return nil, err
		}
		res = sr
	}
	return res, nil
}
