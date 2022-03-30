package symbol

import (
	"github.com/gojek/merlin/pkg/transformer/types/operation"
)

func (sr Registry) GreaterOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.Greater)
}

func (sr Registry) GreaterEqOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.GreaterEq)
}

func (sr Registry) LessOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.Less)
}

func (sr Registry) LessEqOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.LessEq)
}

func (sr Registry) EqualOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.Eq)
}

func (sr Registry) NeqOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.Neq)
}

func (sr Registry) AndOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.And)
}

func (sr Registry) OrOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.Or)
}

func (sr Registry) AddOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.Add)
}

func (sr Registry) SubstractOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.Substract)
}

func (sr Registry) MultiplyOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.Multiply)
}

func (sr Registry) DivideOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.Divide)
}

func (sr Registry) ModuloOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operation.Modulo)
}
