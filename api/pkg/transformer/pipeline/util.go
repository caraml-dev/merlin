package pipeline

import (
	"fmt"

	"github.com/antonmedv/expr"
)

func evalJSONPath(env *Environment, jsonPath string) (interface{}, error) {
	c := env.CompiledJSONPath(jsonPath)
	if c == nil {
		return nil, fmt.Errorf("compiled jsonpath %s not found", jsonPath)
	}

	return c.LookupFromContainer(env.JSONContainer())
}

func evalExpression(env *Environment, expression string) (interface{}, error) {
	cplExpr := env.CompiledExpression(expression)
	if cplExpr == nil {
		return nil, fmt.Errorf("compiled expression %s not found", expression)
	}

	return expr.Run(env.CompiledExpression(expression), env.SymbolRegistry())
}
