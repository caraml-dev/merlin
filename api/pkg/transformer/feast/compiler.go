package feast

import (
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"

	"github.com/caraml-dev/merlin/pkg/transformer/jsonpath"
	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/pkg/transformer/symbol"
)

func CompileJSONPaths(featureTableSpecs []*spec.FeatureTable, jsonpathSourceType jsonpath.SourceType) (map[string]*jsonpath.Compiled, error) {
	compiledJsonPath := make(map[string]*jsonpath.Compiled)
	for _, ft := range featureTableSpecs {
		for _, configEntity := range ft.Entities {
			switch configEntity.Extractor.(type) {
			case *spec.Entity_JsonPath:
				c, err := jsonpath.CompileWithOption(jsonpath.JsonPathOption{
					JsonPath: configEntity.GetJsonPath(),
					SrcType:  jsonpathSourceType,
				})
				if err != nil {
					return nil, fmt.Errorf("unable to compile jsonpath for entity %s: %s", configEntity.Name, configEntity.GetJsonPath())
				}
				compiledJsonPath[configEntity.GetJsonPath()] = c
			case *spec.Entity_JsonPathConfig:
				jsonPathCfg := configEntity.GetJsonPathConfig()
				c, err := jsonpath.CompileWithOption(jsonpath.JsonPathOption{
					JsonPath:     jsonPathCfg.JsonPath,
					DefaultValue: jsonPathCfg.DefaultValue,
					TargetType:   jsonPathCfg.ValueType,
					SrcType:      jsonpathSourceType,
				})
				if err != nil {
					return nil, fmt.Errorf("unable to compile jsonpath config for entity %s: %s. err: %w", configEntity.Name, jsonPathCfg, err)
				}
				compiledJsonPath[jsonPathCfg.JsonPath] = c
			default:
				continue
			}
		}
	}
	return compiledJsonPath, nil
}

func CompileExpressions(featureTableSpecs []*spec.FeatureTable, symbolRegistry symbol.Registry) (map[string]*vm.Program, error) {
	compiledExpression := make(map[string]*vm.Program)
	for _, ft := range featureTableSpecs {
		for _, configEntity := range ft.Entities {
			switch configEntity.Extractor.(type) {
			case *spec.Entity_Udf, *spec.Entity_Expression:
				expressionExtractor := getExpressionExtractor(configEntity)
				c, err := expr.Compile(expressionExtractor, expr.Env(symbolRegistry))
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
