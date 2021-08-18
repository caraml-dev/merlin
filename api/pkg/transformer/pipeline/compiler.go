package pipeline

import (
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/gojek/merlin/pkg/transformer/cache"
	"github.com/gojek/merlin/pkg/transformer/feast"
	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
	"github.com/gojek/merlin/pkg/transformer/types/expression"
	"github.com/gojek/merlin/pkg/transformer/types/table"
)

type Compiler struct {
	sr symbol.Registry

	feastClients feast.Clients
	feastOptions *feast.Options
	cacheOptions *cache.Options

	logger *zap.Logger
}

func NewCompiler(sr symbol.Registry, feastClients feast.Clients, feastOptions *feast.Options, cacheOptions *cache.Options, logger *zap.Logger) *Compiler {
	return &Compiler{sr: sr, feastClients: feastClients, feastOptions: feastOptions, cacheOptions: cacheOptions, logger: logger}
}

func (c *Compiler) Compile(spec *spec.StandardTransformerConfig) (*CompiledPipeline, error) {
	preprocessOps := make([]Op, 0)
	postprocessOps := make([]Op, 0)
	jsonPathStorage := jsonpath.NewStorage()
	expressionStorage := expression.NewStorage()
	if spec.TransformerConfig == nil {
		return NewCompiledPipeline(
			jsonPathStorage,
			expressionStorage,
			preprocessOps,
			postprocessOps,
		), nil
	}

	if spec.TransformerConfig.Preprocess != nil {
		ops, err := c.doCompilePipeline(spec.TransformerConfig.Preprocess, jsonPathStorage, expressionStorage)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to compile preprocessing pipeline")
		}
		preprocessOps = append(preprocessOps, ops...)
	}

	if spec.TransformerConfig.Postprocess != nil {
		ops, err := c.doCompilePipeline(spec.TransformerConfig.Postprocess, jsonPathStorage, expressionStorage)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to compile postprocessing pipeline")
		}
		postprocessOps = append(postprocessOps, ops...)
	}

	return NewCompiledPipeline(
		jsonPathStorage,
		expressionStorage,
		preprocessOps,
		postprocessOps,
	), nil
}

func (c *Compiler) doCompilePipeline(pipeline *spec.Pipeline, compiledJsonPaths *jsonpath.Storage, compiledExpressions *expression.Storage) ([]Op, error) {
	ops := make([]Op, 0)

	// input
	for _, input := range pipeline.Inputs {
		if input.Variables != nil {
			varOp, err := c.parseVariablesSpec(input.Variables, compiledJsonPaths, compiledExpressions)
			if err != nil {
				return nil, err
			}
			ops = append(ops, varOp)
		}

		if input.Tables != nil {
			tableOp, err := c.parseTablesSpec(input.Tables, compiledJsonPaths, compiledExpressions)
			if err != nil {
				return nil, err
			}
			ops = append(ops, tableOp)
		}

		if input.Feast != nil {
			feastOp, err := c.parseFeastSpec(input.Feast, compiledJsonPaths, compiledExpressions)
			if err != nil {
				return nil, err
			}
			ops = append(ops, feastOp)
		}
	}

	// transformation
	for _, transformation := range pipeline.Transformations {
		if transformation.TableTransformation != nil {
			tableTransformOps, err := c.parseTableTransform(transformation.TableTransformation, compiledJsonPaths, compiledExpressions)
			if err != nil {
				return nil, err
			}
			ops = append(ops, tableTransformOps)
		}

		if transformation.TableJoin != nil {
			tableJoinOp, err := c.parseTableJoin(transformation.TableJoin, compiledJsonPaths, compiledExpressions)
			if err != nil {
				return nil, err
			}
			ops = append(ops, tableJoinOp)
		}
	}

	// output stage
	for _, output := range pipeline.Outputs {
		if jsonOutput := output.JsonOutput; jsonOutput != nil {
			jsonOutputOp, err := c.parseJsonOutputSpec(jsonOutput, compiledJsonPaths, compiledExpressions)
			if err != nil {
				return nil, err
			}
			ops = append(ops, jsonOutputOp)
		}
	}

	return ops, nil
}

func (c *Compiler) parseVariablesSpec(variables []*spec.Variable, compiledJsonPaths *jsonpath.Storage, compiledExpressions *expression.Storage) (Op, error) {
	for _, variable := range variables {
		c.registerDummyVariable(variable.Name)
		switch v := variable.Value.(type) {
		case *spec.Variable_Literal:
			continue

		case *spec.Variable_Expression:
			compiledExpression, err := c.compileExpression(v.Expression)
			if err != nil {
				return nil, nil
			}
			compiledExpressions.Set(v.Expression, compiledExpression)

		case *spec.Variable_JsonPath:
			compiledJsonPath, err := jsonpath.Compile(v.JsonPath)
			if err != nil {
				return nil, nil
			}
			compiledJsonPaths.Set(v.JsonPath, compiledJsonPath)

		default:
			return nil, nil
		}
	}

	return NewVariableDeclarationOp(variables), nil
}

func (c *Compiler) parseFeastSpec(featureTableSpecs []*spec.FeatureTable, compiledJsonPaths *jsonpath.Storage, compiledExpressions *expression.Storage) (Op, error) {
	jsonPaths, err := feast.CompileJSONPaths(featureTableSpecs)
	if err != nil {
		return nil, err
	}
	compiledJsonPaths.AddAll(jsonPaths)

	expressions, err := feast.CompileExpressions(featureTableSpecs, c.sr)
	if err != nil {
		return nil, err
	}
	compiledExpressions.AddAll(expressions)

	for _, featureTableSpec := range featureTableSpecs {
		c.registerDummyTable(feast.GetTableName(featureTableSpec))
	}

	var memoryCache cache.Cache
	if c.feastOptions.CacheEnabled {
		memoryCache = cache.NewInMemoryCache(c.cacheOptions)
	}

	entityExtractor := feast.NewEntityExtractor(compiledJsonPaths, compiledExpressions)
	return NewFeastOp(c.feastClients, c.feastOptions, memoryCache, entityExtractor, featureTableSpecs, c.logger), nil
}

func (c *Compiler) parseTablesSpec(tableSpecs []*spec.Table, compiledJsonPaths *jsonpath.Storage, compiledExpressions *expression.Storage) (Op, error) {
	for _, tableSpec := range tableSpecs {
		c.registerDummyTable(tableSpec.Name)
		if tableSpec.BaseTable != nil {
			switch bt := tableSpec.BaseTable.BaseTable.(type) {
			case *spec.BaseTable_FromJson:
				compiledJsonPath, err := jsonpath.Compile(bt.FromJson.JsonPath)
				if err != nil {
					return nil, err
				}
				compiledJsonPaths.Set(bt.FromJson.JsonPath, compiledJsonPath)
			}
		}

		if tableSpec.Columns != nil {
			for _, column := range tableSpec.Columns {
				switch cv := column.ColumnValue.(type) {
				case *spec.Column_Expression:
					compiledExpression, err := c.compileExpression(cv.Expression)
					if err != nil {
						return nil, err
					}
					compiledExpressions.Set(cv.Expression, compiledExpression)
				case *spec.Column_FromJson:
					compiledJsonPath, err := jsonpath.Compile(cv.FromJson.JsonPath)
					if err != nil {
						return nil, err
					}
					compiledJsonPaths.Set(cv.FromJson.JsonPath, compiledJsonPath)
				}
			}
		}
	}

	return NewCreateTableOp(tableSpecs), nil
}

func (c *Compiler) parseTableTransform(transformationSpecs *spec.TableTransformation, paths *jsonpath.Storage, compiledExpressions *expression.Storage) (Op, error) {
	err := c.checkVariableRegistered(transformationSpecs.InputTable)
	if err != nil {
		return nil, err
	}

	for _, step := range transformationSpecs.Steps {
		for _, updateColumn := range step.UpdateColumns {
			compiledExpression, err := c.compileExpression(updateColumn.Expression)
			if err != nil {
				return nil, err
			}
			compiledExpressions.Set(updateColumn.Expression, compiledExpression)
		}
	}

	c.registerDummyTable(transformationSpecs.OutputTable)
	return NewTableTransformOp(transformationSpecs), nil
}

func (c *Compiler) parseTableJoin(tableJoinSpecs *spec.TableJoin, paths *jsonpath.Storage, expressions *expression.Storage) (Op, error) {
	err := c.checkVariableRegistered(tableJoinSpecs.LeftTable)
	if err != nil {
		return nil, err
	}

	err = c.checkVariableRegistered(tableJoinSpecs.RightTable)
	if err != nil {
		return nil, err
	}

	c.registerDummyTable(tableJoinSpecs.OutputTable)
	return NewTableJoinOp(tableJoinSpecs), nil
}

func (c *Compiler) parseJsonOutputSpec(jsonSpec *spec.JsonOutput, compiledJsonPaths *jsonpath.Storage, compiledExpressions *expression.Storage) (Op, error) {
	template := jsonSpec.JsonTemplate
	if template == nil {
		return nil, errors.New("jsontemplate must be specified")
	}
	if template.BaseJson != nil {
		compiledJsonPath, err := jsonpath.Compile(template.BaseJson.JsonPath)
		if err != nil {
			return nil, err
		}
		compiledJsonPaths.Set(template.BaseJson.JsonPath, compiledJsonPath)
	}

	if err := c.parseJsonFields(template.Fields, compiledJsonPaths, compiledExpressions); err != nil {
		return nil, err
	}

	jsonOutputOp := NewJsonOutputOp(jsonSpec)

	return jsonOutputOp, nil
}

func (c *Compiler) parseJsonFields(fields []*spec.Field, compiledJsonPaths *jsonpath.Storage, compiledExpressions *expression.Storage) error {
	for _, field := range fields {

		if field.FieldName == "" {
			return errors.New("field name must be specified")
		}
		switch val := field.Value.(type) {
		case *spec.Field_FromJson:
			if len(field.Fields) > 0 {
				return errors.New("can't specify nested json, if field has value to set")
			}
			compiledJsonPath, err := jsonpath.Compile(val.FromJson.JsonPath)
			if err != nil {
				return err
			}
			compiledJsonPaths.Set(val.FromJson.JsonPath, compiledJsonPath)
		case *spec.Field_FromTable:
			if len(field.Fields) > 0 {
				return errors.New("can't specify nested json, if field has value to set")
			}

			// Check whether format specified is valid or not
			if val.FromTable.Format == spec.FromTable_INVALID {
				return errors.New("table format specified is invalid")
			}

			// Check whether table name already registered or not
			tableName := val.FromTable.TableName
			if err := c.checkVariableRegistered(tableName); err != nil {
				return err
			}

		case *spec.Field_Expression:
			if len(field.Fields) > 0 {
				return errors.New("can't specify nested json, if field has value to set")
			}
			compiledExpression, err := c.compileExpression(val.Expression)
			if err != nil {
				return err
			}
			compiledExpressions.Set(val.Expression, compiledExpression)
		default:
			// value is not specified
			if len(field.Fields) == 0 {
				return errors.New("fields must be specified if value is empty")
			}
			return c.parseJsonFields(field.Fields, compiledJsonPaths, compiledExpressions)
		}
	}
	return nil
}

func (c *Compiler) compileExpression(expression string) (*vm.Program, error) {
	return expr.Compile(expression, expr.Env(c.sr))
}

func (c *Compiler) registerDummyVariable(varName string) {
	c.sr[varName] = nil
}

func (c *Compiler) registerDummyTable(tableName string) {
	c.sr[tableName] = table.New()
}

func (c *Compiler) checkVariableRegistered(varName string) error {
	isRegistered := c.sr[varName] != nil
	if !isRegistered {
		return fmt.Errorf("variable %s is not registered", varName)
	}

	return nil
}
