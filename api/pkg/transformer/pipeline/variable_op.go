package pipeline

import (
	"context"
	"fmt"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/opentracing/opentracing-go"
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
	span, _ := opentracing.StartSpanFromContext(context, "pipeline.VariableOp")
	defer span.Finish()

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
			case *spec.Literal_BoolValue:
				env.SetSymbol(varDef.Name, val.BoolValue)
			}

		case *spec.Variable_Expression:
			result, err := evalExpression(env, v.Expression)
			if err != nil {
				return err
			}
			env.SetSymbol(varDef.Name, result)
		case *spec.Variable_JsonPath:
			result, err := evalJSONPath(env, v.JsonPath)
			if err != nil {
				return nil
			}
			env.SetSymbol(varDef.Name, result)
		default:
			return fmt.Errorf("Variable.Value has unexpected type %T", v)
		}

		env.LogOperation("set variable", varDef.Name)
	}

	return nil
}
