package pipeline

import (
	"context"
	"fmt"

	"github.com/antonmedv/expr"

	"github.com/gojek/merlin/pkg/transformer/spec"
)

type VariableDeclarationOp struct {
	variableSpec []*spec.Variable
}

func NewVariableDeclarationOp(variables []*spec.Variable) Op {
	return &VariableDeclarationOp{
		variableSpec: variables,
	}
}

func (v *VariableDeclarationOp) Execute(context context.Context, env *Environment) error {
	for _, varDef := range v.variableSpec {
		switch v := varDef.Value.(type) {
		case *spec.Variable_Literal:
			switch val := v.Literal.LiteralValue.(type) {
			case *spec.Literal_IntValue:
				env.SetSymbol(varDef.Name, val.IntValue)
			case *spec.Literal_FloatValue:
				env.SetSymbol(varDef.Name, val.FloatValue)
			case *spec.Literal_StringValue:
				env.SetSymbol(varDef.Name, val.StringValue)
			}

		case *spec.Variable_Expression:
			cplExpr := env.CompiledExpression(v.Expression)
			if cplExpr == nil {
				return fmt.Errorf("compiled expression %s not found", v.Expression)
			}

			result, err := expr.Run(env.CompiledExpression(v.Expression), env.SymbolRegistry())
			if err != nil {
				return err
			}
			env.SetSymbol(varDef.Name, result)
		case *spec.Variable_JsonPath:
			jsonPath := env.CompiledJSONPath(v.JsonPath)
			if jsonPath == nil {
				return fmt.Errorf("compiled jsonpath %s not found", v.JsonPath)
			}

			result, err := jsonPath.LookupFromContainer(env.JSONContainer())
			if err != nil {
				return nil
			}
			env.SetSymbol(varDef.Name, result)
		default:
			return fmt.Errorf("Variable.Value has unexpected type %T", v)
		}
	}

	return nil
}
