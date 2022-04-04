package symbol

import (
	"github.com/gojek/merlin/pkg/transformer/types/operation"
)

// GreaterOp is function that override default '>' operator that originally only applicable to numeric type
// This override method enable user to do operation '>' to series
// Returning chain of operation that needs to be executed later
func (sr Registry) GreaterOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.Greater)
}

// GreaterEqOp is function that override default '>=' operator that originally only applicable to numeric type
// This override method enable user to do operation '>=' to series
// Returning chain of operation that needs to be executed later
func (sr Registry) GreaterEqOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.GreaterEq)
}

// LessOp is function that override default '<' operator that originally only applicable to numeric type
// This override method enable user to do operation '<' to series
// Returning chain of operation that needs to be executed later
func (sr Registry) LessOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.Less)
}

// LessEqOp is function that override default '<=' operator that originally only applicable to numeric type
// This override method enable user to do operation '<=' to series
// Returning chain of operation that needs to be executed later
func (sr Registry) LessEqOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.LessEq)
}

// EqualOp is function that override default '==' operator that originally only applicable to primitive type
// This override method enable user to do operation '==' to series
// Returning chain of operation that needs to be executed later
func (sr Registry) EqualOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.Eq)
}

// NeqOp is function that override default '!=' operator that originally only applicable to primitive type
// This override method enable user to do operation '!=' to series
// Returning chain of operation that needs to be executed later
func (sr Registry) NeqOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.Neq)
}

// AndOp is function that override default '&&' operator that originally only applicable to boolean type
// This override method enable user to do operation '&&' to series
// Returning chain of operation that needs to be executed later
func (sr Registry) AndOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.And)
}

// OrOp is function that override default '||' operator that originally only applicable to boolean type
// This override method enable user to do operation '||' to series
// Returning chain of operation that needs to be executed later
func (sr Registry) OrOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.Or)
}

// AddOp is function that override default '+' operator that originally only applicable to numeric type
// This override method enable user to do operation '+' to series and string
// Returning chain of operation that needs to be executed later
func (sr Registry) AddOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.Add)
}

// SubstractOp is function that override default '-' operator that originally only applicable to numeric type
// This override method enable user to do operation '-' to series
// Returning chain of operation that needs to be executed later
func (sr Registry) SubstractOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.Substract)
}

// MultiplyOp is function that override default '*' operator that originally only applicable to numeric type
// This override method enable user to do operation '*' to series
// Returning chain of operation that needs to be executed later
func (sr Registry) MultiplyOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.Multiply)
}

// DivideOp is function that override default '/' operator that originally only applicable to numeric type
// This override method enable user to do operation '/' to series
// Returning chain of operation that needs to be executed later
func (sr Registry) DivideOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.Divide)
}

// ModuloOp is function that override default '%' operator that originally only applicable to integer type
// This override method enable user to do operation '%' to series
// Returning chain of operation that needs to be executed later
func (sr Registry) ModuloOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.Modulo)
}
