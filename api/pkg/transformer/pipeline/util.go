package pipeline

import (
	"github.com/antonmedv/expr"

	mErrors "github.com/gojek/merlin/pkg/errors"
	"github.com/gojek/merlin/pkg/transformer/types/operation"
	"github.com/gojek/merlin/pkg/transformer/types/series"
)

func evalJSONPath(env *Environment, jsonPath string) (interface{}, error) {
	c := env.CompiledJSONPath(jsonPath)
	if c == nil {
		return nil, mErrors.NewInvalidInputErrorf("compiled jsonpath %s not found", jsonPath)
	}

	val, err := c.LookupFromContainer(env.PayloadContainer())
	if err != nil {
		return nil, mErrors.NewInvalidInputError(err.Error())
	}
	return val, nil
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
		return nil, mErrors.NewInvalidInputError(err.Error())
	}
	return val, nil
}

func seriesFromExpression(env *Environment, expression string) (*series.Series, error) {
	val, err := evalExpression(env, expression)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, mErrors.NewInvalidInputErrorf("series is empty due to expression %s returning nil", expression)
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
		return nil, mErrors.NewInvalidInputError(err.Error())
	}

	return series.NewInferType(val, "")
}

func getVal(env *Environment, expression string) (interface{}, error) {
	cplExpr := env.CompiledExpression(expression)
	if cplExpr == nil {
		return nil, mErrors.NewInvalidInputErrorf("compiled expression %s not found", expression)
	}

	val, err := expr.Run(env.CompiledExpression(expression), env.SymbolRegistry())
	if err != nil {
		return nil, mErrors.NewInvalidInputError(err.Error())
	}
	return val, nil
}
