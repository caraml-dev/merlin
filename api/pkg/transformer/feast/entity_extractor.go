package feast

import (
	"fmt"

	"github.com/antonmedv/expr"
	feastType "github.com/feast-dev/feast/sdk/go/protos/feast/types"

	"github.com/caraml-dev/merlin/pkg/transformer/jsonpath"
	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/pkg/transformer/symbol"
	"github.com/caraml-dev/merlin/pkg/transformer/types/converter"
	"github.com/caraml-dev/merlin/pkg/transformer/types/expression"
	"github.com/caraml-dev/merlin/pkg/transformer/types/series"
)

// EntityExtractor is responsible to extract entity values from symbol registry
type EntityExtractor struct {
	compiledJsonPath   *jsonpath.Storage
	compiledExpression *expression.Storage
}

func NewEntityExtractor(compiledJsonPath *jsonpath.Storage, compiledExpression *expression.Storage) *EntityExtractor {
	return &EntityExtractor{
		compiledJsonPath:   compiledJsonPath,
		compiledExpression: compiledExpression,
	}
}

// ExtractValuesFromSymbolRegistry extracts entity values from symbol registry
// Which means it can be used to extract entity values from json object, variables, expression, or table.
func (er *EntityExtractor) ExtractValuesFromSymbolRegistry(symbolRegistry symbol.Registry, entitySpec *spec.Entity) ([]*feastType.Value, error) {
	feastValType := feastType.ValueType_Enum(feastType.ValueType_Enum_value[entitySpec.ValueType])

	var entityVal interface{}
	switch entitySpec.Extractor.(type) {
	case *spec.Entity_JsonPath:
		compiledJsonPath := er.compiledJsonPath.Get(entitySpec.GetJsonPath())
		if compiledJsonPath == nil {
			return nil, fmt.Errorf("jsonpath %s in entity %s is not found", entitySpec.GetJsonPath(), entitySpec.Name)
		}

		entityValFromJsonPath, err := compiledJsonPath.LookupFromContainer(symbolRegistry.PayloadContainer())
		if err != nil {
			return nil, err
		}
		entityVal = entityValFromJsonPath
	case *spec.Entity_JsonPathConfig:
		compiledJsonPath := er.compiledJsonPath.Get(entitySpec.GetJsonPathConfig().GetJsonPath())
		if compiledJsonPath == nil {
			return nil, fmt.Errorf("jsonpath %s in entity %s is not found", entitySpec.GetJsonPathConfig().GetJsonPath(), entitySpec.Name)
		}

		entityValFromJsonPath, err := compiledJsonPath.LookupFromContainer(symbolRegistry.PayloadContainer())
		if err != nil {
			return nil, err
		}
		entityVal = entityValFromJsonPath
	case *spec.Entity_Udf, *spec.Entity_Expression:
		exp := getExpressionExtractor(entitySpec)
		compiledExpression := er.compiledExpression.Get(exp)
		if compiledExpression == nil {
			return nil, fmt.Errorf("expression %s in entity %s is not found", exp, entitySpec.Name)
		}

		exprResult, err := expr.Run(compiledExpression, symbolRegistry)
		if err != nil {
			return nil, err
		}

		entityVal = exprResult
	}

	switch val := entityVal.(type) {
	case *series.Series:
		records := val.GetRecords()
		vals := make([]*feastType.Value, 0)
		for _, v := range records {
			if v == nil {
				continue
			}
			nv, err := converter.ToFeastValue(v, feastValType)
			if err != nil {
				return nil, err
			}
			vals = append(vals, nv)
		}
		return vals, nil
	case []interface{}:
		vals := make([]*feastType.Value, 0)
		for _, v := range val {
			nv, err := converter.ToFeastValue(v, feastValType)
			if err != nil {
				return nil, err
			}
			vals = append(vals, nv)
		}
		return vals, nil
	case interface{}:
		v, err := converter.ToFeastValue(val, feastValType)
		if err != nil {
			return nil, err
		}
		return []*feastType.Value{v}, nil

	default:
		return nil, fmt.Errorf("unknown value type: %T", entityVal)
	}
}

func getExpressionExtractor(entitySpec *spec.Entity) string {
	if extractor := entitySpec.GetExpression(); extractor != "" {
		return extractor
	}
	return entitySpec.GetUdf()
}
