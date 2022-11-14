package pipeline

import (
	"context"
	"fmt"

	mErrors "github.com/gojek/merlin/pkg/errors"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/pkg/transformer/types/table"
	"github.com/opentracing/opentracing-go"
)

type CreateTableOp struct {
	tableSpecs []*spec.Table
	*OperationTracing
}

func NewCreateTableOp(tableSpecs []*spec.Table, tracingEnabled bool) *CreateTableOp {
	createTableOp := &CreateTableOp{
		tableSpecs: tableSpecs,
	}

	if tracingEnabled {
		createTableOp.OperationTracing = NewOperationTracing(tableSpecs, types.CreateTableOpType)
	}
	return createTableOp
}

func (c CreateTableOp) Execute(ctx context.Context, env *Environment) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "pipeline.CreateTableOp")
	defer span.Finish()

	for _, tableSpec := range c.tableSpecs {
		var t *table.Table
		if tableSpec.BaseTable != nil {
			switch tableSpec.BaseTable.BaseTable.(type) {
			case *spec.BaseTable_FromFile: // handle table creation from file separately
				continue
			default:
				tbl, err := createBaseTable(env, tableSpec.BaseTable)
				if err != nil {
					return fmt.Errorf("unable to create base table for %s: %w", tableSpec.Name, err)
				}
				t = tbl
			}
		}

		if tableSpec.Columns != nil {
			tbl, err := overrideColumns(env, t, tableSpec.Columns)
			if err != nil {
				return fmt.Errorf("unable to override column for table %s: %w", tableSpec.Name, err)
			}
			t = tbl
		}

		// register to environment
		env.SetSymbol(tableSpec.Name, t)
		if c.OperationTracing != nil {
			if err := c.AddInputOutput(nil, map[string]interface{}{tableSpec.Name: t}); err != nil {
				return err
			}
		}
		env.LogOperation("create_table", tableSpec.Name)
	}

	return nil
}

func createBaseTable(env *Environment, baseTableSpec *spec.BaseTable) (*table.Table, error) {
	switch baseTable := baseTableSpec.BaseTable.(type) {
	case *spec.BaseTable_FromJson:
		jsonObj, err := evalJSONPath(env, baseTable.FromJson.JsonPath)
		if err != nil {
			return nil, err
		}

		rawTable, err := toRawTable(jsonObj, baseTable.FromJson.AddRowNumber)
		if err != nil {
			return nil, fmt.Errorf("invalid json pointed by %s: %w", baseTable.FromJson.JsonPath, err)
		}

		return table.NewRaw(rawTable)
	case *spec.BaseTable_FromTable:
		s := env.SymbolRegistry()[baseTable.FromTable.TableName]
		if s == nil {
			return nil, fmt.Errorf("table %s is not found", baseTable.FromTable.TableName)
		}

		t, ok := s.(*table.Table)
		if !ok {
			return nil, fmt.Errorf("variable %s is not a table", baseTable.FromTable.TableName)
		}

		return t.Copy(), nil
	default:
		return nil, fmt.Errorf("unsupported base table spec %T", baseTable)
	}
}

func overrideColumns(env *Environment, t *table.Table, columns []*spec.Column) (*table.Table, error) {
	columnValues := make(map[string]interface{}, len(columns))
	for _, col := range columns {
		switch c := col.GetColumnValue().(type) {
		case *spec.Column_FromJson:
			result, err := evalJSONPath(env, c.FromJson.JsonPath)
			if err != nil {
				return nil, err
			}

			columnValues[col.Name] = result
		case *spec.Column_Expression:
			result, err := evalExpression(env, c.Expression)
			if err != nil {
				return nil, err
			}

			columnValues[col.Name] = result
		}
	}

	// create a new table
	if t == nil {
		return table.NewRaw(columnValues)
	}

	// update existing table
	err := t.UpdateColumnsRaw(columnValues)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func toRawTable(jsonObj interface{}, addRowNumber bool) (map[string]interface{}, error) {
	jsonArray, ok := jsonObj.([]interface{})
	if !ok {
		return nil, mErrors.NewInvalidInputErrorf("not an array")
	}

	nrow := len(jsonArray)
	rawTable := make(map[string]interface{})
	if addRowNumber {
		rawTable["row_number"] = make([]interface{}, nrow)
	}
	for idx, j := range jsonArray {
		node, ok := j.(map[string]interface{})
		if !ok {
			return nil, mErrors.NewInvalidInputError("not an array of JSON object")
		}
		for k, v := range node {
			ss, ok := rawTable[k]
			if !ok {
				ss = make([]interface{}, nrow)
			}
			ss.([]interface{})[idx] = v
			rawTable[k] = ss
		}

		if addRowNumber {
			rawTable["row_number"].([]interface{})[idx] = idx
		}
	}

	return rawTable, nil
}
