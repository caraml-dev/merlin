package operation

import (
	"fmt"

	"github.com/gojek/merlin/pkg/transformer/types/series"
)

// Operator interface including arithmetic operator, logical operator and comparator
type Operator interface {
	Type() string
	Name() string
}

// OperationNode holds information about chain of operations
type OperationNode struct {
	LeftVal   interface{}
	RightVal  interface{}
	Operation Operator
	Next      *OperationNode
}

// RegisterOperation register chain of operation that already overriden
// Returning root of operation
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

// Execute run all the operations to whole row
func (s *OperationNode) Execute() (interface{}, error) {
	return s.runExecution(nil)
}

// ExecuteSubset run all the operations to subset of row in a series
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
		case Greater:
			result, err = greaterOp(leftVal, rightVal, indexes)
		case GreaterEq:
			result, err = greaterEqOp(leftVal, rightVal, indexes)
		case Less:
			result, err = lessOp(leftVal, rightVal, indexes)
		case LessEq:
			result, err = lessEqOp(leftVal, rightVal, indexes)
		case Eq:
			result, err = equalOp(leftVal, rightVal, indexes)
		case Neq:
			result, err = neqOp(leftVal, rightVal, indexes)
		case And:
			result, err = andOp(leftVal, rightVal)
		case Or:
			result, err = orOp(leftVal, rightVal)
		case Add:
			result, err = addOp(leftVal, rightVal, indexes)
		case Substract:
			result, err = substractOp(leftVal, rightVal, indexes)
		case Multiply:
			result, err = multiplyOp(leftVal, rightVal, indexes)
		case Divide:
			result, err = divideOp(leftVal, rightVal, indexes)
		case Modulo:
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
	case Greater:
		return leftSeries.Greater(rightSeries)
	case GreaterEq:
		return leftSeries.GreaterEq(rightSeries)
	case Less:
		return leftSeries.Less(rightSeries)
	case LessEq:
		return leftSeries.LessEq(rightSeries)
	case Eq:
		return leftSeries.Eq(rightSeries)
	case Neq:
		return leftSeries.Neq(rightSeries)
	case And:
		return leftSeries.And(rightSeries)
	case Or:
		return leftSeries.Or(rightSeries)
	case Add:
		return leftSeries.Add(rightSeries, indexes)
	case Substract:
		return leftSeries.Substract(rightSeries, indexes)
	case Multiply:
		return leftSeries.Multiply(rightSeries, indexes)
	case Divide:
		return leftSeries.Divide(rightSeries, indexes)
	case Modulo:
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
