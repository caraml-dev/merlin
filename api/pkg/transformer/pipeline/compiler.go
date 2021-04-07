package pipeline

import (
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/pkg/errors"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/spec"
)

type Compiler struct {
	sr *SymbolRegistry
}

func NewCompiler(sr *SymbolRegistry) *Compiler {
	return &Compiler{
		sr: sr,
	}
}

func (c *Compiler) Compile(spec *spec.StandardTransformerConfig) (*CompiledPipeline, error) {
	preprocessOps, preprocessJSONPath, preprocessExpressions, err := c.doCompilePipeline(spec.TransformerConfig.Preprocess)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to compile preprocessing pipeline")
	}

	postprocessOps, postprocessJSONPath, postprocessExpressions, err := c.doCompilePipeline(spec.TransformerConfig.Postprocess)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to compile postprocessing pipeline")
	}

	// merge compiled jsonpath
	preprocessJSONPath = mergeJSONPathMap(preprocessJSONPath, postprocessJSONPath)

	// merge compiled expressions
	preprocessExpressions = mergeExpressionMap(preprocessExpressions, postprocessExpressions)

	return &CompiledPipeline{
		compiledJsonpath:   preprocessJSONPath,
		compiledExpression: preprocessExpressions,
		preprocessOps:      preprocessOps,
		postprocessOps:     postprocessOps,
	}, nil
}

func (c *Compiler) doCompilePipeline(pipeline *spec.Pipeline) (ops []Op, compiledJsonPaths map[string]*jsonpath.CompiledJSONPath, compiledExpressions map[string]*vm.Program, err error) {
	ops = make([]Op, 0)
	compiledJsonPaths = make(map[string]*jsonpath.CompiledJSONPath)
	compiledExpressions = make(map[string]*vm.Program)

	// input
	for _, input := range pipeline.Inputs {
		if input.Variables != nil {
			varOp, cplJsonPaths, cplExpressions, err := c.parseVariablesSpec(input.Variables)
			if err != nil {
				return nil, nil, nil, err
			}
			ops = append(ops, varOp)
			compiledJsonPaths = mergeJSONPathMap(compiledJsonPaths, cplJsonPaths)
			compiledExpressions = mergeExpressionMap(compiledExpressions, cplExpressions)
		}

		if input.Tables != nil {
			tableOp, cplJsonPaths, cplExpressions, err := c.parseTablesSpec(input.Tables)
			if err != nil {
				return nil, nil, nil, err
			}
			ops = append(ops, tableOp)
			compiledJsonPaths = mergeJSONPathMap(compiledJsonPaths, cplJsonPaths)
			compiledExpressions = mergeExpressionMap(compiledExpressions, cplExpressions)
		}

		if input.Feast != nil {
			feastOp, cplJsonPaths, cplExpressions, err := c.parseFeastSpec(input.Feast)
			if err != nil {
				return nil, nil, nil, err
			}
			ops = append(ops, feastOp)
			compiledJsonPaths = mergeJSONPathMap(compiledJsonPaths, cplJsonPaths)
			compiledExpressions = mergeExpressionMap(compiledExpressions, cplExpressions)
		}
	}

	// TODO: transformation

	// TODO: output

	return
}

func (c *Compiler) parseVariablesSpec(variables []*spec.Variable) (Op, map[string]*jsonpath.CompiledJSONPath, map[string]*vm.Program, error) {
	compiledJsonPaths := make(map[string]*jsonpath.CompiledJSONPath)
	compiledExpressions := make(map[string]*vm.Program)

	for _, variable := range variables {
		switch v := variable.Value.(type) {
		case *spec.Variable_Literal:
			continue

		case *spec.Variable_Expression:
			compiledExpression, err := c.compileExpression(v.Expression)
			if err != nil {
				return nil, nil, nil, errors.Wrapf(err, "unable to compile expression %s for variable %s", v.Expression, variable.Name)
			}
			compiledExpressions[v.Expression] = compiledExpression

		case *spec.Variable_JsonPath:
			compiledJsonPath, err := jsonpath.CompileJsonPath(v.JsonPath)
			if err != nil {
				return nil, nil, nil, errors.Wrapf(err, "unable to compile jsonpath %s for variable %s", v.JsonPath, variable.Name)
			}
			compiledJsonPaths[v.JsonPath] = compiledJsonPath

		default:
			return nil, nil, nil, fmt.Errorf("variable %s has unexpected type %T, it should be Expression, Literal, or JsonPath", variable.Name, v)
		}
	}

	return NewVariableDeclarationOp(variables), compiledJsonPaths, compiledExpressions, nil
}

func (c *Compiler) parseFeastSpec(featureTables []*spec.FeatureTable) (Op, map[string]*jsonpath.CompiledJSONPath, map[string]*vm.Program, error) {
	compiledJsonPaths := make(map[string]*jsonpath.CompiledJSONPath)
	compiledExpressions := make(map[string]*vm.Program)

	for _, featureTable := range featureTables {
		for _, entity := range featureTable.Entities {
			switch extractor := entity.Extractor.(type) {
			case *spec.Entity_Udf:
				compiledExpression, err := c.compileExpression(entity.GetUdf())
				if err != nil {
					return nil, nil, nil, errors.Wrapf(err, "unable to compile expression %s for variable %s", entity.GetUdf(), entity.Name)
				}
				compiledExpressions[entity.GetUdf()] = compiledExpression
			case *spec.Entity_Expression:
				compiledExpression, err := c.compileExpression(entity.GetExpression())
				if err != nil {
					return nil, nil, nil, errors.Wrapf(err, "unable to compile expression %s for variable %s", entity.GetExpression(), entity.Name)
				}
				compiledExpressions[entity.GetExpression()] = compiledExpression
			case *spec.Entity_JsonPath:
				compiledJsonPath, err := jsonpath.CompileJsonPath(extractor.JsonPath)
				if err != nil {
					return nil, nil, nil, errors.Wrapf(err, "unable to compile jsonpath %s for variable %s", extractor.JsonPath, entity.Name)
				}
				compiledJsonPaths[extractor.JsonPath] = compiledJsonPath
			default:
				return nil, nil, nil, fmt.Errorf("entity %s has unexpected type %T, it should be Expression, Udf, or JsonPath", entity.Name, extractor)
			}
		}
	}

	return NewFeastOp(featureTables), compiledJsonPaths, compiledExpressions, nil
}

func (c *Compiler) parseTablesSpec(tables []*spec.Table) (Op, map[string]*jsonpath.CompiledJSONPath, map[string]*vm.Program, error) {
	panic("TODO: parseTablesSpec")
}

func (c *Compiler) compileExpression(expression string) (*vm.Program, error) {
	return expr.Compile(expression, expr.Env(c.sr), expr.AllowUndefinedVariables())
}

func mergeJSONPathMap(dst, src map[string]*jsonpath.CompiledJSONPath) map[string]*jsonpath.CompiledJSONPath {
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func mergeExpressionMap(dst, src map[string]*vm.Program) map[string]*vm.Program {
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
