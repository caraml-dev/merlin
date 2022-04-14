package symbol

import (
	"github.com/gojek/merlin/pkg/transformer/types/operation"
	"github.com/gojek/merlin/pkg/transformer/types/series"
)

// GreaterOp is function that override default '>' operator that originally only applicable to numeric type
// This override method enable user to do operation '>' to series
// Returning chain of operation that can be executed lazily if the operation involving series or eagerly if operation only involving primitive values
func (sr Registry) GreaterOp(left, right interface{}) interface{} {
	return eval(left, right, operation.Greater)
}

// GreaterEqOp is function that override default '>=' operator that originally only applicable to numeric type
// This override method enable user to do operation '>=' to series
// Returning chain of operation that can be executed lazily if the operation involving series or eagerly if operation only involving primitive values
func (sr Registry) GreaterEqOp(left, right interface{}) interface{} {
	return eval(left, right, operation.GreaterEq)
}

// LessOp is function that override default '<' operator that originally only applicable to numeric type
// This override method enable user to do operation '<' to series
// Returning chain of operation that can be executed lazily if the operation involving series or eagerly if operation only involving primitive values
func (sr Registry) LessOp(left, right interface{}) interface{} {
	return eval(left, right, operation.Less)
}

// LessEqOp is function that override default '<=' operator that originally only applicable to numeric type
// This override method enable user to do operation '<=' to series
// Returning chain of operation that can be executed lazily if the operation involving series or eagerly if operation only involving primitive values
func (sr Registry) LessEqOp(left, right interface{}) interface{} {
	return eval(left, right, operation.LessEq)
}

// EqualOp is function that override default '==' operator that originally only applicable to primitive type
// This override method enable user to do operation '==' to series
// Returning chain of operation that can be executed lazily if the operation involving series or eagerly if operation only involving primitive values
func (sr Registry) EqualOp(left, right interface{}) interface{} {
	return eval(left, right, operation.Eq)
}

// NeqOp is function that override default '!=' operator that originally only applicable to primitive type
// This override method enable user to do operation '!=' to series
// Returning chain of operation that can be executed lazily if the operation involving series or eagerly if operation only involving primitive values
func (sr Registry) NeqOp(left, right interface{}) interface{} {
	return eval(left, right, operation.Neq)
}

// AndOp is function that override default '&&' operator that originally only applicable to boolean type
// This override method enable user to do operation '&&' to series
// Returning chain of operation that can be executed lazily if the operation involving series or eagerly if operation only involving primitive values
func (sr Registry) AndOp(left, right interface{}) interface{} {
	return eval(left, right, operation.And)
}

// OrOp is function that override default '||' operator that originally only applicable to boolean type
// This override method enable user to do operation '||' to series
// Returning chain of operation that can be executed lazily if the operation involving series or eagerly if operation only involving primitive values
func (sr Registry) OrOp(left, right interface{}) interface{} {
	return eval(left, right, operation.Or)
}

// AddOp is function that override default '+' operator that originally only applicable to numeric type
// This override method enable user to do operation '+' to series and string
// Returning chain of operation that can be executed lazily if the operation involving series or eagerly if operation only involving primitive values
func (sr Registry) AddOp(left, right interface{}) interface{} {
	return eval(left, right, operation.Add)
}

// SubstractOp is function that override default '-' operator that originally only applicable to numeric type
// This override method enable user to do operation '-' to series
// Returning chain of operation that can be executed lazily if the operation involving series or eagerly if operation only involving primitive values
func (sr Registry) SubstractOp(left, right interface{}) interface{} {
	return eval(left, right, operation.Substract)
}

// MultiplyOp is function that override default '*' operator that originally only applicable to numeric type
// This override method enable user to do operation '*' to series
// Returning chain of operation that can be executed lazily if the operation involving series or eagerly if operation only involving primitive values
func (sr Registry) MultiplyOp(left, right interface{}) interface{} {
	return eval(left, right, operation.Multiply)
}

// DivideOp is function that override default '/' operator that originally only applicable to numeric type
// This override method enable user to do operation '/' to series
// Returning chain of operation that can be executed lazily if the operation involving series or eagerly if operation only involving primitive values
func (sr Registry) DivideOp(left, right interface{}) interface{} {
	return eval(left, right, operation.Divide)
}

// ModuloOp is function that override default '%' operator that originally only applicable to integer type
// This override method enable user to do operation '%' to series
// Returning chain of operation that can be executed lazily if the operation involving series or eagerly if operation only involving primitive values
func (sr Registry) ModuloOp(left, right interface{}) interface{} {
	return eval(left, right, operation.Modulo)
}

func eval(left, right interface{}, operator operation.Operator) interface{} {
	op := operation.RegisterOperation(left, right, operator)
	if !(shouldEagerlyEvaluate(left) && shouldEagerlyEvaluate(right)) {
		return op
	}

	val, err := op.Execute()
	if err != nil {
		panic(err)
	}
	return val
}

func shouldEagerlyEvaluate(v interface{}) bool {
	switch v.(type) {
	case *series.Series, series.Series, *operation.OperationNode, operation.OperationNode:
		return false
	default:
		return true
	}
}
