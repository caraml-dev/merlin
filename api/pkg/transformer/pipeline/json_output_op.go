package pipeline

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/pkg/transformer/types/series"
	"github.com/gojek/merlin/pkg/transformer/types/table"
)

type JsonOutputOp struct {
	outputSpec *spec.JsonOutput
}

func NewJsonOutputOp(outputSpec *spec.JsonOutput) Op {
	return &JsonOutputOp{outputSpec: outputSpec}
}

func (j JsonOutputOp) Execute(ctx context.Context, env *Environment) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "pipeline.JsonOutputOp")
	defer span.Finish()

	template := j.outputSpec.JsonTemplate
	outputJson := make(types.JSONObject)
	if template.BaseJson != nil {
		baseJsonOutput, err := j.createBaseJsonOutput(env, template.BaseJson)
		if err != nil {
			return err
		}
		outputJson = baseJsonOutput
	}

	for _, field := range template.Fields {
		var err error
		outputJson, err = generateJsonOutput(field, outputJson, env)
		if err != nil {
			return err
		}
	}

	env.SetOutputJSON(outputJson)
	return nil
}

func generateJsonOutput(field *spec.Field, output map[string]interface{}, env *Environment) (map[string]interface{}, error) {
	fieldName := field.FieldName
	switch val := field.Value.(type) {
	case *spec.Field_FromJson:
		jsonObj, err := evalJSONPath(env, val.FromJson.JsonPath)
		if err != nil {
			return nil, err
		}
		output[fieldName] = jsonObj
	case *spec.Field_FromTable:
		tbl, err := getTable(env, val.FromTable.TableName)
		if err != nil {
			return nil, err
		}

		tableJsonOutput, err := table.TableToJson(tbl, val.FromTable.Format)
		if err != nil {
			return nil, err
		}
		output[fieldName] = tableJsonOutput

	case *spec.Field_Expression:
		exprVal, err := evalExpression(env, val.Expression)
		if err != nil {
			return nil, err
		}

		var fieldVal interface{}
		switch valPerType := exprVal.(type) {
		case *series.Series:
			fieldVal = valPerType.GetRecords()
		case *table.Table:
			tableJsonOutput, err := table.TableToJson(valPerType, spec.FromTable_SPLIT)
			if err != nil {
				return nil, err
			}
			fieldVal = tableJsonOutput
		default:
			fieldVal = valPerType
		}

		output[fieldName] = fieldVal
	default:
		// check whether field already set from basejson
		jsonObj, ok := output[field.FieldName].(map[string]interface{})
		if ok {
			output[fieldName] = jsonObj
		} else {
			output[fieldName] = make(map[string]interface{})
		}

		for _, fieldVal := range field.Fields {
			var err error
			output[fieldName], err = generateJsonOutput(fieldVal, output[fieldName].(map[string]interface{}), env)
			if err != nil {
				return nil, err
			}
		}
	}
	return output, nil
}

func (j JsonOutputOp) createBaseJsonOutput(env *Environment, baseJson *spec.BaseJson) (types.JSONObject, error) {
	jsonObj, err := evalJSONPath(env, baseJson.JsonPath)
	if err != nil {
		return nil, err
	}

	jsonOutput, ok := jsonObj.(map[string]interface{})
	if !ok {
		return nil, errors.New("value in jsonpath must be object")
	}
	return types.JSONObject(jsonOutput), nil
}
