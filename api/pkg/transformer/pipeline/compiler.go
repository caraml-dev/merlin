package pipeline

import (
	"fmt"
	"net/url"
	"os"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	gota "github.com/go-gota/gota/series"
	ptc "github.com/gojek/merlin/pkg/protocol"
	"github.com/gojek/merlin/pkg/transformer/feast"
	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/pkg/transformer/types/expression"
	"github.com/gojek/merlin/pkg/transformer/types/scaler"
	"github.com/gojek/merlin/pkg/transformer/types/table"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	artifactsFolder        = "/mnt/models/artifacts"
	envPredictorStorageURI = "STORAGE_URI"
)

type Compiler struct {
	sr symbol.Registry

	feastClients feast.Clients
	feastOptions *feast.Options

	logger                  *zap.Logger
	operationTracingEnabled bool
	transformerValidationFn func(*spec.StandardTransformerConfig) error
	jsonpathSourceType      jsonpath.SourceType
}

func NewCompiler(sr symbol.Registry,
	feastClients feast.Clients,
	feastOptions *feast.Options,
	logger *zap.Logger,
	tracingEnabled bool,
	protocol ptc.Protocol) *Compiler {

	var validationFn func(*spec.StandardTransformerConfig) error
	jsonpathSourceType := jsonpath.Map
	if protocol == ptc.HttpJson {
		validationFn = httpTransformerValidation
	} else {
		validationFn = upiTransformerValidation
		jsonpathSourceType = jsonpath.Proto
	}
	return &Compiler{
		sr:                      sr,
		feastClients:            feastClients,
		feastOptions:            feastOptions,
		logger:                  logger,
		operationTracingEnabled: tracingEnabled,
		transformerValidationFn: validationFn,
		jsonpathSourceType:      jsonpathSourceType,
	}
}

func (c *Compiler) Compile(spec *spec.StandardTransformerConfig) (*CompiledPipeline, error) {
	preprocessOps := make([]Op, 0)
	postprocessOps := make([]Op, 0)
	preloadedTables := map[string]table.Table{}
	jsonPathStorage := jsonpath.NewStorage()
	expressionStorage := expression.NewStorage()
	if spec.TransformerConfig == nil {
		return NewCompiledPipeline(
			jsonPathStorage,
			expressionStorage,
			preloadedTables,
			preprocessOps,
			postprocessOps,
			c.operationTracingEnabled,
		), nil
	}

	if err := c.transformerValidationFn(spec); err != nil {
		return nil, err
	}

	if spec.TransformerConfig.Preprocess != nil {
		ops, loadedTables, err := c.doCompilePipeline(spec.TransformerConfig.Preprocess, types.Preprocess, jsonPathStorage, expressionStorage)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to compile preprocessing pipeline")
		}
		preprocessOps = append(preprocessOps, ops...)
		for k, v := range loadedTables {
			preloadedTables[k] = v
		}
	}

	if spec.TransformerConfig.Postprocess != nil {
		ops, loadedTables, err := c.doCompilePipeline(spec.TransformerConfig.Postprocess, types.Postprocess, jsonPathStorage, expressionStorage)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to compile postprocessing pipeline")
		}
		postprocessOps = append(postprocessOps, ops...)
		for k, v := range loadedTables {
			if _, ok := preloadedTables[k]; ok {
				return nil, fmt.Errorf("table name %s in postprocess already used in preprocess", k)
			}
			preloadedTables[k] = v
		}
	}

	return NewCompiledPipeline(
		jsonPathStorage,
		expressionStorage,
		preloadedTables,
		preprocessOps,
		postprocessOps,
		c.operationTracingEnabled,
	), nil
}

func (c *Compiler) doCompilePipeline(pipeline *spec.Pipeline, pipelineType types.Pipeline, compiledJsonPaths *jsonpath.Storage, compiledExpressions *expression.Storage) ([]Op, map[string]table.Table, error) {
	ops := make([]Op, 0)
	preloadedTables := map[string]table.Table{}

	// input
	for _, input := range pipeline.Inputs {
		if input.Variables != nil {
			varOp, err := c.parseVariablesSpec(input.Variables, compiledJsonPaths, compiledExpressions)
			if err != nil {
				return nil, nil, err
			}
			ops = append(ops, varOp)
		}

		if input.Tables != nil {
			tableOp, loadedTables, err := c.parseTablesSpec(input.Tables, compiledJsonPaths, compiledExpressions)
			if err != nil {
				return nil, nil, err
			}

			ops = append(ops, tableOp)
			for k, v := range loadedTables {
				preloadedTables[k] = v
				if c.operationTracingEnabled {
					if err := tableOp.AddInputOutput(nil, map[string]interface{}{k: &v}); err != nil {
						return nil, nil, err
					}
				}
			}
		}

		if input.Feast != nil {
			feastOp, err := c.parseFeastSpec(input.Feast, compiledJsonPaths, compiledExpressions)
			if err != nil {
				return nil, nil, err
			}
			ops = append(ops, feastOp)
		}

		if input.Encoders != nil {
			encoderOp, err := c.parseEncodersSpec(input.Encoders, compiledExpressions)
			if err != nil {
				return nil, nil, err
			}
			ops = append(ops, encoderOp)
		}
		if input.Autoload != nil {
			autoloadOp, err := c.parseUPIAutoloadSpec(input.Autoload, pipelineType, compiledExpressions)
			if err != nil {
				return nil, nil, err
			}
			ops = append(ops, autoloadOp)
		}
	}

	// transformation
	for _, transformation := range pipeline.Transformations {
		if transformation.TableTransformation != nil {
			tableTransformOps, err := c.parseTableTransform(transformation.TableTransformation, compiledJsonPaths, compiledExpressions)
			if err != nil {
				return nil, nil, err
			}
			ops = append(ops, tableTransformOps)
		}

		if transformation.TableJoin != nil {
			tableJoinOp, err := c.parseTableJoin(transformation.TableJoin, compiledJsonPaths, compiledExpressions)
			if err != nil {
				return nil, nil, err
			}
			ops = append(ops, tableJoinOp)
		}
		if len(transformation.Variables) > 0 {
			varOp, err := c.parseVariablesSpec(transformation.Variables, compiledJsonPaths, compiledExpressions)
			if err != nil {
				return nil, nil, err
			}
			ops = append(ops, varOp)
		}
	}

	// output stage
	for _, output := range pipeline.Outputs {
		if jsonOutput := output.JsonOutput; jsonOutput != nil {
			jsonOutputOp, err := c.parseJsonOutputSpec(jsonOutput, compiledJsonPaths, compiledExpressions)
			if err != nil {
				return nil, nil, err
			}
			ops = append(ops, jsonOutputOp)
		}
		if upiPreprocessOutput := output.UpiPreprocessOutput; upiPreprocessOutput != nil {
			preprocesOutput, err := c.parseUpiPreprocessOutput(upiPreprocessOutput)
			if err != nil {
				return nil, nil, err
			}
			ops = append(ops, preprocesOutput)
		}
		if upiPostprocessOutput := output.UpiPostprocessOutput; upiPostprocessOutput != nil {
			postprocesOutput, err := c.parseUpiPostprocessOutput(upiPostprocessOutput)
			if err != nil {
				return nil, nil, err
			}
			ops = append(ops, postprocesOutput)
		}
	}

	return ops, preloadedTables, nil
}

func (c *Compiler) parseUpiPreprocessOutput(outputSpec *spec.UPIPreprocessOutput) (Op, error) {
	if outputSpec.PredictionTableName != "" {
		if err := c.checkVariableRegistered(outputSpec.PredictionTableName); err != nil {
			return nil, err
		}
	}

	for _, tableInput := range outputSpec.TransformerInputTableNames {
		if err := c.checkVariableRegistered(tableInput); err != nil {
			return nil, err
		}
	}
	output := NewUPIPreprocessOutputOp(outputSpec, c.operationTracingEnabled)
	return output, nil
}

func (c *Compiler) parseUpiPostprocessOutput(outputSpec *spec.UPIPostprocessOutput) (Op, error) {
	if err := c.checkVariableRegistered(outputSpec.PredictionResultTableName); err != nil {
		return nil, err
	}
	output := NewUPIPostprocessOutputOp(outputSpec, c.operationTracingEnabled)
	return output, nil
}

func (c *Compiler) parseUPIAutoloadSpec(autoloadSpec *spec.UPIAutoload, pipelineType types.Pipeline, compiledExpressions *expression.Storage) (Op, error) {
	for _, variableName := range autoloadSpec.VariableNames {
		c.registerDummyVariable(variableName)

	}
	for _, tableName := range autoloadSpec.TableNames {
		c.registerDummyTable(tableName)
	}
	return NewUPIAutoloadingOp(pipelineType, c.operationTracingEnabled), nil
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
				return nil, err
			}
			compiledExpressions.Set(v.Expression, compiledExpression)

		case *spec.Variable_JsonPath:
			compiledJsonPath, err := jsonpath.CompileWithOption(jsonpath.JsonPathOption{
				JsonPath: v.JsonPath,
				SrcType:  c.jsonpathSourceType,
			})
			if err != nil {
				return nil, err
			}
			compiledJsonPaths.Set(v.JsonPath, compiledJsonPath)
		case *spec.Variable_JsonPathConfig:
			jsonPathCfg := v.JsonPathConfig
			compiledJsonPath, err := jsonpath.CompileWithOption(jsonpath.JsonPathOption{
				JsonPath:     jsonPathCfg.JsonPath,
				DefaultValue: jsonPathCfg.DefaultValue,
				TargetType:   jsonPathCfg.ValueType,
				SrcType:      c.jsonpathSourceType,
			})
			if err != nil {
				return nil, err
			}
			compiledJsonPaths.Set(jsonPathCfg.JsonPath, compiledJsonPath)

		default:
			return nil, nil
		}
	}

	return NewVariableDeclarationOp(variables, c.operationTracingEnabled), nil
}

func (c *Compiler) parseFeastSpec(featureTableSpecs []*spec.FeatureTable, compiledJsonPaths *jsonpath.Storage, compiledExpressions *expression.Storage) (Op, error) {
	jsonPaths, err := feast.CompileJSONPaths(featureTableSpecs, c.jsonpathSourceType)
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

	entityExtractor := feast.NewEntityExtractor(compiledJsonPaths, compiledExpressions)
	return NewFeastOp(c.feastClients, c.feastOptions, entityExtractor, featureTableSpecs, c.logger, c.operationTracingEnabled), nil
}

func (c *Compiler) parseTablesSpec(tableSpecs []*spec.Table, compiledJsonPaths *jsonpath.Storage, compiledExpressions *expression.Storage) (*CreateTableOp, map[string]table.Table, error) {
	// for storing pre-loaded tables
	preloadedTables := map[string]table.Table{}

	for _, tableSpec := range tableSpecs {
		c.registerDummyTable(tableSpec.Name)
		if tableSpec.BaseTable != nil {
			switch bt := tableSpec.BaseTable.BaseTable.(type) {
			case *spec.BaseTable_FromJson:
				compiledJsonPath, err := jsonpath.CompileWithOption(jsonpath.JsonPathOption{
					JsonPath:     bt.FromJson.JsonPath,
					DefaultValue: bt.FromJson.DefaultValue,
					TargetType:   bt.FromJson.ValueType,
					SrcType:      c.jsonpathSourceType,
				})
				if err != nil {
					return nil, nil, err
				}
				compiledJsonPaths.Set(bt.FromJson.JsonPath, compiledJsonPath)
			case *spec.BaseTable_FromFile:
				var records [][]string
				var err error
				var colType map[string]gota.Type

				// parse filepath
				filePath, err := url.Parse(tableSpec.BaseTable.GetFromFile().GetUri())
				if err != nil {
					return nil, nil, err
				}

				// relative path in merlin
				if !filePath.IsAbs() && os.Getenv(envPredictorStorageURI) != "" {
					filePath, err = url.Parse(artifactsFolder + tableSpec.BaseTable.GetFromFile().GetUri())
					if err != nil {
						return nil, nil, err
					}
				}

				if tableSpec.BaseTable.GetFromFile().GetFormat() == spec.FromFile_CSV {
					records, err = table.RecordsFromCsv(filePath)
					colType = nil
				} else if tableSpec.BaseTable.GetFromFile().GetFormat() == spec.FromFile_PARQUET {
					records, colType, err = table.RecordsFromParquet(filePath)
				} else {
					return nil, nil, fmt.Errorf("unsupported/unspecified file type: %s", tableSpec.BaseTable.GetFromFile().GetFormat())
				}

				if err != nil {
					return nil, nil, fmt.Errorf("failed creating records from file %w", err)
				}

				loadedTable, err := table.NewFromRecords(records, colType, tableSpec.BaseTable.GetFromFile().GetSchema())
				if err != nil {
					return nil, nil, err
				}
				preloadedTables[tableSpec.Name] = *loadedTable
			}
		}

		if tableSpec.Columns != nil {
			for _, column := range tableSpec.Columns {
				switch cv := column.ColumnValue.(type) {
				case *spec.Column_Expression:
					compiledExpression, err := c.compileExpression(cv.Expression)
					if err != nil {
						return nil, nil, err
					}
					compiledExpressions.Set(cv.Expression, compiledExpression)
				case *spec.Column_FromJson:
					compiledJsonPath, err := jsonpath.CompileWithOption(jsonpath.JsonPathOption{
						JsonPath:     cv.FromJson.JsonPath,
						DefaultValue: cv.FromJson.DefaultValue,
						TargetType:   cv.FromJson.ValueType,
						SrcType:      c.jsonpathSourceType,
					})
					if err != nil {
						return nil, nil, err
					}
					compiledJsonPaths.Set(cv.FromJson.JsonPath, compiledJsonPath)
				}
			}
		}
	}

	return NewCreateTableOp(tableSpecs, c.operationTracingEnabled), preloadedTables, nil
}

func (c *Compiler) parseEncodersSpec(encoderSpecs []*spec.Encoder, compiledExpression *expression.Storage) (Op, error) {
	for _, encoderSpec := range encoderSpecs {
		c.registerDummyVariable(encoderSpec.Name)
	}
	return NewEncoderOp(encoderSpecs, c.operationTracingEnabled), nil
}

func (c *Compiler) parseTableTransform(transformationSpecs *spec.TableTransformation, paths *jsonpath.Storage, compiledExpressions *expression.Storage) (Op, error) {
	err := c.checkVariableRegistered(transformationSpecs.InputTable)
	if err != nil {
		return nil, err
	}

	for _, step := range transformationSpecs.Steps {
		for _, updateColumn := range step.UpdateColumns {
			if updateColumn.Expression != "" {
				compiledExpression, err := c.compileExpression(updateColumn.Expression)
				if err != nil {
					return nil, err
				}
				compiledExpressions.Set(updateColumn.Expression, compiledExpression)
				continue
			}
			for _, condition := range updateColumn.Conditions {
				if condition.Default != nil {
					compiledExpression, err := c.compileExpression(condition.Default.Expression)
					if err != nil {
						return nil, err
					}
					compiledExpressions.Set(condition.Default.Expression, compiledExpression)
					continue
				}
				compiledIfExpression, err := c.compileExpression(condition.RowSelector)
				if err != nil {
					return nil, err
				}
				compiledExpressions.Set(condition.RowSelector, compiledIfExpression)

				compiledExpression, err := c.compileExpression(condition.Expression)
				if err != nil {
					return nil, err
				}
				compiledExpressions.Set(condition.Expression, compiledExpression)
			}
		}

		for _, scaleCol := range step.ScaleColumns {
			if scaleCol.Column == "" {
				return nil, fmt.Errorf("scale column require non empty column")
			}

			scalerImpl, err := scaler.NewScaler(scaleCol)
			if err != nil {
				return nil, err
			}

			if err := scalerImpl.Validate(); err != nil {
				return nil, err
			}
		}

		for _, encodeColumn := range step.EncodeColumns {
			compiledExpression, err := c.compileExpression(encodeColumn.Encoder)
			if err != nil {
				return nil, err
			}
			compiledExpressions.Set(encodeColumn.Encoder, compiledExpression)
		}

		if step.FilterRow != nil {
			compiledExpression, err := c.compileExpression(step.FilterRow.Condition)
			if err != nil {
				return nil, err
			}
			compiledExpressions.Set(step.FilterRow.Condition, compiledExpression)
		}
	}

	c.registerDummyTable(transformationSpecs.OutputTable)
	return NewTableTransformOp(transformationSpecs, c.operationTracingEnabled), nil
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
	return NewTableJoinOp(tableJoinSpecs, c.operationTracingEnabled), nil
}

func (c *Compiler) parseJsonOutputSpec(jsonSpec *spec.JsonOutput, compiledJsonPaths *jsonpath.Storage, compiledExpressions *expression.Storage) (Op, error) {
	template := jsonSpec.JsonTemplate
	if template == nil {
		return nil, errors.New("jsontemplate must be specified")
	}
	if template.BaseJson != nil {
		compiledJsonPath, err := jsonpath.CompileWithOption(jsonpath.JsonPathOption{
			JsonPath: template.BaseJson.JsonPath,
			SrcType:  c.jsonpathSourceType,
		})
		if err != nil {
			return nil, err
		}
		compiledJsonPaths.Set(template.BaseJson.JsonPath, compiledJsonPath)
	}

	if err := c.parseJsonFields(template.Fields, compiledJsonPaths, compiledExpressions); err != nil {
		return nil, err
	}

	jsonOutputOp := NewJsonOutputOp(jsonSpec, c.operationTracingEnabled)

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
			compiledJsonPath, err := jsonpath.CompileWithOption(jsonpath.JsonPathOption{
				JsonPath: val.FromJson.JsonPath,
				SrcType:  c.jsonpathSourceType,
			})
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
	return expr.Compile(expression,
		expr.Env(c.sr),
		expr.Operator("&&", "AndOp"),
		expr.Operator("||", "OrOp"),
		expr.Operator(">", "GreaterOp"),
		expr.Operator(">=", "GreaterEqOp"),
		expr.Operator("<", "LessOp"),
		expr.Operator("<=", "LessEqOp"),
		expr.Operator("==", "EqualOp"),
		expr.Operator("!=", "NeqOp"),
		expr.Operator("+", "AddOp"),
		expr.Operator("-", "SubstractOp"),
		expr.Operator("*", "MultiplyOp"),
		expr.Operator("/", "DivideOp"),
		expr.Operator("%", "ModuloOp"),
	)
}

func (c *Compiler) registerDummyVariable(varName string) {
	c.sr[varName] = nil
}

func (c *Compiler) registerDummyTable(tableName string) {
	c.sr[tableName] = table.New()
}

func (c *Compiler) checkVariableRegistered(varName string) error {
	_, exist := c.sr[varName]
	if !exist {
		return fmt.Errorf("variable %s is not registered", varName)
	}

	return nil
}
