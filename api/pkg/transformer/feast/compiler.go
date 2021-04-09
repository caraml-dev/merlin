package feast

import (
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	feastTypes "github.com/feast-dev/feast/sdk/go/protos/feast/types"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
)

func CompileJSONPaths(featureTableSpecs []*spec.FeatureTable) (map[string]*jsonpath.Compiled, error) {
	compiledJsonPath := make(map[string]*jsonpath.Compiled)
	for _, ft := range featureTableSpecs {
		for _, configEntity := range ft.Entities {
			switch configEntity.Extractor.(type) {
			case *spec.Entity_JsonPath:
				c, err := jsonpath.Compile(configEntity.GetJsonPath())
				if err != nil {
					return nil, fmt.Errorf("unable to compile jsonpath for entity %s: %s", configEntity.Name, configEntity.GetJsonPath())
				}
				compiledJsonPath[configEntity.GetJsonPath()] = c
			default:
				continue
			}
		}
	}
	return compiledJsonPath, nil
}

func CompileExpressions(featureTableSpecs []*spec.FeatureTable) (map[string]*vm.Program, error) {
	compiledExpression := make(map[string]*vm.Program)
	for _, ft := range featureTableSpecs {
		for _, configEntity := range ft.Entities {
			switch configEntity.Extractor.(type) {
			case *spec.Entity_Udf, *spec.Entity_Expression:
				expressionExtractor := getExpressionExtractor(configEntity)
				c, err := expr.Compile(expressionExtractor, expr.Env(symbol.NewRegistry()), expr.AllowUndefinedVariables())
				if err != nil {
					return nil, err
				}
				compiledExpression[expressionExtractor] = c
			default:
				continue
			}
		}
	}

	return compiledExpression, nil
}

func compileDefaultValues(featureTableSpecs []*spec.FeatureTable) map[string]*feastTypes.Value {
	defaultValues := make(map[string]*feastTypes.Value)
	// populate default values
	for _, ft := range featureTableSpecs {
		for _, f := range ft.Features {
			if len(f.DefaultValue) != 0 {
				feastValType := feastTypes.ValueType_Enum(feastTypes.ValueType_Enum_value[f.ValueType])
				defVal, err := getValue(f.DefaultValue, feastValType)
				if err != nil {
					continue
				}
				defaultValues[f.Name] = defVal
			}
		}
	}

	return defaultValues
}
