package pipeline

import (
	"context"
	"fmt"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/opentracing/opentracing-go"
)

type VariableDeclarationOp struct {
	variableSpec []*spec.Variable
	*OperationTracing
}

func NewVariableDeclarationOp(variables []*spec.Variable, tracingEnabled bool) Op {
	varOp := &VariableDeclarationOp{
		variableSpec: variables,
	}

	if tracingEnabled {
		varOp.OperationTracing = NewOperationTracing(variables, types.VariableOpType)
	}
	return varOp
}

func (v *VariableDeclarationOp) Execute(context context.Context, env *Environment) error {
	span, _ := opentracing.StartSpanFromContext(context, "pipeline.VariableOp")
	defer span.Finish()

	for _, varDef := range v.variableSpec {
		name := varDef.Name

		var value interface{}
		switch v := varDef.Value.(type) {
		case *spec.Variable_Literal:
			switch val := v.Literal.LiteralValue.(type) {
			case *spec.Literal_IntValue:
				value = val.IntValue
			case *spec.Literal_FloatValue:
				value = val.FloatValue
			case *spec.Literal_StringValue:
				value = val.StringValue
			case *spec.Literal_BoolValue:
				value = val.BoolValue
			default:
				return fmt.Errorf("Variable.Literal.LiteralValue has unexpected type %T", v)
			}

		case *spec.Variable_Expression:
			result, err := evalExpression(env, v.Expression)
			if err != nil {
				return err
			}
			value = result

		case *spec.Variable_JsonPath:
			result, err := evalJSONPath(env, v.JsonPath)
			if err != nil {
				return nil
			}
			value = result

		case *spec.Variable_JsonPathConfig:
			result, err := evalJSONPath(env, v.JsonPathConfig.JsonPath)
			if err != nil {
				return nil
			}
			value = result

		default:
			return fmt.Errorf("Variable.Value has unexpected type %T", v)
		}

		env.SetSymbol(name, value)
		if v.OperationTracing != nil {
			if err := v.AddInputOutput(nil, map[string]interface{}{name: value}); err != nil {
				return err
			}
		}
		env.LogOperation("set variable", varDef.Name)
	}

	return nil
}
