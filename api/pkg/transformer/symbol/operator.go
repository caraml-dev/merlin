package symbol

import (
	"github.com/gojek/merlin/pkg/transformer/types/operation"
	"github.com/gojek/merlin/pkg/transformer/types/operator"
)

func (sr Registry) GreaterOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operator.Greater)
}

func (sr Registry) GreaterEqOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operator.GreaterEq)
}

func (sr Registry) LessOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operator.Less)
}

func (sr Registry) LessEqOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operator.LessEq)
}

func (sr Registry) EqualOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operator.Eq)
}

func (sr Registry) NeqOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operator.Neq)
}

func (sr Registry) AndOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operator.And)
}

func (sr Registry) OrOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operator.Or)
}

func (sr Registry) AddOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operator.Add)
}

func (sr Registry) SubstractOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operator.Substract)
}

func (sr Registry) MultiplyOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operator.Multiply)
}

func (sr Registry) DivideOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operator.Divide)
}

func (sr Registry) ModuloOp(left, right interface{}) interface{} {
	return operation.RegisterOperation(left, right, operator.Modulo)
}
