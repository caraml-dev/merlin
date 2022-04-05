package pipeline

import (
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/gojek/merlin/pkg/transformer/types/operation"
	"github.com/gojek/merlin/pkg/transformer/types/series"
)

func evalJSONPath(env *Environment, jsonPath string) (interface{}, error) {
	c := env.CompiledJSONPath(jsonPath)
	if c == nil {
		return nil, fmt.Errorf("compiled jsonpath %s not found", jsonPath)
	}

	return c.LookupFromContainer(env.JSONContainer())
}

func evalExpression(env *Environment, expression string) (interface{}, error) {
	val, err := getVal(env, expression)
	if err != nil {
		return nil, err
	}
	switch exprVal := val.(type) {
	case *operation.OperationNode:
		val, err = exprVal.Execute()
	case operation.OperationNode:
		val, err = exprVal.Execute()
	}

	if err != nil {
		return nil, err
	}
	return val, nil
}

func seriesFromExpression(env *Environment, expression string) (*series.Series, error) {
	val, err := evalExpression(env, expression)
	if err != nil {
		return nil, err
	}
	return series.NewInferType(val, "")
}

func subsetSeriesFromExpression(env *Environment, expression string, subsetIdx *series.Series) (*series.Series, error) {
	val, err := getVal(env, expression)
	if err != nil {
		return nil, err
	}
	switch exprVal := val.(type) {
	case *operation.OperationNode:
		val, err = exprVal.ExecuteSubset(subsetIdx)
	case operation.OperationNode:
		val, err = exprVal.ExecuteSubset(subsetIdx)
	}

	if err != nil {
		return nil, err
	}

	return series.NewInferType(val, "")
}

func getVal(env *Environment, expression string) (interface{}, error) {
	cplExpr := env.CompiledExpression(expression)
	if cplExpr == nil {
		return nil, fmt.Errorf("compiled expression %s not found", expression)
	}

	return expr.Run(env.CompiledExpression(expression), env.SymbolRegistry())
}
